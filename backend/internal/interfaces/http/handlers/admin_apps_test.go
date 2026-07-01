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
