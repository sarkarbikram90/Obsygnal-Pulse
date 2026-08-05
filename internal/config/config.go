// Package config provides centralized configuration management for Obsygnal-Pulse.
// All configuration is loaded from environment variables with sensible defaults.
// This approach follows 12-factor app principles and avoids config file sprawl.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server     ServerConfig
	ClickHouse ClickHouseConfig
	Redis      RedisConfig
	Prometheus PrometheusConfig
	Auth       AuthConfig
	Collector  CollectorConfig
	Detection  DetectionConfig
	Alerting   AlertingConfig
	Telemetry  TelemetryConfig
	Nats       NatsConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host               string
	Port               int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	Environment        string // "development", "staging", "production"
	CORSAllowedOrigins string
}

// ClickHouseConfig holds ClickHouse connection configuration.
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	// MaxOpenConns limits concurrent connections to prevent resource exhaustion.
	MaxOpenConns int
	// MaxIdleConns keeps warm connections to avoid cold-start latency on queries.
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	// BatchSize controls how many rows are buffered before flushing to ClickHouse.
	// Larger batches = fewer round trips but higher memory usage and insertion latency.
	BatchSize int
	// FlushInterval ensures data is written even during low-throughput periods.
	FlushInterval time.Duration
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// PrometheusConfig holds Prometheus query client configuration.
type PrometheusConfig struct {
	URL            string
	QueryTimeout   time.Duration
	ScrapeInterval time.Duration
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	BCryptCost         int
}

// CollectorConfig holds telemetry collector configuration.
type CollectorConfig struct {
	// ScrapeInterval controls how often the collector pulls from Prometheus.
	ScrapeInterval time.Duration
	// BatchSize controls how many metrics are batched before publishing to Redis Streams.
	BatchSize int
	// MaxConcurrentScrapes prevents overwhelming Prometheus under high cardinality.
	MaxConcurrentScrapes int
	// StalenessThreshold defines when a metric is considered stale and should trigger an alert.
	StalenessThreshold time.Duration
}

// DetectionConfig holds inefficiency detection engine configuration.
type DetectionConfig struct {
	// EvaluationInterval controls how often the detection engine runs.
	EvaluationInterval time.Duration
	// GPUUtilizationThreshold: GPUs below this % are flagged as underutilized.
	GPUUtilizationThreshold float64
	// MemoryUtilizationThreshold: GPU memory usage below this % is flagged.
	MemoryUtilizationThreshold float64
	// IdleTimeThreshold: GPUs idle longer than this are flagged.
	IdleTimeThreshold time.Duration
	// RetryThreshold: Jobs with retry count above this are flagged.
	RetryThreshold int
}

// AlertingConfig holds alerting system configuration.
type AlertingConfig struct {
	// SlackWebhookURL for Slack notifications.
	SlackWebhookURL string
	// WebhookURLs for generic webhook notifications.
	WebhookURLs []string
	// DeduplicationWindow prevents duplicate alerts within this window.
	DeduplicationWindow time.Duration
	// RateLimitPerMinute caps alert frequency per channel.
	RateLimitPerMinute int
}

// TelemetryConfig holds OpenTelemetry configuration.
type TelemetryConfig struct {
	ServiceName    string
	OTLPEndpoint   string
	SamplingRate   float64
	MetricsEnabled bool
	TracesEnabled  bool
}

// NatsConfig holds NATS JetStream configuration.
type NatsConfig struct {
	URL        string
	StreamName string
	MaxRetries int
	AckWait    time.Duration
}

