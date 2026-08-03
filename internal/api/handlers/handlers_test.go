package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	memoryadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/memory"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/routes"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
)

func setupTestRouter() (*gin.Engine, *authsvc.Service) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logger.New("test")
	healthRegistry := health.NewRegistry("0.1.0")
	memStore := memoryadapter.NewRepositoryStore(log)
	ingestSvc := ingestion.NewService(memStore, nil, log, 10, 100*time.Millisecond)
	authSvc := authsvc.NewService("test-secret-key-32-bytes-long", time.Hour, log)

	routes.Register(r, routes.Deps{
		Log:          log,
		Registry:     healthRegistry,
		MetricsRepo:  memStore,
		IneffRepo:    memStore,
		RecRepo:      memStore,
		AlertRepo:    memStore,
		IngestionSvc: ingestSvc,
		AuthService:  authSvc,
	})

	return r, authSvc
}

func TestHealthCheckEndpoint(t *testing.T) {
	r, _ := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "version")
}

func TestAuthRegisterAndLogin(t *testing.T) {
	r, _ := setupTestRouter()

	// Register
	regPayload := map[string]string{
		"email":        "test@obsygnal.io",
		"password":     "SecretPass123!",
		"display_name": "Test Engineer",
		"org_name":     "Test AI Corp",
	}
	body, _ := json.Marshal(regPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "token")

	// Login
	loginPayload := map[string]string{
		"email":    "test@obsygnal.io",
		"password": "SecretPass123!",
	}
	body, _ = json.Marshal(loginPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}

func TestInefficienciesAndRecommendationsEndpoints(t *testing.T) {
	r, authSvc := setupTestRouter()
	token, _ := authSvc.GenerateToken("usr_test", "org_test", "admin")

	// List Inefficiencies
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/inefficiencies?cluster_id=sim-cluster-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data")

	// List Recommendations
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/recommendations?cluster_id=sim-cluster-01", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Apply Recommendation
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/recommendations/rec-001/apply", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "applied")
}

func TestAPIKeysAndBillingEndpoints(t *testing.T) {
	r, authSvc := setupTestRouter()
	token, _ := authSvc.GenerateToken("usr_test", "org_test", "admin")

	// Create API Key
	keyPayload := map[string]string{
		"name":        "Test Agent Key",
		"description": "Collector Key for Test Suite",
	}
	body, _ := json.Marshal(keyPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/api-keys", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "secret_key")

	// Get Billing Usage
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/billing/usage", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "gpu_hours_used")
}
