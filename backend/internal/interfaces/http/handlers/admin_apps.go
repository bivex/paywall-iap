package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AppsHandler handles /v1/admin/apps endpoints
type AppsHandler struct {
	appRepo domainRepo.AppRepository
}

// NewAppsHandler creates a new AppsHandler
func NewAppsHandler(appRepo domainRepo.AppRepository) *AppsHandler {
	return &AppsHandler{appRepo: appRepo}
}

type appDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	BundleID    string    `json:"bundle_id"`
	Platform    string    `json:"platform"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type listAppsResponse struct {
	Apps []appDTO `json:"apps"`
}

type createAppRequest struct {
	Name        string `json:"name"         binding:"required,min=1,max=128"`
	DisplayName string `json:"display_name"`
	BundleID    string `json:"bundle_id"    binding:"required,min=1,max=256"`
	Platform    string `json:"platform"     binding:"required,oneof=ios android both"`
}

type updateAppRequest struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
	BundleID    *string `json:"bundle_id"`
	Platform    *string `json:"platform" binding:"omitempty,oneof=ios android both"`
	IsActive    *bool   `json:"is_active"`
}

// ListApps GET /v1/admin/apps
func (h *AppsHandler) ListApps(c *gin.Context) {
	apps, err := h.appRepo.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list apps")
		return
	}
	dtos := make([]appDTO, 0, len(apps))
	for _, a := range apps {
		dtos = append(dtos, appDTO{
			ID:          a.ID,
			Name:        a.Name,
			DisplayName: a.DisplayName,
			BundleID:    a.BundleID,
			Platform:    a.Platform,
			IsActive:    a.IsActive,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listAppsResponse{Apps: dtos})
}

// GetApp GET /v1/admin/apps/:id
func (h *AppsHandler) GetApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	app, err := h.appRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "app not found")
		return
	}
	c.JSON(http.StatusOK, appDTO{
		ID:          app.ID,
		Name:        app.Name,
		DisplayName: app.DisplayName,
		BundleID:    app.BundleID,
		Platform:    app.Platform,
		IsActive:    app.IsActive,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	})
}

// CreateApp POST /v1/admin/apps
func (h *AppsHandler) CreateApp(c *gin.Context) {
	var req createAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	app, err := h.appRepo.Create(c.Request.Context(), req.Name, req.BundleID, req.Platform)
	if err != nil {
		response.InternalError(c, "failed to create app")
		return
	}
	// Use display_name from request if provided
	if req.DisplayName != "" {
		app.DisplayName = req.DisplayName
		_ = h.appRepo.Update(c.Request.Context(), app)
	}
	c.JSON(http.StatusCreated, appDTO{
		ID:          app.ID,
		Name:        app.Name,
		DisplayName: app.DisplayName,
		BundleID:    app.BundleID,
		Platform:    app.Platform,
		IsActive:    app.IsActive,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	})
}

// UpdateApp PUT /v1/admin/apps/:id
func (h *AppsHandler) UpdateApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	var req updateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	app, err := h.appRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "app not found")
		return
	}
	if req.Name != nil {
		app.Name = *req.Name
	}
	if req.DisplayName != nil {
		app.DisplayName = *req.DisplayName
	}
	if req.BundleID != nil {
		app.BundleID = *req.BundleID
	}
	if req.Platform != nil {
		app.Platform = *req.Platform
	}
	if req.IsActive != nil {
		app.IsActive = *req.IsActive
	}
	if err := h.appRepo.Update(c.Request.Context(), app); err != nil {
		response.InternalError(c, "failed to update app")
		return
	}
	c.JSON(http.StatusOK, appDTO{
		ID:          app.ID,
		Name:        app.Name,
		DisplayName: app.DisplayName,
		BundleID:    app.BundleID,
		Platform:    app.Platform,
		IsActive:    app.IsActive,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	})
}

// DeleteApp DELETE /v1/admin/apps/:id
func (h *AppsHandler) DeleteApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	if err := h.appRepo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete app")
		return
	}
	response.NoContent(c)
}
