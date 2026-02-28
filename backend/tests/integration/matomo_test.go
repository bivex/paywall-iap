package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	matomoClient "github.com/bivex/paywall-iap/internal/infrastructure/external/matomo"
	"github.com/bivex/paywall-iap/internal/domain/service"
)

// MockMatomoHTTPServer mocks the Matomo HTTP API
type MockMatomoHTTPServer struct {
	server *httptest.Server
	events []map[string]string
}

func NewMockMatomoHTTPServer() *MockMatomoHTTPServer {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Store event for verification
		event := make(map[string]string)
		for key, values := range r.Form {
			if len(values) > 0 {
				event[key] = values[0]
			}
		}

		// Return success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))

	return &MockMatomoHTTPServer{
		server: server,
		events: make([]map[string]string, 0),
	}
}

func (m *MockMatomoHTTPServer) Close() {
	m.server.Close()
}

func (m *MockMatomoHTTPServer) URL() string {
	return m.server.URL
}

func (m *MockMatomoHTTPServer) GetEvents() []map[string]string {
	return m.events
}

func (m *MockMatomoHTTPServer) EventCount() int {
	return len(m.events)
}

// TestMatomoEventDelivery tests Matomo event delivery
func TestMatomoEventDelivery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	// Setup mock Matomo server
	matomoServer := NewMockMatomoHTTPServer()
	defer matomoServer.Close()

	config := matomoClient.Config{
		BaseURL:    matomoServer.URL(),
		SiteID:     "1",
		TokenAuth:  "test_token",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	matomo := matomoClient.NewClient(config, logger)

	t.Run("EnqueueEvent", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)
		forwarder := service.NewMatomoForwarder(matomo, repo, logger)

		userID := uuid.New()

		// Enqueue a standard event
		err := forwarder.TrackEvent(ctx, &userID, "paywall", "shown", "premium_monthly", 0, map[string]string{
			"experiment_id": "exp_123",
			"variant":      "control",
		})
		assert.NoError(t, err)

		// Verify event was enqueued
		events, err := repo.GetPendingEvents(ctx, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
	})

	t.Run("BatchProcessing", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)
		forwarder := service.NewMatomoForwarder(matomo, repo, logger)

		// Enqueue multiple events
		userID := uuid.New()
		for i := 0; i < 5; i++ {
			err := forwarder.TrackEvent(ctx, &userID, "test", "action", fmt.Sprintf("event_%d", i), 0, nil)
			require.NoError(t, err)
		}

		// Process batch
		processed, succeeded, failed, err := forwarder.ProcessBatch(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 5, processed)
		assert.Equal(t, 5, succeeded) // All should succeed with mock server
		assert.Equal(t, 0, failed)
	})

	t.Run("RetryLogic", func(t *testing.T) {
		// Create a server that fails initially then succeeds
		attemptCount := 0
		flakyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer flakyServer.Close()

		config := matomoClient.Config{
			BaseURL:    flakyServer.URL,
			SiteID:     "1",
			TokenAuth:  "test_token",
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		}

		flakyMatomo := matomoClient.NewClient(config, logger)

		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)
		forwarder := service.NewMatomoForwarder(flakyMatomo, repo, logger)

		userID := uuid.New()
		err := forwarder.TrackEvent(ctx, &userID, "test", "retry", "", 0, nil)
		require.NoError(t, err)

		// First attempt should fail, event should be scheduled for retry
		processed, succeeded, failed, err := forwarder.ProcessBatch(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
		assert.Equal(t, 0, succeeded) // Failed on first try

		// Check that event is scheduled for retry
		events, err := repo.GetPendingEvents(ctx, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		if len(events) > 0 {
			event := events[0]
			assert.True(t, event.NextRetryAt.After(time.Now()))
			assert.Equal(t, 1, event.RetryCount)
		}
	})

	t.Run("FallbackStorage", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)

		// Create event that will fail permanently
		failedEvent := &service.MatomoStagedEvent{
			ID:           uuid.New(),
			EventType:    "event",
			UserID:       nil,
			Payload:      map[string]interface{}{"test": "data"},
			RetryCount:   3,
			MaxRetries:   3,
			NextRetryAt:  time.Now(),
			Status:       "failed",
			CreatedAt:    time.Now(),
			FailedAt:     timePtr(time.Now()),
			ErrorMessage: strPtr("Connection refused"),
		}

		err := repo.EnqueueEvent(ctx, failedEvent)
		assert.NoError(t, err)

		// Update status to failed
		err = repo.UpdateEventStatus(ctx, failedEvent.ID, "failed", fmt.Errorf("test error"))
		assert.NoError(t, err)

		// Retrieve failed events
		failedEvents, err := repo.GetFailedEvents(ctx, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(failedEvents), 1)
	})

	t.Run("EcommerceEvent", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)
		forwarder := service.NewMatomoForwarder(matomo, repo, logger)

		userID := uuid.New()

		// Create ecommerce event with items
		items := []matomoClient.EcommerceItem{
			{
				SKU:      "premium_monthly",
				Name:     "Premium Monthly",
				Price:    9.99,
				Quantity: 1,
				Category: "subscription",
			},
		}

		err := forwarder.TrackPurchase(ctx, &userID, "order_123", 9.99, items, map[string]string{
			"experiment_id": "exp_456",
			"variant":      "variant_a",
		})
		assert.NoError(t, err)

		// Verify event was enqueued with correct payload
		events, err := repo.GetPendingEvents(ctx, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)

		// Find the ecommerce event
		var ecommerceEvent *service.MatomoStagedEvent
		for _, event := range events {
			if event.EventType == "ecommerce" {
				ecommerceEvent = event
				break
			}
		}

		require.NotNil(t, ecommerceEvent)
		assert.Equal(t, "ecommerce", ecommerceEvent.EventType)

		// Verify payload structure
		assert.Contains(t, ecommerceEvent.Payload, "revenue")
		assert.Contains(t, ecommerceEvent.Payload, "items")
	})

	t.Run("CleanupOldEvents", func(t *testing.T) {
		// Setup test database
		db := testutil.SetupTestDB(t)
		defer testutil.TeardownTestDB(t, db)

		repo := service.NewPostgresMatomoEventRepository(db, logger)

		// Create old sent events
		oldTime := time.Now().Add(-31 * 24 * time.Hour)

		// Manually insert old events (bypassing repository for timestamp control)
		_, err := db.Exec(ctx, `
			INSERT INTO matomo_staged_events (id, event_type, user_id, payload, status, sent_at, created_at)
			VALUES ($1, $2, NULL, $3, $4, $5, $5)
		`,
			uuid.New(),
			"event",
			map[string]interface{}{"test": "old"},
			"sent",
			oldTime,
		)
		require.NoError(t, err)

		// Cleanup events older than 30 days
		count, err := repo.CleanupOldSentEvents(ctx, 30*24*time.Hour)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})
}

