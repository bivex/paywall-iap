package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"sub"`
	JTI    string `json:"jti"` // JWT ID for revocation
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// JWTMiddleware handles JWT validation and revocation checking
type JWTMiddleware struct {
	secret          []byte
	refreshCache    *redis.Client
	accessTTL       time.Duration
	blocklistPrefix string
	logger          *zap.Logger
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(secret string, redisClient *redis.Client, accessTTL time.Duration) *JWTMiddleware {
	return &JWTMiddleware{
		secret:          []byte(secret),
		refreshCache:    redisClient,
		accessTTL:       accessTTL,
		blocklistPrefix: "jwt:blocked:",
		logger:          logging.Logger,
	}
}

// Authenticate validates the JWT token and sets user context
func (j *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "Missing authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return j.secret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token is revoked
		ctx := c.Request.Context()
		blocklisted, err := j.refreshCache.Get(ctx, j.blocklistPrefix+claims.JTI).Result()
		if err != nil && err != redis.Nil {
			j.logger.Error("failed to check token blocklist", zap.Error(err))
			// Fail closed for security
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SERVICE_UNAVAILABLE", "message": "Token validation unavailable"})
			c.Abort()
			return
		}

		if blocklisted != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_REVOKED", "message": "Token has been revoked"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("jti", claims.JTI)
		if claims.Role != "" {
			c.Set("role", claims.Role)
		}

		c.Next()
	}
}

// GenerateAccessToken creates a new access token
func (j *JWTMiddleware) GenerateAccessToken(userID string) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := &JWTClaims{
		UserID: userID,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			Issuer:    "iap-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", err
	}

	return tokenString, jti, nil
}

// GenerateRefreshToken creates a new refresh token with longer TTL
func (j *JWTMiddleware) GenerateRefreshToken(userID string) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := &JWTClaims{
		UserID: userID,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)), // 30 days
			Issuer:    "iap-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", err
	}

	return tokenString, jti, nil
}

// ParseToken parses a token string and returns the claims without checking the Redis blocklist.
// Useful for testing and internal token inspection.
func (j *JWTMiddleware) ParseToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// RevokeToken adds a token to the blocklist
func (j *JWTMiddleware) RevokeToken(ctx context.Context, jti string, remainingTTL time.Duration) error {
	return j.refreshCache.Set(ctx, j.blocklistPrefix+jti, "1", remainingTTL).Err()
}
