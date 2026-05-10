// Package tracing provides OpenTelemetry tracing utilities.
//
// This package wraps the OpenTelemetry SDK to provide a simple interface
// for creating spans, adding attributes, and propagating trace context.
package tracing

import (
	"context"
)

// SpanOption configures a span.
type SpanOption func(*spanConfig)

type spanConfig struct {
	attributes map[string]string
}

// WithAttribute adds a key-value attribute to the span.
func WithAttribute(key, value string) SpanOption {
	return func(cfg *spanConfig) {
		if cfg.attributes == nil {
			cfg.attributes = make(map[string]string)
		}
		cfg.attributes[key] = value
	}
}

// StartSpan creates a new span in the current trace.
// In MVP, this is a no-op that returns the context unchanged.
// When OpenTelemetry is configured, it will create real spans.
func StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, func()) {
	// TODO: Initialize OpenTelemetry tracer and create real spans
	// tracer := otel.Tracer("obsygnal-pulse")
	// ctx, span := tracer.Start(ctx, name)
	// for _, opt := range opts {
	//     cfg := &spanConfig{}
	//     opt(cfg)
	//     for k, v := range cfg.attributes {
	//         span.SetAttributes(attribute.String(k, v))
	//     }
	// }
	// return ctx, func() { span.End() }

	return ctx, func() {} // No-op for MVP
}

// TraceID extracts the trace ID from context, if present.
func TraceID(ctx context.Context) string {
	// TODO: Extract from otel span context
	return ""
}
