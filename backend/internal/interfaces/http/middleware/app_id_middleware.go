package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const AppIDKey = "app_id"

// RequireAppID reads X-App-ID header, validates UUID, stores in gin context.
// Returns 400 if missing, 422 if not a valid UUID.
// Use GetAppID(c) to retrieve the value in handlers.
func RequireAppID() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-App-ID")
		if raw == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-App-ID header required"})
			c.Abort()
			return
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid X-App-ID"})
			c.Abort()
			return
		}
		c.Set(AppIDKey, id)
		c.Next()
	}
}

// GetAppID retrieves the validated app UUID from gin context.
// Must only be called inside handlers registered under RequireAppID().
func GetAppID(c *gin.Context) uuid.UUID {
	return c.MustGet(AppIDKey).(uuid.UUID)
}
