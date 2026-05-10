package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("default port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Environment != "development" {
		t.Errorf("environment = %s, want development", cfg.Server.Environment)
	}
	if cfg.ClickHouse.Host != "localhost" {
		t.Errorf("clickhouse host = %s, want localhost", cfg.ClickHouse.Host)
	}
	if cfg.ClickHouse.Port != 9000 {
		t.Errorf("clickhouse port = %d, want 9000", cfg.ClickHouse.Port)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("redis port = %d, want 6379", cfg.Redis.Port)
	}
	if cfg.Detection.GPUUtilizationThreshold != 20.0 {
		t.Errorf("gpu util threshold = %f, want 20.0", cfg.Detection.GPUUtilizationThreshold)
	}
}

func TestLoad_OverrideFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("CLICKHOUSE_HOST", "ch-server")
	os.Setenv("REDIS_PORT", "6380")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.ClickHouse.Host != "ch-server" {
		t.Errorf("clickhouse host = %s, want ch-server", cfg.ClickHouse.Host)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("redis port = %d, want 6380", cfg.Redis.Port)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("SERVER_PORT", "99999")

	_, err := Load()
	if err == nil {
		t.Error("expected validation error for invalid port")
	}
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "invalid")

	_, err := Load()
	if err == nil {
		t.Error("expected validation error for invalid environment")
	}
}

func TestValidate_JWTRequiredInProduction(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Error("expected validation error for missing JWT_SECRET in production")
	}
}

func TestValidate_InvalidGPUThreshold(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("DETECTION_GPU_UTIL_THRESHOLD", "150.0")

	_, err := Load()
	if err == nil {
		t.Error("expected validation error for GPU threshold > 100")
	}
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Environment: "development"}}
	if !cfg.IsDevelopment() {
		t.Error("IsDevelopment() should return true")
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() should return false")
	}
}
