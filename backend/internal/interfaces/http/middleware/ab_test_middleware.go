package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/bivex/paywall-iap/internal/domain/service"
)

// ABTestMiddleware handles A/B test variant assignment and tracking
type ABTestMiddleware struct {
	ffService *service.FeatureFlagService
}

// NewABTestMiddleware creates a new A/B test middleware
func NewABTestMiddleware(ffService *service.FeatureFlagService) *ABTestMiddleware {
	return &ABTestMiddleware{
		ffService: ffService,
	}
}

// AssignVariants assigns A/B test variants to the request context
func (m *ABTestMiddleware) AssignVariants() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.Next()
			return
		}

		// Evaluate paywall variant
		variant, err := m.ffService.EvaluatePaywallTest(c.Request.Context(), userID)
		if err == nil {
			c.Set("paywall_variant", variant)
		}

		c.Next()
	}
}
