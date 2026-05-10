package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListInefficiencies returns active inefficiencies.
func ListInefficiencies() gin.HandlerFunc {
	return func(c *gin.Context) {
		// MVP Mock
		c.JSON(http.StatusOK, gin.H{"data": []string{}})
	}
}

// ListRecommendations returns active recommendations.
func ListRecommendations() gin.HandlerFunc {
	return func(c *gin.Context) {
		// MVP Mock
		c.JSON(http.StatusOK, gin.H{"data": []string{}})
	}
}

// ApplyRecommendation applies a specific recommendation.
func ApplyRecommendation() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"status": "applied", "id": id})
	}
}

// DismissRecommendation dismisses a specific recommendation.
func DismissRecommendation() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"status": "dismissed", "id": id})
	}
}

// RecentAlerts returns recent alerts.
func RecentAlerts() gin.HandlerFunc {
	return func(c *gin.Context) {
		// MVP Mock
		c.JSON(http.StatusOK, gin.H{"data": []string{}})
	}
}
