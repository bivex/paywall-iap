package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Meta contains response metadata
type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message,omitempty"`
	Code     string `json:"code,omitempty"`
	Meta     Meta   `json:"meta"`
}

// Send sends a successful response
func Send(c *gin.Context, statusCode int, data interface{}) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	c.JSON(statusCode, SuccessResponse{
		Data: data,
		Meta: Meta{
			RequestID: requestID,
			Timestamp: time.Now(),
		},
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	Send(c, http.StatusCreated, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	Send(c, http.StatusOK, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, errCode string, message string) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	c.JSON(statusCode, ErrorResponse{
		Error:   errCode,
		Message: message,
		Meta: Meta{
			RequestID: requestID,
			Timestamp: time.Now(),
		},
	})
}

// Common error response helpers

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "INVALID_REQUEST", message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// RateLimited sends a 429 Too Many Requests response
func RateLimited(c *gin.Context, retryAfter int) {
	c.Header("Retry-After", string(retryAfter))
	Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// UnprocessableEntity sends a 422 Unprocessable Entity response
func UnprocessableEntity(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message)
}
