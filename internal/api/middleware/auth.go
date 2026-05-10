package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// AuthRequired ensures that the request has a valid Bearer token.
// In a real production system, this would validate the JWT against an IDP or JWKS.
func AuthRequired(log *logger.Logger, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn().Str("path", c.Request.URL.Path).Msg("missing authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn().Str("path", c.Request.URL.Path).Msg("invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		token := parts[1]

		// For MVP: Simple hardcoded token verification.
		// Production: Use jwt.Parse(token, ...) to validate claims.
		if token != secret {
			log.Warn().Str("path", c.Request.URL.Path).Msg("invalid token provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", "admin") // Mock user ID for MVP
		c.Next()
	}
}
