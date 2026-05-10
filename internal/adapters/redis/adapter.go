// Package redis implements the CacheStore port using Redis.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/config"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// Adapter implements ports.CacheStore for Redis.
type Adapter struct {
	client *redis.Client
	log    *logger.Logger
}


// NewAdapter creates a new Redis adapter.
func NewAdapter(cfg config.RedisConfig, log *logger.Logger) (*Adapter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Msg("connected to Redis")

	return &Adapter{client: client, log: log.WithComponent("redis")}, nil
}

// Get retrieves a value from cache.
func (a *Adapter) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := a.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with TTL.
func (a *Adapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return a.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from cache.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	return a.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	n, err := a.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return n > 0, nil
}

// Ping checks the Redis connection.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (a *Adapter) Close() error {
	return a.client.Close()
}

// Client returns the underlying Redis client for advanced operations.
func (a *Adapter) Client() *redis.Client {
	return a.client
}
