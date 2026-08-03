package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", GetCurrentUser())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSaaSHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/keys", ListAPIKeys())
	r.DELETE("/keys/:id", RevokeAPIKey())
	r.POST("/upgrade", UpgradeTier())
	r.GET("/audit", ListAuditLogs())

	// Test ListAPIKeys
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/keys", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test RevokeAPIKey
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/keys/key-01", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test UpgradeTier
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/upgrade", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test ListAuditLogs
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/audit", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
