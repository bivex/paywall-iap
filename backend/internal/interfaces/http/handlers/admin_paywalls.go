package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	httpmiddleware "github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AppPaywall represents a saved paywall configuration for an app.
type AppPaywall struct {
	ID          string          `json:"id"`
	AppID       string          `json:"app_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type paywallUpsertRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"`
	IsActive    bool            `json:"is_active"`
}

// AdminPaywallsHandler handles CRUD for per-app paywall configurations.
type AdminPaywallsHandler struct {
	pool *pgxpool.Pool
}

func NewAdminPaywallsHandler(pool *pgxpool.Pool) *AdminPaywallsHandler {
	return &AdminPaywallsHandler{pool: pool}
}

func scanPaywall(row pgx.Row) (AppPaywall, error) {
	var p AppPaywall
	var defRaw []byte
	err := row.Scan(&p.ID, &p.AppID, &p.Name, &p.Description, &defRaw, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	p.Definition = json.RawMessage(defRaw)
	return p, nil
}

// ListPaywalls GET /v1/admin/paywalls
func (h *AdminPaywallsHandler) ListPaywalls(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, app_id, name, description, definition, is_active, created_at, updated_at
		FROM app_paywalls
		WHERE app_id = $1
		ORDER BY is_active DESC, updated_at DESC
	`, appID)
	if err != nil {
		response.InternalError(c, "Failed to list paywalls")
		return
	}
	defer rows.Close()

	paywalls := make([]AppPaywall, 0)
	for rows.Next() {
		var p AppPaywall
		var defRaw []byte
		if err := rows.Scan(&p.ID, &p.AppID, &p.Name, &p.Description, &defRaw, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			response.InternalError(c, "Failed to scan paywall row")
			return
		}
		p.Definition = json.RawMessage(defRaw)
		paywalls = append(paywalls, p)
	}

	response.OK(c, gin.H{"paywalls": paywalls, "total": len(paywalls)})
}

// GetPaywall GET /v1/admin/paywalls/:id
func (h *AdminPaywallsHandler) GetPaywall(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)
	id := c.Param("id")

	p, err := scanPaywall(h.pool.QueryRow(c.Request.Context(), `
		SELECT id, app_id, name, description, definition, is_active, created_at, updated_at
		FROM app_paywalls
		WHERE id = $1 AND app_id = $2
	`, id, appID))
	if err == pgx.ErrNoRows {
		response.NotFound(c, "Paywall not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to get paywall")
		return
	}
	response.OK(c, p)
}

// CreatePaywall POST /v1/admin/paywalls
func (h *AdminPaywallsHandler) CreatePaywall(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)

	var req paywallUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	if req.Definition == nil {
		req.Definition = json.RawMessage("{}")
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	if req.IsActive {
		if _, err := tx.Exec(c.Request.Context(),
			`UPDATE app_paywalls SET is_active = false, updated_at = now() WHERE app_id = $1`, appID); err != nil {
			response.InternalError(c, "Failed to deactivate existing paywalls")
			return
		}
	}

	p, err := scanPaywall(tx.QueryRow(c.Request.Context(), `
		INSERT INTO app_paywalls (app_id, name, description, definition, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, app_id, name, description, definition, is_active, created_at, updated_at
	`, appID, req.Name, req.Description, []byte(req.Definition), req.IsActive))
	if err != nil {
		response.InternalError(c, "Failed to create paywall")
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		response.InternalError(c, "Failed to commit")
		return
	}

	response.Created(c, p)
}

// UpdatePaywall PUT /v1/admin/paywalls/:id
func (h *AdminPaywallsHandler) UpdatePaywall(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)
	id := c.Param("id")

	var req paywallUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	if req.Definition == nil {
		req.Definition = json.RawMessage("{}")
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	if req.IsActive {
		if _, err := tx.Exec(c.Request.Context(),
			`UPDATE app_paywalls SET is_active = false, updated_at = now() WHERE app_id = $1 AND id != $2`, appID, id); err != nil {
			response.InternalError(c, "Failed to deactivate existing paywalls")
			return
		}
	}

	p, err := scanPaywall(tx.QueryRow(c.Request.Context(), `
		UPDATE app_paywalls
		SET name=$1, description=$2, definition=$3, is_active=$4, updated_at=now()
		WHERE id=$5 AND app_id=$6
		RETURNING id, app_id, name, description, definition, is_active, created_at, updated_at
	`, req.Name, req.Description, []byte(req.Definition), req.IsActive, id, appID))
	if err == pgx.ErrNoRows {
		response.NotFound(c, "Paywall not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to update paywall")
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		response.InternalError(c, "Failed to commit")
		return
	}

	response.OK(c, p)
}

// ActivatePaywall POST /v1/admin/paywalls/:id/activate
func (h *AdminPaywallsHandler) ActivatePaywall(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)
	id := c.Param("id")

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(c.Request.Context()) //nolint:errcheck

	if _, err := tx.Exec(c.Request.Context(),
		`UPDATE app_paywalls SET is_active = false, updated_at = now() WHERE app_id = $1`, appID); err != nil {
		response.InternalError(c, "Failed to deactivate paywalls")
		return
	}

	p, err := scanPaywall(tx.QueryRow(c.Request.Context(), `
		UPDATE app_paywalls SET is_active = true, updated_at = now()
		WHERE id = $1 AND app_id = $2
		RETURNING id, app_id, name, description, definition, is_active, created_at, updated_at
	`, id, appID))
	if err == pgx.ErrNoRows {
		response.NotFound(c, "Paywall not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to activate paywall")
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		response.InternalError(c, "Failed to commit")
		return
	}

	response.OK(c, p)
}

// DeletePaywall DELETE /v1/admin/paywalls/:id
func (h *AdminPaywallsHandler) DeletePaywall(c *gin.Context) {
	appID := httpmiddleware.GetAppID(c)
	id := c.Param("id")

	tag, err := h.pool.Exec(c.Request.Context(),
		`DELETE FROM app_paywalls WHERE id = $1 AND app_id = $2`, id, appID)
	if err != nil {
		response.InternalError(c, "Failed to delete paywall")
		return
	}
	if tag.RowsAffected() == 0 {
		response.NotFound(c, "Paywall not found")
		return
	}

	response.OK(c, gin.H{"deleted": true})
}