// TestMatomoAPIClient tests the Matomo HTTP client directly
func TestMatomoAPIClient(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("TrackEvent", func(t *testing.T) {
		matomoServer := NewMockMatomoHTTPServer()
		defer matomoServer.Close()

		config := matomoClient.Config{
			BaseURL:    matomoServer.URL(),
			SiteID:     "1",
			TokenAuth:  "test_token",
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		}

		matomo := matomoClient.NewClient(config, logger)

		req := matomoClient.TrackEventRequest{
			Category: "paywall",
			Action:   "shown",
			Name:     "premium",
			Value:    0,
			UserID:   "user_123",
			CustomVariables: map[string]string{
				"experiment_id": "exp_123",
				"variant":      "control",
			},
		}

		err := matomo.TrackEvent(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("TrackEcommerce", func(t *testing.T) {
		matomoServer := NewMockMatomoHTTPServer()
		defer matomoServer.Close()

		config := matomoClient.Config{
			BaseURL:    matomoServer.URL(),
			SiteID:     "1",
			TokenAuth:  "test_token",
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		}

		matomo := matomoClient.NewClient(config, logger)

		items := []matomoClient.EcommerceItem{
			{
				SKU:      "com.app.premium.monthly",
				Name:     "Premium Monthly",
				Price:    9.99,
				Quantity: 1,
			},
		}

		req := matomoClient.TrackEcommerceRequest{
			UserID:  "user_456",
			Revenue: 9.99,
			OrderID: "order_789",
			Items:   items,
		}

		err := matomo.TrackEcommerce(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		matomoServer := NewMockMatomoHTTPServer()
		defer matomoServer.Close()

		config := matomoClient.Config{
			BaseURL:    matomoServer.URL(),
			SiteID:     "1",
			TokenAuth:  "test_token",
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		}

		matomo := matomoClient.NewClient(config, logger)

		err := matomo.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("RetryOnFailure", func(t *testing.T) {
		attemptCount := 0
		flakyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer flakyServer.Close()

		config := matomoClient.Config{
			BaseURL:    flakyServer.URL,
			SiteID:     "1",
			TokenAuth:  "test_token",
			Timeout:    1 * time.Second,
			MaxRetries: 3,
		}

		matomo := matomoClient.NewClient(config, logger)

		req := matomoClient.TrackEventRequest{
			Category: "test",
			Action:   "retry",
			UserID:   "user_123",
		}

		err := matomo.TrackEvent(ctx, req)
		assert.NoError(t, err) // Should succeed after retries
		assert.Equal(t, 3, attemptCount) // Should have retried 3 times
	})
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func strPtr(s string) *string {
	return &s
}
