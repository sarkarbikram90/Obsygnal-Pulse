package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
)

// RegisterRequest holds payloads for user registration.
type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	DisplayName  string `json:"display_name" binding:"required"`
	OrgName      string `json:"org_name" binding:"required"`
}

// LoginRequest holds payloads for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterUser handles new tenant and user signup.
func RegisterUser(authSvc *authsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		userID := uuid.New().String()
		orgID := uuid.New().String()
		role := "owner"

		token, err := authSvc.GenerateToken(userID, orgID, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"token": token,
			"user": gin.H{
				"id":           userID,
				"email":        req.Email,
				"display_name": req.DisplayName,
			},
			"organization": gin.H{
				"id":   orgID,
				"name": req.OrgName,
				"tier": "pro",
				"role": role,
			},
		})
	}
}

// LoginUser handles authenticating an existing user.
func LoginUser(authSvc *authsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Demo authentication logic for SaaS starter
		userID := "usr_" + hex.EncodeToString([]byte(req.Email))[:12]
		orgID := "org_demo_101"
		role := "admin"

		token, err := authSvc.GenerateToken(userID, orgID, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":           userID,
				"email":        req.Email,
				"display_name": "AI Ops Admin",
			},
			"organization": gin.H{
				"id":   orgID,
				"name": "Obsygnal Labs Enterprise",
				"tier": "pro",
				"role": role,
			},
		})
	}
}

// GetCurrentUser returns authenticated context.
func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		orgID := c.GetString("org_id")
		role := c.GetString("role")

		if userID == "" {
			c.JSON(http.StatusOK, gin.H{
				"user_id": "usr_demo",
				"org_id":  "org_demo_101",
				"role":    "admin",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"org_id":  orgID,
			"role":    role,
		})
	}
}

// APIKeyRequest holds requests to create a new API key.
type APIKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateAPIKey generates a new agent telemetry key.
func CreateAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req APIKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rawBytes := make([]byte, 16)
		_, _ = rand.Read(rawBytes)
		keySecret := "obsygnal_sec_" + hex.EncodeToString(rawBytes)

		c.JSON(http.StatusCreated, gin.H{
			"id":          uuid.New().String(),
			"name":        req.Name,
			"prefix":      keySecret[:16] + "...",
			"secret_key":  keySecret, // Displayed once to user
			"created_at":  time.Now(),
			"description": req.Description,
		})
	}
}

// ListAPIKeys returns active API keys.
func ListAPIKeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"api_keys": []gin.H{
				{
					"id":          "key-01",
					"name":        "K8s-Production-Cluster-Agent",
					"prefix":      "obsygnal_sec_a8b9...",
					"created_at":  time.Now().Add(-72 * time.Hour),
					"last_used":   time.Now().Add(-2 * time.Minute),
					"status":      "active",
				},
				{
					"id":          "key-02",
					"name":        "Ray-Cluster-Collector",
					"prefix":      "obsygnal_sec_f4e1...",
					"created_at":  time.Now().Add(-24 * time.Hour),
					"last_used":   time.Now().Add(-15 * time.Second),
					"status":      "active",
				},
			},
		})
	}
}

// RevokeAPIKey deletes/revokes an API key.
func RevokeAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"status": "revoked",
			"id":     keyID,
		})
	}
}

// GetUsageSummary returns billing and GPU-hours usage data.
func GetUsageSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"organization_id": "org_demo_101",
			"tier":            "pro",
			"billing_cycle":   "August 2026",
			"monitored_gpus":  16,
			"gpu_hours_used":  3420.5,
			"gpu_hours_limit": 5000.0,
			"estimated_cost_usd": 12450.00,
			"estimated_savings_usd": 3820.00,
			"active_clusters": 2,
		})
	}
}

// UpgradeTier handles subscription upgrades.
func UpgradeTier() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "upgraded",
			"new_tier": "enterprise",
			"message": "Subscription plan successfully upgraded to Enterprise tier.",
		})
	}
}

// ListAuditLogs returns SOC2 compliance audit events.
func ListAuditLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"events": []gin.H{
				{
					"id":         uuid.New().String(),
					"actor":      "admin@obsygnal.io",
					"action":     "recommendation.applied",
					"resource":   "rec-idle-gpu-01",
					"timestamp":  time.Now().Add(-10 * time.Minute),
					"ip_address": "192.168.1.45",
				},
				{
					"id":         uuid.New().String(),
					"actor":      "api_key:obsygnal_sec_a8b9",
					"action":     "telemetry.batch_ingest",
					"resource":   "cluster-sim-01",
					"timestamp":  time.Now().Add(-1 * time.Minute),
					"ip_address": "10.244.0.12",
				},
			},
		})
	}
}
