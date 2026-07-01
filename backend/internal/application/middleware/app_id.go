package middleware

import (
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/gin-gonic/gin"
)

const ContextKeyAppID = "app_id"

// RequireAppID is a Gin middleware that aborts with 401 if the
// app_id claim is missing from the JWT context (set by Authenticate).
// Mount it after JWTMiddleware.Authenticate() on all player-facing routes.
func RequireAppID() gin.HandlerFunc {
	return func(c *gin.Context) {
		appID, exists := c.Get(ContextKeyAppID)
		if !exists || appID == "" {
			response.Unauthorized(c, "Missing app_id claim in token")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAppID is a helper that retrieves the app_id string from gin context.
// Returns empty string if not set.
func GetAppID(c *gin.Context) string {
	v, _ := c.Get(ContextKeyAppID)
	s, _ := v.(string)
	return s
}
