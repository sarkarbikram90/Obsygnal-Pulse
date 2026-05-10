package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// RateLimiter returns a middleware that limits requests using a Redis sliding window.
func RateLimiter(redisClient *redis.Client, log *logger.Logger, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		now := time.Now().UnixNano()
		windowStart := now - window.Nanoseconds()

		// Redis transaction (pipeline)
		pipe := redisClient.Pipeline()

		// Remove old requests outside the window
		pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))
		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		// Count requests in the window
		countCmd := pipe.ZCard(ctx, key)
		// Set TTL to window duration
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Error().Err(err).Str("ip", clientIP).Msg("rate limiter redis error, failing open")
			c.Next() // Fail open if Redis is down
			return
		}

		count := countCmd.Val()

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)))

		if count > int64(limit) {
			log.Warn().Str("ip", clientIP).Int64("count", count).Msg("rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}