// Load reads configuration from environment variables with defaults.
// We prefer env vars over config files because:
// 1. They work seamlessly with Docker, K8s, and CI/CD
// 2. Secrets never touch the filesystem
// 3. Environment-specific overrides don't require file management
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:               getEnv("SERVER_HOST", "0.0.0.0"),
			Port:               getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:        getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:       getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:        getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			Environment:        getEnv("ENVIRONMENT", "development"),
			CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
		ClickHouse: ClickHouseConfig{
			Host:            getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:            getEnvInt("CLICKHOUSE_PORT", 9000),
			Database:        getEnv("CLICKHOUSE_DATABASE", "obsygnal"),
			Username:        getEnv("CLICKHOUSE_USERNAME", "default"),
			Password:        getEnv("CLICKHOUSE_PASSWORD", ""),
			MaxOpenConns:    getEnvInt("CLICKHOUSE_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("CLICKHOUSE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("CLICKHOUSE_CONN_MAX_LIFETIME", 10*time.Minute),
			BatchSize:       getEnvInt("CLICKHOUSE_BATCH_SIZE", 10000),
			FlushInterval:   getEnvDuration("CLICKHOUSE_FLUSH_INTERVAL", 5*time.Second),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 20),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Prometheus: PrometheusConfig{
			URL:            getEnv("PROMETHEUS_URL", "http://localhost:9090"),
			QueryTimeout:   getEnvDuration("PROMETHEUS_QUERY_TIMEOUT", 30*time.Second),
			ScrapeInterval: getEnvDuration("PROMETHEUS_SCRAPE_INTERVAL", 15*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("JWT_SECRET", ""),
			JWTExpiry:          getEnvDuration("JWT_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			BCryptCost:         getEnvInt("BCRYPT_COST", 12),
		},
		Collector: CollectorConfig{
			ScrapeInterval:       getEnvDuration("COLLECTOR_SCRAPE_INTERVAL", 15*time.Second),
			BatchSize:            getEnvInt("COLLECTOR_BATCH_SIZE", 1000),
			MaxConcurrentScrapes: getEnvInt("COLLECTOR_MAX_CONCURRENT_SCRAPES", 5),
			StalenessThreshold:   getEnvDuration("COLLECTOR_STALENESS_THRESHOLD", 5*time.Minute),
		},
		Detection: DetectionConfig{
			EvaluationInterval:         getEnvDuration("DETECTION_EVALUATION_INTERVAL", 60*time.Second),
			GPUUtilizationThreshold:    getEnvFloat("DETECTION_GPU_UTIL_THRESHOLD", 20.0),
			MemoryUtilizationThreshold: getEnvFloat("DETECTION_MEM_UTIL_THRESHOLD", 15.0),
			IdleTimeThreshold:          getEnvDuration("DETECTION_IDLE_TIME_THRESHOLD", 30*time.Minute),
			RetryThreshold:             getEnvInt("DETECTION_RETRY_THRESHOLD", 3),
		},
		Alerting: AlertingConfig{
			SlackWebhookURL:     getEnv("ALERTING_SLACK_WEBHOOK_URL", ""),
			DeduplicationWindow: getEnvDuration("ALERTING_DEDUP_WINDOW", 15*time.Minute),
			RateLimitPerMinute:  getEnvInt("ALERTING_RATE_LIMIT_PER_MIN", 10),
		},
		Telemetry: TelemetryConfig{
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "obsygnal-pulse"),
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			SamplingRate:   getEnvFloat("OTEL_SAMPLING_RATE", 0.1),
			MetricsEnabled: getEnvBool("OTEL_METRICS_ENABLED", true),
			TracesEnabled:  getEnvBool("OTEL_TRACES_ENABLED", true),
		},
		Nats: NatsConfig{
			URL:        getEnv("NATS_URL", "nats://localhost:4222"),
			StreamName: getEnv("NATS_STREAM_NAME", "telemetry"),
			MaxRetries: getEnvInt("NATS_MAX_RETRIES", 3),
			AckWait:    getEnvDuration("NATS_ACK_WAIT", 30*time.Second),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that required configuration values are set and within valid ranges.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[c.Server.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, or production)", c.Server.Environment)
	}

	// JWT secret is required in non-development environments
	if c.Server.Environment != "development" && c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required in %s environment", c.Server.Environment)
	}

	if c.ClickHouse.BatchSize < 1 {
		return fmt.Errorf("invalid ClickHouse batch size: %d", c.ClickHouse.BatchSize)
	}

	if c.Detection.GPUUtilizationThreshold < 0 || c.Detection.GPUUtilizationThreshold > 100 {
		return fmt.Errorf("GPU utilization threshold must be 0-100, got: %f", c.Detection.GPUUtilizationThreshold)
	}

	return nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// Helper functions for reading typed environment variables.

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			return dur
		}
	}
	return defaultVal
}
