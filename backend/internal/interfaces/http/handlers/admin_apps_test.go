package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/interfaces/http/handlers"
)

// mockAppRepository satisfies domain/repository.AppRepository
type mockAppRepository struct {
	mock.Mock
}

func (m *mockAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.App), args.Error(1)
}

func (m *mockAppRepository) GetByBundleID(ctx context.Context, bundleID string) (*entity.App, error) {
	args := m.Called(ctx, bundleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.App), args.Error(1)
}

func (m *mockAppRepository) List(ctx context.Context) ([]*entity.App, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.App), args.Error(1)
}

func (m *mockAppRepository) Create(ctx context.Context, name, bundleID, platform string) (*entity.App, error) {
	args := m.Called(ctx, name, bundleID, platform)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.App), args.Error(1)
}

func (m *mockAppRepository) Update(ctx context.Context, app *entity.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *mockAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// helpers

func newRouter(h *handlers.AppsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/v1/admin/apps")
	g.GET("", h.ListApps)
	g.POST("", h.CreateApp)
	g.GET("/:id", h.GetApp)
	g.PUT("/:id", h.UpdateApp)
	g.DELETE("/:id", h.DeleteApp)
	return r
}

func sampleApp() *entity.App {
	return &entity.App{
		ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:        "com.mothsalt.game1",
		DisplayName: "Game One",
		Platform:    "ios",
		BundleID:    "com.mothsalt.game1",
		IsActive:    true,
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ----- ListApps -----

func TestListApps_OK(t *testing.T) {
	repo := new(mockAppRepository)
	apps := []*entity.App{sampleApp()}
	repo.On("List", mock.Anything).Return(apps, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps", nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	list := body["apps"].([]interface{})
	assert.Len(t, list, 1)
	assert.Equal(t, "com.mothsalt.game1", list[0].(map[string]interface{})["name"])
	repo.AssertExpectations(t)
}

func TestListApps_Empty(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("List", mock.Anything).Return([]*entity.App{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps", nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Empty(t, body["apps"])
	repo.AssertExpectations(t)
}

func TestListApps_RepoError(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("List", mock.Anything).Return(nil, fmt.Errorf("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps", nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

// ----- GetApp -----

func TestGetApp_OK(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps/"+app.ID.String(), nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var dto map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dto))
	assert.Equal(t, app.ID.String(), dto["id"])
	repo.AssertExpectations(t)
}

func TestGetApp_InvalidID(t *testing.T) {
	repo := new(mockAppRepository)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps/not-a-uuid", nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "GetByID")
}

func TestGetApp_NotFound(t *testing.T) {
	repo := new(mockAppRepository)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, fmt.Errorf("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps/"+id.String(), nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

// ----- CreateApp -----

func TestCreateApp_OK(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("Create", mock.Anything, "com.mothsalt.game1", "com.mothsalt.game1", "ios").Return(app, nil)

	body, _ := json.Marshal(map[string]string{
		"name":      "com.mothsalt.game1",
		"bundle_id": "com.mothsalt.game1",
		"platform":  "ios",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var dto map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dto))
	assert.Equal(t, "com.mothsalt.game1", dto["name"])
	repo.AssertExpectations(t)
}

func TestCreateApp_WithDisplayName(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("Create", mock.Anything, "com.mothsalt.game1", "com.mothsalt.game1", "ios").Return(app, nil)
	// display_name triggers an Update call
	repo.On("Update", mock.Anything, mock.MatchedBy(func(a *entity.App) bool {
		return a.DisplayName == "My Game"
	})).Return(nil)

	body, _ := json.Marshal(map[string]string{
		"name":         "com.mothsalt.game1",
		"bundle_id":    "com.mothsalt.game1",
		"platform":     "ios",
		"display_name": "My Game",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

func TestCreateApp_MissingFields(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{"name": "only-name"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

func TestCreateApp_InvalidPlatform(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{
		"name":      "com.mothsalt.game1",
		"bundle_id": "com.mothsalt.game1",
		"platform":  "windows", // not in oneof
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

func TestCreateApp_RepoError(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("db error"))

	body, _ := json.Marshal(map[string]string{
		"name":      "com.mothsalt.game1",
		"bundle_id": "com.mothsalt.game1",
		"platform":  "android",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

// ----- UpdateApp -----

func TestUpdateApp_OK(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(a *entity.App) bool {
		return a.Name == "com.mothsalt.game2" && a.Platform == "android"
	})).Return(nil)

	newName := "com.mothsalt.game2"
	newPlatform := "android"
	body, _ := json.Marshal(map[string]interface{}{
		"name":     newName,
		"platform": newPlatform,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var dto map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dto))
	assert.Equal(t, newName, dto["name"])
	assert.Equal(t, newPlatform, dto["platform"])
	repo.AssertExpectations(t)
}

func TestUpdateApp_Deactivate(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(a *entity.App) bool {
		return !a.IsActive
	})).Return(nil)

	isActive := false
	body, _ := json.Marshal(map[string]interface{}{"is_active": isActive})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestUpdateApp_InvalidID(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{"name": "x"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/bad-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "GetByID")
}

func TestUpdateApp_NotFound(t *testing.T) {
	repo := new(mockAppRepository)
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, fmt.Errorf("not found"))

	body, _ := json.Marshal(map[string]string{"name": "x"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestUpdateApp_RepoUpdateError(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(fmt.Errorf("db error"))

	body, _ := json.Marshal(map[string]string{"name": "new-name"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

// ----- CreateApp edge cases -----

// TestCreateApp_EmptyBody sends {} — all required fields missing → 400, Create not called.
func TestCreateApp_EmptyBody(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

// TestCreateApp_EmptyBundleID sends bundle_id: "" — fails min=1 → 400, Create not called.
func TestCreateApp_EmptyBundleID(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{
		"name":      "com.example.app",
		"bundle_id": "",
		"platform":  "ios",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

// TestCreateApp_WhitespaceName sends name: "   " — length 3 passes min=1/max=128, so binding
// succeeds and Create IS called. We verify the handler reaches the repo layer.
func TestCreateApp_WhitespaceName(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("Create", mock.Anything, "   ", "com.example.app", "ios").
		Return(nil, fmt.Errorf("db error"))

	body, _ := json.Marshal(map[string]string{
		"name":      "   ",
		"bundle_id": "com.example.app",
		"platform":  "ios",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	// Binding passes (length > 0); handler forwards to repo → repo error → 500.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

// TestCreateApp_BothPlatforms sends platform: "both" — "both" IS in the oneof list,
// so this is a valid request and should return 201 Created.
func TestCreateApp_BothPlatforms(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	app.Platform = "both"
	repo.On("Create", mock.Anything, "com.example.app", "com.example.app", "both").Return(app, nil)

	body, _ := json.Marshal(map[string]string{
		"name":      "com.example.app",
		"bundle_id": "com.example.app",
		"platform":  "both",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	// "both" is a valid platform per the binding tag — expect 201.
	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

// TestCreateApp_UnknownPlatform sends platform: "xbox" — not in oneof → 400, Create not called.
func TestCreateApp_UnknownPlatform(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{
		"name":      "com.example.app",
		"bundle_id": "com.example.app",
		"platform":  "xbox",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

// TestCreateApp_VeryLongName sends a 300-char name — exceeds max=128 → 400, Create not called.
func TestCreateApp_VeryLongName(t *testing.T) {
	repo := new(mockAppRepository)

	longName := string(make([]byte, 300))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}
	// Build a 300-char string simply.
	longName = fmt.Sprintf("%0300d", 0) // 300 chars of '0'

	body, _ := json.Marshal(map[string]string{
		"name":      longName,
		"bundle_id": "com.example.app",
		"platform":  "ios",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	// max=128 enforced by binding → 400.
	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create")
}

// TestCreateApp_ConflictError — repo.Create returns a "duplicate" error. The handler has no
// special conflict handling, so it returns 500 InternalServerError.
func TestCreateApp_ConflictError(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("duplicate key value violates unique constraint"))

	body, _ := json.Marshal(map[string]string{
		"name":      "com.example.app",
		"bundle_id": "com.example.app",
		"platform":  "ios",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/apps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	// No conflict-specific handling in the handler → generic 500.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

// ----- UpdateApp edge cases -----

// TestUpdateApp_EmptyBody sends {} with a valid ID. All update fields are optional pointers,
// so binding succeeds, GetByID is called, Update is called with no changes → 200.
func TestUpdateApp_EmptyBody(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)
	repo.On("Update", mock.Anything, app).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

// TestUpdateApp_DeactivateAndRename sends both is_active: false and a new name in one request.
// Expects 200 and repo.Update called once with both changes applied.
func TestUpdateApp_DeactivateAndRename(t *testing.T) {
	repo := new(mockAppRepository)
	app := sampleApp()
	repo.On("GetByID", mock.Anything, app.ID).Return(app, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(a *entity.App) bool {
		return a.Name == "renamed-app" && !a.IsActive
	})).Return(nil)

	newName := "renamed-app"
	isActive := false
	body, _ := json.Marshal(map[string]interface{}{
		"name":      newName,
		"is_active": isActive,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+app.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var dto map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dto))
	assert.Equal(t, newName, dto["name"])
	assert.Equal(t, false, dto["is_active"])
	repo.AssertExpectations(t)
	// Update was called exactly once (not twice).
	repo.AssertNumberOfCalls(t, "Update", 1)
}

// TestUpdateApp_InvalidPlatform sends platform: "xbox" — fails omitempty,oneof binding → 400,
// GetByID is never called.
func TestUpdateApp_InvalidPlatform(t *testing.T) {
	repo := new(mockAppRepository)

	body, _ := json.Marshal(map[string]string{"platform": "xbox"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/admin/apps/"+sampleApp().ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "GetByID")
}

// ----- Middleware / routing -----

// TestAppsHandler_RequiresNoAppID verifies that GET /v1/admin/apps works without
// an X-App-ID header — apps are global, not per-app scoped.
func TestAppsHandler_RequiresNoAppID(t *testing.T) {
	repo := new(mockAppRepository)
	repo.On("List", mock.Anything).Return([]*entity.App{sampleApp()}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/apps", nil)
	// Deliberately omit X-App-ID header.
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

// ----- DeleteApp -----

func TestDeleteApp_OK(t *testing.T) {
	repo := new(mockAppRepository)
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	repo.On("Delete", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/apps/"+id.String(), nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestDeleteApp_InvalidID(t *testing.T) {
	repo := new(mockAppRepository)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/apps/not-a-uuid", nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Delete")
}

func TestDeleteApp_RepoError(t *testing.T) {
	repo := new(mockAppRepository)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(fmt.Errorf("db error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/admin/apps/"+id.String(), nil)
	newRouter(handlers.NewAppsHandler(repo)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}
