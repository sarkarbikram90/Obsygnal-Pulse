// Package auth defines authentication, authorization, and multi-tenancy models.
package auth

import "time"

// Organization represents a top-level tenant.
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Tier      string    `json:"tier"` // free, pro, enterprise
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a platform user.
type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	GlobalRole   Role       `json:"global_role,omitempty"` // For superadmins
	IsActive     bool       `json:"is_active"`
	MFAEnabled   bool       `json:"mfa_enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// OrgMember maps a user to an organization with a specific role.
type OrgMember struct {
	OrgID    string    `json:"org_id"`
	UserID   string    `json:"user_id"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// Project represents an isolated environment within an organization (e.g. cluster grouping).
type Project struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIKey represents a machine-to-machine authentication token.
type APIKey struct {
	ID        string     `json:"id"`
	OrgID     string     `json:"org_id"`
	ProjectID *string    `json:"project_id,omitempty"` // Optional scoping to project
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"`
	Prefix    string     `json:"prefix"` // First 8 chars for identification
	Role      Role       `json:"role"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy string     `json:"created_by"`
}

// ServiceAccount represents an identity for automated workloads.
type ServiceAccount struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Role defines access levels.
type Role string

const (
	RoleViewer   Role = "viewer"
	RoleOperator Role = "operator"
	RoleAdmin    Role = "admin"
	RoleBilling  Role = "billing"
	RoleOwner    Role = "owner"
)

// Permission defines what actions a role can perform.
type Permission string

const (
	PermViewDashboard         Permission = "view:dashboard"
	PermViewMetrics           Permission = "view:metrics"
	PermViewInefficiencies    Permission = "view:inefficiencies"
	PermViewRecommendations   Permission = "view:recommendations"
	PermManageRecommendations Permission = "manage:recommendations"
	PermManageAlerts          Permission = "manage:alerts"
	PermManageUsers           Permission = "manage:users"
	PermManageAPIKeys         Permission = "manage:apikeys"
	PermManageBilling         Permission = "manage:billing"
	PermManageOrg             Permission = "manage:org"
)

// RolePermissions maps roles to their allowed permissions.
var RolePermissions = map[Role][]Permission{
	RoleViewer: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
	},
	RoleOperator: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
		PermManageRecommendations, PermManageAlerts,
	},
	RoleAdmin: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
		PermManageRecommendations, PermManageAlerts, PermManageUsers, PermManageAPIKeys,
	},
	RoleBilling: {
		PermManageBilling,
	},
	RoleOwner: {
		PermViewDashboard, PermViewMetrics, PermViewInefficiencies, PermViewRecommendations,
		PermManageRecommendations, PermManageAlerts, PermManageUsers, PermManageAPIKeys,
		PermManageBilling, PermManageOrg,
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

// AuditEvent tracks actions taken by users or API keys for compliance and security.
type AuditEvent struct {
	ID        string            `json:"id"`
	OrgID     string            `json:"org_id"`
	ActorID   string            `json:"actor_id"`   // UserID or APIKeyID
	ActorType string            `json:"actor_type"` // "user", "api_key", "service_account"
	Action    string            `json:"action"`     // e.g., "user.invited", "recommendation.applied"
	Resource  string            `json:"resource"`   // e.g., "recommendation:123"
	Metadata  map[string]string `json:"metadata,omitempty"`
	IPAddress string            `json:"ip_address,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}
