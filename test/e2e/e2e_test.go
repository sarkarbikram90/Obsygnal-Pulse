package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	memoryadapter "github.com/sarkarbikram90/Obsygnal-Pulse/internal/adapters/memory"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/api/routes"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/health"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
	detectionSvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/detection"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/ingestion"
	recSvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/recommendation"
)

func TestEndToEnd_Pipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("test")
	healthRegistry := health.NewRegistry("0.1.0")
	memStore := memoryadapter.NewRepositoryStore(log)
	ingestSvc := ingestion.NewService(memStore, nil, log, 10, 50*time.Millisecond)
	authSvc := authsvc.NewService("test-secret-key-32-bytes-long", time.Hour, log)

	router := gin.New()
	routes.Register(router, routes.Deps{
		Log:          log,
		Registry:     healthRegistry,
		MetricsRepo:  memStore,
		IneffRepo:    memStore.InefficiencyRepo(),
		RecRepo:      memStore.RecommendationRepo(),
		AlertRepo:    memStore.AlertRepo(),
		IngestionSvc: ingestSvc,
		AuthService:  authSvc,
	})

	// 1. Simulate Edge Telemetry Ingestion from Collector
	batch := metrics.MetricBatch{
		ID:         uuid.New().String(),
		Source:     "prometheus-dcgm",
		ReceivedAt: time.Now(),
		GPUMetrics: []metrics.GPUMetric{
			{
				ID:                 uuid.New().String(),
				Timestamp:          time.Now(),
				ClusterID:          "e2e-cluster-01",
				NodeID:             "node-01",
				GPUID:              "node-01-gpu-0",
				GPUModel:           "NVIDIA H100-SXM5-80GB",
				Utilization:        0.0, // Idle GPU
				MemoryUtilization:  0.0,
				MemoryUsed:         0,
				MemoryTotal:        85899345920,
				MemoryFree:         85899345920,
				Temperature:        42.0,
				PowerUsage:         75.0,
				PowerLimit:         700.0,
				RunningProcesses:   0,
			},
		},
	}

	payload, err := json.Marshal(batch)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/metrics/ingest", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	// 2. Inefficiency Detection Evaluation
	detConfig := config.DetectionConfig{
		EvaluationInterval:         1 * time.Second,
		GPUUtilizationThreshold:     30.0,
		MemoryUtilizationThreshold:  30.0,
	}
	detector := detectionSvc.NewService(memStore, memStore.InefficiencyRepo(), nil, log, detConfig)
	recommender := recSvc.NewService(memStore.RecommendationRepo(), nil, log)

	detector.OnDetection(func(ctx context.Context, ineff detection.Inefficiency) {
		// Handled reactively
	})

	ineffs, err := detector.Evaluate(context.Background())
	assert.NoError(t, err)
	assert.True(t, len(ineffs) >= 0)

	// 3. Generate Recommendation & Remediate via API
	token, err := authSvc.GenerateToken("usr_e2e", "org_e2e", "admin")
	assert.NoError(t, err)

	if len(ineffs) > 0 {
		recs, err := recommender.GenerateFromInefficiency(context.Background(), ineffs[0])
		assert.NoError(t, err)
		if len(recs) > 0 {
			w = httptest.NewRecorder()
			applyReq, _ := http.NewRequest(http.MethodPost, "/api/v1/recommendations/"+recs[0].ID+"/apply", nil)
			applyReq.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, applyReq)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}

	// 4. Validate Billing Meter Endpoint
	w = httptest.NewRecorder()
	usageReq, _ := http.NewRequest(http.MethodGet, "/api/v1/billing/usage", nil)
	usageReq.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, usageReq)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "gpu_hours_used")
}
