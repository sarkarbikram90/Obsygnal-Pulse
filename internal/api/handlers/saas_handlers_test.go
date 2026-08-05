package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
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
	log := logger.New("test")
	authSvc := authsvc.NewService("test-jwt-secret-key-32-bytes-long", time.Hour, log)

	r := gin.New()
	r.GET("/keys", ListAPIKeys(authSvc))
	r.DELETE("/keys/:id", RevokeAPIKey(authSvc))
	r.POST("/upgrade", UpgradeTier())
	r.GET("/audit", ListAuditLogs(authSvc))

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

