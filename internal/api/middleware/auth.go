// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
)

// AuthRequired ensures that the request has a valid Bearer JWT.
func AuthRequired(log *logger.Logger, authService *authsvc.Service) gin.HandlerFunc {
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

		tokenString := parts[1]

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			log.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("invalid token provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Set multi-tenant context variables
		c.Set("user_id", claims.UserID)
		c.Set("org_id", claims.OrgID)
		c.Set("role", claims.Role)
		c.Set("auth_type", claims.Type) // "user", "api_key", "service_account"
		c.Next()
	}
}

// RequirePermission ensures the authenticated identity has a specific permission.
func RequirePermission(log *logger.Logger, authService *authsvc.Service, reqPerm auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var role auth.Role
		switch v := roleVal.(type) {
		case auth.Role:
			role = v
		case string:
			role = auth.Role(v)
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid role context"})
			return
		}

		if !authService.Authorize(role, reqPerm) {
			log.Warn().
				Str("user_id", c.GetString("user_id")).
				Str("org_id", c.GetString("org_id")).
				Str("role", string(role)).
				Str("required_permission", string(reqPerm)).
				Msg("permission denied")

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}

// TenantScope enforces that requests can only access data belonging to their Organization.
// It checks if an 'org_id' path param or query param matches the JWT's org_id.
// Superadmins (RoleOwner) might be permitted to bypass this in some architectures,
// but for standard RBAC, strict tenant isolation is required.
func TenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenOrgID := c.GetString("org_id")

		// Typically, the org ID is in the URL: /api/v1/orgs/:org_id/...
		reqOrgID := c.Param("org_id")
		if reqOrgID == "" {
			reqOrgID = c.Query("org_id")
		}

		if reqOrgID != "" && reqOrgID != tokenOrgID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tenant isolation violation"})
			return
		}

		c.Next()
	}
}
