package middleware

import (
	"github.com/bivex/paywall-iap/internal/appctx"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ContextKeyAppID = "app_id"

// RequireAppID is a Gin middleware that ensures app_id is present.
// It checks (in order):
//  1. app_id from JWT claims (set by Authenticate for player tokens)
//  2. X-App-ID request header (used by admin UI when selecting an app)
//
// Aborts with 400 if neither is present.
func RequireAppID() gin.HandlerFunc {
	return func(c *gin.Context) {
		appID, exists := c.Get(ContextKeyAppID)
		if !exists || appID == "" {
			// Fall back to X-App-ID header (admin UI sends this)
			headerID := c.GetHeader("X-App-ID")
			if headerID == "" {
				response.BadRequest(c, "Missing app_id: provide X-App-ID header or include app_id in JWT")
				c.Abort()
				return
			}
			parsed, err := uuid.Parse(headerID)
			if err != nil {
				response.BadRequest(c, "Invalid X-App-ID: must be a valid UUID")
				c.Abort()
				return
			}
			c.Set(ContextKeyAppID, headerID)
			r := c.Request.WithContext(appctx.WithAppID(c.Request.Context(), parsed))
			c.Request = r
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
