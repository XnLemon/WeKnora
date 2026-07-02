package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestModelHeaderForwardingCapturesOnlyAllowedNonReservedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ModelHeaderForwarding(&config.ModelHeaderForwardingConfig{
		Enabled: true,
		Allow:   []string{"X-Trace-Id", "X-User-Id", "Authorization"},
	}))
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.ModelForwardHeadersFromContext(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-Id", "trace-1")
	req.Header.Set("X-User-Id", "user-1")
	req.Header.Set("X-Tenant-Id", "tenant-ignored")
	req.Header.Set("Authorization", "Bearer should-not-forward")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"X-Trace-Id":"trace-1","X-User-Id":"user-1"}`, w.Body.String())
}

func TestModelHeaderForwardingCanExplicitlyAllowReservedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ModelHeaderForwarding(&config.ModelHeaderForwardingConfig{
		Enabled:       true,
		ReservedAllow: []string{"Authorization"},
	}))
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.ModelForwardHeadersFromContext(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer gateway-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"Authorization":"Bearer gateway-token"}`, w.Body.String())
}
