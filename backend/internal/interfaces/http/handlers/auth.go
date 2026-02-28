package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/bivex/paywall-iap/internal/application/command"
	"github.com/bivex/paywall-iap/internal/application/dto"
	"github.com/bivex/paywall-iap/internal/application/middleware"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	registerCmd      *command.RegisterCommand
	jwtMiddleware    *middleware.JWTMiddleware
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(registerCmd *command.RegisterCommand, jwtMiddleware *middleware.JWTMiddleware) *AuthHandler {
	return &AuthHandler{
		registerCmd:   registerCmd,
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
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.registerCmd.Execute(c.Request.Context(), &req)
	if err != nil {
		// Handle specific error cases
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, resp)
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

	// Validate refresh token and generate new tokens
	// TODO: Implement refresh token validation and rotation

	// For now, return error indicating not implemented
	response.InternalError(c, "Refresh token not yet implemented")
}
