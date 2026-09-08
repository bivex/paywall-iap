package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("invalid token format")
	}
	return tokenString, nil
}

func parseAdminClaims(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func resolveAdminUserID(claims jwt.MapClaims) (uuid.UUID, error) {
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid token sub")
	}
	return uuid.Parse(userIDStr)
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware(userRepo repository.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		claims, err := parseAdminClaims(tokenString, jwtSecret)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		userID, err := resolveAdminUserID(claims)
		if err != nil {
			response.Unauthorized(c, "Invalid user ID in token")
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			response.Unauthorized(c, "User not found")
			c.Abort()
			return
		}

		if !user.IsAdmin() {
			response.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		c.Set("admin_id", userID)
		c.Next()
	}
}
