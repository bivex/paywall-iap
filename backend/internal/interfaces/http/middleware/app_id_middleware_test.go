package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bivex/paywall-iap/internal/interfaces/http/middleware"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestRequireAppID_Missing(t *testing.T) {
	r := setupRouter()
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireAppID_Invalid(t *testing.T) {
	r := setupRouter()
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", "not-a-uuid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestRequireAppID_Valid(t *testing.T) {
	r := setupRouter()
	var captured uuid.UUID
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		captured = middleware.GetAppID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	validID := "00000000-0000-0000-0000-000000000001"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", validID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uuid.MustParse(validID), captured)
}
