// Package tracing provides OpenTelemetry tracing initialization and lifecycle management.
package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// Provider manages OpenTelemetry tracer provider lifecycle.
type Provider struct {
	tp  *sdktrace.TracerProvider
	log *logger.Logger
}

// InitTracer initializes OpenTelemetry distributed tracing.
func InitTracer(ctx context.Context, cfg config.TelemetryConfig, log *logger.Logger) (*Provider, error) {
	if !cfg.TracesEnabled {
		log.Info().Msg("OpenTelemetry tracing is disabled by configuration")
		return &Provider{log: log.WithComponent("tracing")}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRate))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info().
		Str("service_name", cfg.ServiceName).
		Float64("sampling_rate", cfg.SamplingRate).
		Msg("OpenTelemetry tracer provider initialized")

	return &Provider{tp: tp, log: log.WithComponent("tracing")}, nil
}

// Shutdown gracefully flushes and stops the tracer provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tp == nil {
		return nil
	}
	p.log.Info().Msg("flushing and shutting down OpenTelemetry tracer provider")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return p.tp.Shutdown(shutdownCtx)
}

// Tracer returns a named OpenTelemetry tracer.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
