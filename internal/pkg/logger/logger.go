// Package logger provides structured logging for Obsygnal-Pulse.
//
// We use zerolog because:
// 1. Zero-allocation JSON logging — critical for high-throughput telemetry paths
// 2. Structured fields are first-class, not afterthoughts
// 3. Level-based filtering with minimal overhead when disabled
// 4. Context propagation via logger embedding
//
// All logs are JSON in production (machine-parseable for Loki) and
// human-readable console output in development.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with application-specific methods.
type Logger struct {
	zerolog.Logger
}

// New creates a new Logger instance configured for the given environment.
func New(environment string) *Logger {
	var writer io.Writer

	if environment == "development" {
		// Console writer for human-readable dev output
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		// JSON output for production — compatible with Loki, Datadog, etc.
		writer = os.Stdout
	}

	level := zerolog.InfoLevel
	if environment == "development" {
		level = zerolog.DebugLevel
	}

	zl := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("service", "obsygnal-pulse").
		Caller().
		Logger()

	return &Logger{Logger: zl}
}

// WithComponent creates a child logger tagged with a component name.
// Use this to create module-specific loggers (e.g., "collector", "detection").
func (l *Logger) WithComponent(component string) *Logger {
	child := l.Logger.With().Str("component", component).Logger()
	return &Logger{Logger: child}
}

// WithRequestID creates a child logger tagged with a request ID for tracing.
func (l *Logger) WithRequestID(requestID string) *Logger {
	child := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{Logger: child}
}

// WithTraceID creates a child logger tagged with an OpenTelemetry trace ID.
func (l *Logger) WithTraceID(traceID string) *Logger {
	child := l.Logger.With().Str("trace_id", traceID).Logger()
	return &Logger{Logger: child}
}

// WithJobID creates a child logger tagged with a training job ID.
func (l *Logger) WithJobID(jobID string) *Logger {
	child := l.Logger.With().Str("job_id", jobID).Logger()
	return &Logger{Logger: child}
}

// WithNodeID creates a child logger tagged with a cluster node ID.
func (l *Logger) WithNodeID(nodeID string) *Logger {
	child := l.Logger.With().Str("node_id", nodeID).Logger()
	return &Logger{Logger: child}
}

// WithGPUID creates a child logger tagged with a GPU identifier.
func (l *Logger) WithGPUID(gpuID string) *Logger {
	child := l.Logger.With().Str("gpu_id", gpuID).Logger()
	return &Logger{Logger: child}
}
