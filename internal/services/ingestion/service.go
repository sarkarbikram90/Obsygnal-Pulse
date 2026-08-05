// Package ingestion implements the core telemetry ingestion service pipeline.
//
// This service acts as the central data orchestrator for the platform, responsible for:
// 1. Receiving massive bursts of metric batches from external collectors.
// 2. Enforcing domain-level validation and normalization on raw telemetry.
// 3. Providing strict deduplication via Redis sliding windows to prevent data corruption.
// 4. Batch-buffering data in memory to enable highly efficient bulk ClickHouse inserts.
//
// By sitting at the center of the Hexagonal Architecture, this service is entirely
// agnostic to how data is received (HTTP vs Queue) and focuses solely on robust,
// high-throughput processing and persistence logic.
package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/metrics"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// Service orchestrates the ingestion pipeline.
type Service struct {
	metricsRepo ports.MetricsRepository
	cache       ports.CacheStore
	log         *logger.Logger

	// Buffered writer state
	gpuBuffer     []metrics.GPUMetric
	nodeBuffer    []metrics.NodeMetric
	bufferMu      sync.Mutex
	batchSize     int
	flushInterval time.Duration

	// Metrics
	totalReceived   int64
	totalInserted   int64
	totalDropped    int64
	totalDuplicates int64
}

// NewService creates a new ingestion service.
func NewService(
	metricsRepo ports.MetricsRepository,
	cache ports.CacheStore,
	log *logger.Logger,
	batchSize int,
	flushInterval time.Duration,
) *Service {
	return &Service{
		metricsRepo:   metricsRepo,
		cache:         cache,
		log:           log.WithComponent("ingestion"),
		gpuBuffer:     make([]metrics.GPUMetric, 0, batchSize),
		nodeBuffer:    make([]metrics.NodeMetric, 0, batchSize),
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

// ProcessBatch handles an incoming metric batch.
func (s *Service) ProcessBatch(ctx context.Context, batch *metrics.MetricBatch) error {
	if batch == nil || batch.Size() == 0 {
		return nil
	}

	s.totalReceived += int64(batch.Size())

	// Validate GPU metrics
	validGPU := make([]metrics.GPUMetric, 0, len(batch.GPUMetrics))
	for _, m := range batch.GPUMetrics {
		if err := m.Validate(); err != nil {
			s.log.Warn().
				Str("gpu_id", m.GPUID).
				Err(err).
				Msg("dropping invalid GPU metric")
			s.totalDropped++
			continue
		}

		// Check for duplicates using cache (if cache is provided)
		if s.cache != nil {
			dedupKey := fmt.Sprintf("dedup:%s:%d", m.GPUID, m.Timestamp.UnixMilli())
			exists, err := s.cache.Exists(ctx, dedupKey)
			if err != nil {
				// Cache failures should not block ingestion
				s.log.Warn().Err(err).Msg("dedup check failed, proceeding")
			} else if exists {
				s.totalDuplicates++
				continue
			}

			// Mark as seen (TTL = 5 minutes for dedup window)
			if err := s.cache.Set(ctx, dedupKey, true, 5*time.Minute); err != nil {
				s.log.Warn().Err(err).Msg("failed to set dedup key")
			}
		}

		validGPU = append(validGPU, m)
	}

	// Buffer metrics for batch insertion
	s.bufferMu.Lock()
	s.gpuBuffer = append(s.gpuBuffer, validGPU...)
	s.nodeBuffer = append(s.nodeBuffer, batch.NodeMetrics...)

	shouldFlush := len(s.gpuBuffer) >= s.batchSize || len(s.nodeBuffer) >= s.batchSize
	s.bufferMu.Unlock()

	if shouldFlush {
		return s.Flush(ctx)
	}

	return nil
}

// Flush writes buffered metrics to ClickHouse.
func (s *Service) Flush(ctx context.Context) error {
	s.bufferMu.Lock()
	gpuBatch := s.gpuBuffer
	nodeBatch := s.nodeBuffer
	s.gpuBuffer = make([]metrics.GPUMetric, 0, s.batchSize)
	s.nodeBuffer = make([]metrics.NodeMetric, 0, s.batchSize)
	s.bufferMu.Unlock()

	if len(gpuBatch) == 0 && len(nodeBatch) == 0 {
		return nil
	}

	var errs []error

	if len(gpuBatch) > 0 {
		if err := s.metricsRepo.InsertGPUMetrics(ctx, gpuBatch); err != nil {
			s.log.Error().Err(err).Int("count", len(gpuBatch)).Msg("failed to insert GPU metrics")
			errs = append(errs, err)
		} else {
			s.totalInserted += int64(len(gpuBatch))
			s.log.Debug().Int("count", len(gpuBatch)).Msg("flushed GPU metrics to ClickHouse")
		}
	}

	if len(nodeBatch) > 0 {
		if err := s.metricsRepo.InsertNodeMetrics(ctx, nodeBatch); err != nil {
			s.log.Error().Err(err).Int("count", len(nodeBatch)).Msg("failed to insert node metrics")
			errs = append(errs, err)
		} else {
			s.log.Debug().Int("count", len(nodeBatch)).Msg("flushed node metrics to ClickHouse")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("flush errors: %v", errs)
	}
	return nil
}

// StartPeriodicFlush starts a background goroutine that flushes at fixed intervals.
func (s *Service) StartPeriodicFlush(ctx context.Context) {
	ticker := time.NewTicker(s.flushInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// Final flush on shutdown
				s.Flush(context.Background())
				return
			case <-ticker.C:
				if err := s.Flush(ctx); err != nil {
					s.log.Error().Err(err).Msg("periodic flush failed")
				}
			}
		}
	}()
}

// Stats returns current ingestion statistics.
func (s *Service) Stats() map[string]int64 {
	s.bufferMu.Lock()
	gpuLen := int64(len(s.gpuBuffer))
	nodeLen := int64(len(s.nodeBuffer))
	s.bufferMu.Unlock()

	return map[string]int64{
		"total_received":   s.totalReceived,
		"total_inserted":   s.totalInserted,
		"total_dropped":    s.totalDropped,
		"total_duplicates": s.totalDuplicates,
		"buffer_gpu_size":  gpuLen,
		"buffer_node_size": nodeLen,
	}
}
