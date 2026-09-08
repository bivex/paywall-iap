package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	registerCmd   *command.RegisterCommand
	adminLoginCmd *command.AdminLoginCommand
	jwtMiddleware *middleware.JWTMiddleware
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	registerCmd *command.RegisterCommand,
	adminLoginCmd *command.AdminLoginCommand,
	jwtMiddleware *middleware.JWTMiddleware,
) *AuthHandler {
	return &AuthHandler{
		registerCmd:   registerCmd,
		adminLoginCmd: adminLoginCmd,
		jwtMiddleware: jwtMiddleware,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration request"
// @Success 201 {object} response.SuccessResponse{data=dto.RegisterResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := bindStrictJSON(c, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.AppID == "" {
		if rawAppID := c.GetHeader("X-App-ID"); rawAppID != "" {
			req.AppID = rawAppID
		} else {
			req.AppID = "00000000-0000-0000-0000-000000000001"
		}
	}

	resp, err := h.registerCmd.Execute(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domainErrors.ErrUserAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}

		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, resp)
}

func bindStrictJSON(c *gin.Context, dst interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(dst)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.SuccessResponse{data=dto.RefreshTokenResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()

	// Parse and validate the refresh token JWT
	claims, err := h.jwtMiddleware.ParseToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Invalid refresh token")
		return
	}

	// Check blocklist — token may have been explicitly revoked
	revoked, err := h.jwtMiddleware.IsRevoked(ctx, claims.JTI)
	if err != nil {
		response.InternalError(c, "Token validation unavailable")
		return
	}
	if revoked {
		response.Unauthorized(c, "Refresh token has been revoked")
		return
	}

	// Issue new access token
	accessToken, _, err := h.jwtMiddleware.GenerateAccessToken(claims.UserID)
	if err != nil {
		response.InternalError(c, "Failed to generate access token")
		return
	}

	// Rotate: issue a new refresh token
	newRefreshToken, _, err := h.jwtMiddleware.GenerateRefreshToken(claims.UserID)
	if err != nil {
		response.InternalError(c, "Failed to generate refresh token")
		return
	}

	// Revoke the old refresh token (remaining TTL from its expiry)
	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if remainingTTL > 0 {
		if err := h.jwtMiddleware.RevokeToken(ctx, claims.JTI, remainingTTL); err != nil {
			// Non-fatal: log and continue. Token will expire naturally.
			_ = err
		}
	}

	response.OK(c, dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(h.jwtMiddleware.AccessTTL().Seconds()),
	})
}

// AdminLogin handles admin login with email + password.
// @Summary Admin login
// @Tags admin-auth
// @Accept json
// @Produce json
// @Param request body dto.AdminLoginRequest true "Admin login request"
// @Success 200 {object} response.SuccessResponse{data=dto.AdminLoginResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /admin/auth/login [post]
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.adminLoginCmd.Execute(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, resp)
}

// AdminLogout revokes the provided refresh token.
// @Summary Admin logout
// @Tags admin-auth
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
func extractLogoutRefreshToken(r io.Reader) (string, bool, error) {
	if r == nil {
		return "", false, nil
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return "", false, errors.New("Failed to read request body")
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "", false, nil
	}
	if trimmed == "null" {
		return "", false, errors.New("Invalid request body")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", false, errors.New("Invalid request body")
	}
	rawToken, ok := raw["refresh_token"]
	if !ok {
		return "", false, nil
	}
	if strings.TrimSpace(string(rawToken)) == "null" {
		return "", false, errors.New("Invalid request body")
	}
	var token string
	if err := json.Unmarshal(rawToken, &token); err != nil || strings.TrimSpace(token) == "" {
		return "", false, errors.New("Invalid request body")
	}
	return token, true, nil
}

func (h *AuthHandler) revokeTokenIfValid(ctx context.Context, tokenString string) error {
	claims, err := h.jwtMiddleware.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if remainingTTL := time.Until(claims.ExpiresAt.Time); remainingTTL > 0 {
		_ = h.jwtMiddleware.RevokeToken(ctx, claims.JTI, remainingTTL)
	}
	return nil
}

// AdminLogout handles admin logout by revoking access and refresh tokens
// @Summary Admin logout
// @Tags admin-auth
// @Produce json
// @Param request body dto.AdminLogoutRequest false "Admin logout request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /admin/auth/logout [post]
func (h *AuthHandler) AdminLogout(c *gin.Context) {
	refreshToken, hasRefreshTokenField, err := extractLogoutRefreshToken(c.Request.Body)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	hasAuthHeader := false
	if authHeader := strings.TrimSpace(c.GetHeader("Authorization")); authHeader != "" {
		hasAuthHeader = true
		tokenString, ok := bearerTokenFromHeader(authHeader)
		if !ok {
			response.Unauthorized(c, "Invalid authorization header format")
			return
		}
		if err := h.revokeTokenIfValid(ctx, tokenString); err != nil {
			response.Unauthorized(c, "Invalid token")
			return
		}
	}

	if !hasAuthHeader && !hasRefreshTokenField {
		response.Unauthorized(c, "Missing authorization or refresh_token")
		return
	}

	if hasRefreshTokenField {
		if err := h.revokeTokenIfValid(ctx, refreshToken); err != nil {
			response.Unauthorized(c, "Invalid refresh token")
			return
		}
	}

	response.OK(c, gin.H{"message": "logged out"})
}

func bearerTokenFromHeader(authHeader string) (string, bool) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}
