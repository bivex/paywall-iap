package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

type banditServiceStub struct {
	trackImpressionFunc       func(ctx context.Context, experimentID, armID, userID uuid.UUID, event *service.ImpressionEvent) error
	updateRewardFunc          func(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error
	updateRewardWithEventFunc func(ctx context.Context, experimentID, armID uuid.UUID, reward float64, event *service.ConversionEvent) error
}

func (s banditServiceStub) SelectArm(ctx context.Context, experimentID, userID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (s banditServiceStub) TrackImpression(ctx context.Context, experimentID, armID, userID uuid.UUID, event *service.ImpressionEvent) error {
	if s.trackImpressionFunc != nil {
		return s.trackImpressionFunc(ctx, experimentID, armID, userID, event)
	}
	return nil
}

func (s banditServiceStub) UpdateReward(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
	if s.updateRewardFunc != nil {
		return s.updateRewardFunc(ctx, experimentID, armID, reward)
	}
	return nil
}

func (s banditServiceStub) UpdateRewardWithEvent(ctx context.Context, experimentID, armID uuid.UUID, reward float64, event *service.ConversionEvent) error {
	if s.updateRewardWithEventFunc != nil {
		return s.updateRewardWithEventFunc(ctx, experimentID, armID, reward, event)
	}
	if s.updateRewardFunc != nil {
		return s.updateRewardFunc(ctx, experimentID, armID, reward)
	}
	return nil
}

func (s banditServiceStub) GetArmStatistics(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*service.ArmStats, error) {
	return nil, nil
}

func (s banditServiceStub) CalculateWinProbability(ctx context.Context, experimentID uuid.UUID, simulations int) (map[uuid.UUID]float64, error) {
	return nil, nil
}

func TestReward_AcceptsZeroReward(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var recorded float64
	handler := NewBanditHandler(banditServiceStub{
		updateRewardFunc: func(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
			recorded = reward
			return nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/bandit/reward", strings.NewReader(`{"experiment_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","arm_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","user_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","reward":0}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Reward(ctx)

	require.Equal(t, http.StatusOK, recorder.Code, "body=%s", recorder.Body.String())
	require.Zero(t, recorded)
	require.Contains(t, recorder.Body.String(), `"updated":true`)
	require.Contains(t, recorder.Body.String(), `"reward":0`)
}

func TestImpression_AcceptsMetadata(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var recorded *service.ImpressionEvent
	handler := NewBanditHandler(banditServiceStub{
		trackImpressionFunc: func(ctx context.Context, experimentID, armID, userID uuid.UUID, event *service.ImpressionEvent) error {
			recorded = event
			return nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/bandit/impression", strings.NewReader(`{"experiment_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","arm_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","user_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","metadata":{"placement":"paywall"}}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Impression(ctx)

	require.Equal(t, http.StatusOK, recorder.Code, "body=%s", recorder.Body.String())
	require.NotNil(t, recorded)
	require.Equal(t, service.ImpressionEventTypeImpression, recorded.EventType)
	require.Equal(t, "paywall", recorded.Metadata["placement"])
	require.Contains(t, recorder.Body.String(), `"tracked":true`)
}

func TestImpression_ReturnsNotFoundForMissingArm(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewBanditHandler(banditServiceStub{
		trackImpressionFunc: func(ctx context.Context, experimentID, armID, userID uuid.UUID, event *service.ImpressionEvent) error {
			return service.ErrBanditArmNotFound
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/bandit/impression", strings.NewReader(`{"experiment_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","arm_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","user_id":"e3e70682-c209-4cac-629f-6fbed82c07cd"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Impression(ctx)

	require.Equal(t, http.StatusNotFound, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Arm not found"`)
}

func TestReward_ReturnsNotFoundForMissingArm(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewBanditHandler(banditServiceStub{
		updateRewardFunc: func(ctx context.Context, experimentID, armID uuid.UUID, reward float64) error {
			return service.ErrBanditArmNotFound
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/bandit/reward", strings.NewReader(`{"experiment_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","arm_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","user_id":"e3e70682-c209-4cac-629f-6fbed82c07cd","reward":0}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Reward(ctx)

	require.Equal(t, http.StatusNotFound, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Arm not found"`)
}

func TestStatistics_RejectsEmptyWinProbs(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewBanditHandler(banditServiceStub{})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/bandit/statistics?experiment_id=e3e70682-c209-4cac-629f-6fbed82c07cd&win_probs=", nil)

	handler.Statistics(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid win_probs value"`)
}

func TestStatistics_RejectsNumericWinProbs(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewBanditHandler(banditServiceStub{})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/bandit/statistics?experiment_id=e3e70682-c209-4cac-629f-6fbed82c07cd&win_probs=0", nil)

	handler.Statistics(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid win_probs value"`)
}

func TestStatistics_RejectsUnknownQueryParameter(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewBanditHandler(banditServiceStub{})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/bandit/statistics?experiment_id=e3e70682-c209-4cac-629f-6fbed82c07cd&extra=1", nil)

	handler.Statistics(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Unknown query parameter: extra"`)
}
