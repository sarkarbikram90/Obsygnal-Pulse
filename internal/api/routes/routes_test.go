package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/memory"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("test")
	healthRegistry := health.NewRegistry("0.1.0")
	memStore := memory.NewRepositoryStore(log)
	ingestSvc := ingestion.NewService(memStore, nil, log, 10, 50*time.Millisecond)
	authSvc := authsvc.NewService("test-secret-key-32-bytes-long", time.Hour, log)

	r := gin.New()
	Register(r, Deps{
		Log:          log,
		Registry:     healthRegistry,
		MetricsRepo:  memStore,
		IneffRepo:    memStore.InefficiencyRepo(),
		RecRepo:      memStore.RecommendationRepo(),
		AlertRepo:    memStore.AlertRepo(),
		IngestionSvc: ingestSvc,
		AuthService:  authSvc,
	})

	// Test GET /api/v1/health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test GET /metrics
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
