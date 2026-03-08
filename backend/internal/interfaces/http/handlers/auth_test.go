package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/bivex/paywall-iap/internal/application/middleware"
)

func TestAdminLogout_AcceptsHeaderOnlyLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtm := middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute)
	accessToken, _, err := jwtm.GenerateAccessToken("admin-user")
	require.NoError(t, err)

	handler := NewAuthHandler(nil, nil, jwtm)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+accessToken)

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusOK, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"logged out"`)
}

func TestAdminLogout_RejectsRequestWithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil, nil, middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", nil)

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusUnauthorized, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Missing authorization or refresh_token"`)
}

func TestAdminLogout_RejectsInvalidRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil, nil, middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", strings.NewReader(`{"refresh_token":"invalid"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusUnauthorized, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid refresh token"`)
}

func TestAdminLogout_RejectsNullRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil, nil, middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", strings.NewReader(`{"refresh_token":null}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid request body"`)
}

func TestAdminLogout_RejectsNullJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil, nil, middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", strings.NewReader(`null`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid request body"`)
}

func TestAdminLogout_RejectsEmptyRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler(nil, nil, middleware.NewJWTMiddleware("test-secret", redis.NewClient(&redis.Options{Addr: "localhost:0"}), time.Minute))
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", strings.NewReader(`{"refresh_token":""}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.AdminLogout(ctx)

	require.Equal(t, http.StatusBadRequest, recorder.Code, "body=%s", recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"Invalid request body"`)
}
