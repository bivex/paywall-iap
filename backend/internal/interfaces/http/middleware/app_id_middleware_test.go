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

// TestRequireAppID_Whitespace: a header value that is only whitespace is not
// an empty string, so the "missing" check passes, but uuid.Parse rejects it
// with "invalid UUID length" → 422.
func TestRequireAppID_Whitespace(t *testing.T) {
	r := setupRouter()
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", " ")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "invalid X-App-ID")
}

// TestRequireAppID_UppercaseUUID: uuid.Parse accepts uppercase hex digits and
// normalises them to lowercase internally; the parsed value must equal the
// MustParse of the same string.
func TestRequireAppID_UppercaseUUID(t *testing.T) {
	r := setupRouter()
	var captured uuid.UUID
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		captured = middleware.GetAppID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	upperID := "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", upperID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uuid.MustParse(upperID), captured)
}

// TestRequireAppID_NilUUID: the all-zeros UUID is syntactically valid; the
// middleware must accept it and capture uuid.Nil. Filtering nil UUIDs is the
// responsibility of the handler, not the middleware.
func TestRequireAppID_NilUUID(t *testing.T) {
	r := setupRouter()
	var captured uuid.UUID
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		captured = middleware.GetAppID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", "00000000-0000-0000-0000-000000000000")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uuid.Nil, captured)
}

// TestRequireAppID_WithBraces: uuid.Parse accepts the Microsoft-style braced
// format "{xxxxxxxx-...}" and returns the correct UUID value → 200.
func TestRequireAppID_WithBraces(t *testing.T) {
	r := setupRouter()
	var captured uuid.UUID
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		captured = middleware.GetAppID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	bracedID := "{00000000-0000-0000-0000-000000000001}"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", bracedID)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uuid.MustParse(bracedID), captured)
}

// TestRequireAppID_HandlerNotCalledOnAbort: when the header is missing the
// middleware calls c.Abort(), so the downstream handler must never execute.
func TestRequireAppID_HandlerNotCalledOnAbort(t *testing.T) {
	r := setupRouter()
	handlerCalled := false
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil) // no X-App-ID header
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, handlerCalled, "downstream handler must not be called after Abort()")
}

// TestRequireAppID_MultipleHeaders: when the same header name is sent multiple
// times, gin's GetHeader returns the first value. The middleware must parse
// that first value and succeed.
func TestRequireAppID_MultipleHeaders(t *testing.T) {
	r := setupRouter()
	var captured uuid.UUID
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		captured = middleware.GetAppID(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	firstID := "00000000-0000-0000-0000-000000000001"
	req, _ := http.NewRequest("GET", "/test", nil)
	// Add returns the value in the canonical multi-value header form.
	req.Header.Add("X-App-ID", firstID)
	req.Header.Add("X-App-ID", "00000000-0000-0000-0000-000000000002")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uuid.MustParse(firstID), captured)
}

// TestRequireAppID_EmptyStringExplicit: an explicitly set empty-string header
// value is treated the same as a missing header by the middleware → 400.
// Note: net/http canonicalises an empty Add/Set as no header at all, so we
// confirm the behaviour matches the "missing" path.
func TestRequireAppID_EmptyStringExplicit(t *testing.T) {
	r := setupRouter()
	r.GET("/test", middleware.RequireAppID(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-ID", "")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "X-App-ID header required")
}
