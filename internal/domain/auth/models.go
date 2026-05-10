// Package auth defines authentication and authorization domain models.
package auth

import "time"

// User represents a platform user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never serialized
	DisplayName  string    `json:"display_name"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Role defines user access levels.
type Role string

const (
	RoleViewer  Role = "viewer"  // Read-only access to dashboards
	RoleEditor  Role = "editor"  // Can dismiss/apply recommendations
	RoleAdmin   Role = "admin"   // Full access including user management
)

// Permission defines what actions a role can perform.
type Permission string

const (
	PermViewDashboard       Permission = "view:dashboard"
	PermViewMetrics         Permission = "view:metrics"
	PermViewInefficiencies  Permission = "view:inefficiencies"
	PermViewRecommendations Permission = "view:recommendations"
	PermDismissRecommendation Permission = "dismiss:recommendation"
	PermApplyRecommendation   Permission = "apply:recommendation"
	PermManageAlerts        Permission = "manage:alerts"
	PermManageUsers         Permission = "manage:users"
)

// RolePermissions maps roles to their allowed permissions.
var RolePermissions = map[Role][]Permission{
	RoleViewer: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
	},
	RoleEditor: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
		PermDismissRecommendation, PermApplyRecommendation, PermManageAlerts,
	},
	RoleAdmin: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
		PermDismissRecommendation, PermApplyRecommendation, PermManageAlerts, PermManageUsers,
	},
}

// HasPermission checks if a role has a specific permission.
func (r Role) HasPermission(p Permission) bool {
	perms, ok := RolePermissions[r]
	if !ok {
		return false
	}
	for _, rp := range perms {
		if rp == p {
			return true
		}
	}
	return false
}
