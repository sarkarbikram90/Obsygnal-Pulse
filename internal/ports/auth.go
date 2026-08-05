// Package ports — auth, API key, and audit log repository interfaces.
package ports

import (
	"context"
	"time"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/auth"
)

// AuthRepository defines persistence for users, orgs, and memberships.
type AuthRepository interface {
	CreateUser(ctx context.Context, user *auth.User) error
	GetUserByEmail(ctx context.Context, email string) (*auth.User, error)
	GetUserByID(ctx context.Context, id string) (*auth.User, error)
	CreateOrg(ctx context.Context, org *auth.Organization) error
	GetOrgByID(ctx context.Context, id string) (*auth.Organization, error)
	AddOrgMember(ctx context.Context, member *auth.OrgMember) error
	GetOrgMember(ctx context.Context, orgID, userID string) (*auth.OrgMember, error)
}

// APIKeyRepository defines persistence for machine API keys.
type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, key *auth.APIKey) error
	ListAPIKeys(ctx context.Context, orgID string) ([]auth.APIKey, error)
	RevokeAPIKey(ctx context.Context, id, orgID string) error
	GetAPIKeyByHash(ctx context.Context, hash string) (*auth.APIKey, error)
}

// AuditLogRepository defines persistence for compliance audit events.
type AuditLogRepository interface {
	InsertAuditEvent(ctx context.Context, event *auth.AuditEvent) error
	ListAuditEvents(ctx context.Context, orgID string, limit int) ([]auth.AuditEvent, error)
}

// UsageRepository defines persistence/metering for GPU-hours and billing metrics.
type UsageRepository interface {
	GetUsageSummary(ctx context.Context, orgID string) (*UsageSummary, error)
	RecordUsage(ctx context.Context, orgID string, gpuHours float64) error
}

// UsageSummary holds SaaS billing and usage statistics.
type UsageSummary struct {
	OrganizationID      string    `json:"organization_id"`
	Tier                string    `json:"tier"`
	BillingCycle        string    `json:"billing_cycle"`
	MonitoredGPUs       int       `json:"monitored_gpus"`
	GPUHoursUsed        float64   `json:"gpu_hours_used"`
	GPUHoursLimit       float64   `json:"gpu_hours_limit"`
	EstimatedCostUSD    float64   `json:"estimated_cost_usd"`
	EstimatedSavingsUSD float64   `json:"estimated_savings_usd"`
	ActiveClusters      int       `json:"active_clusters"`
	UpdatedAt           time.Time `json:"updated_at"`
}
