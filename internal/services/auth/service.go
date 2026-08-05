// Package auth implements authentication, authorization, and multi-tenancy services.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/ports"
)

// Service handles user authentication, session management, and authorization.
type Service struct {
	jwtSecret   []byte
	jwtExpiry   time.Duration
	authRepo    ports.AuthRepository
	apiKeyRepo  ports.APIKeyRepository
	auditRepo   ports.AuditLogRepository
	usageRepo   ports.UsageRepository
	log         *logger.Logger
}

// Option configures Service dependencies.
type Option func(*Service)

func WithAuthRepo(repo ports.AuthRepository) Option {
	return func(s *Service) { s.authRepo = repo }
}

func WithAPIKeyRepo(repo ports.APIKeyRepository) Option {
	return func(s *Service) { s.apiKeyRepo = repo }
}

func WithAuditRepo(repo ports.AuditLogRepository) Option {
	return func(s *Service) { s.auditRepo = repo }
}

func WithUsageRepo(repo ports.UsageRepository) Option {
	return func(s *Service) { s.usageRepo = repo }
}

// NewService creates a new auth service.
func NewService(jwtSecret string, jwtExpiry time.Duration, log *logger.Logger, opts ...Option) *Service {
	s := &Service{
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
		log:       log.WithComponent("auth"),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// TokenClaims represents the claims in a multi-tenant JWT token.
type TokenClaims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	OrgID  string    `json:"org_id"`
	Role   auth.Role `json:"role"`
	Type   string    `json:"type"` // "user", "api_key", "service_account"
	jwt.RegisteredClaims
}

// GenerateUserToken creates a JWT token for a user within a specific organization.
func (s *Service) GenerateUserToken(user *auth.User, orgMember *auth.OrgMember) (string, error) {
	if len(s.jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not configured")
	}

	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		OrgID:  orgMember.OrgID,
		Role:   orgMember.Role,
		Type:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "obsygnal-pulse",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateToken generates a JWT string from userID, orgID, and role string.
func (s *Service) GenerateToken(userID, orgID, role string) (string, error) {
	if len(s.jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not configured")
	}

	claims := TokenClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   auth.Role(role),
		Type:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "obsygnal-pulse",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken verifies a JWT token and returns the claims securely.
func (s *Service) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// Authorize checks if a role has a specific permission.
func (s *Service) Authorize(role auth.Role, permission auth.Permission) bool {
	return role.HasPermission(permission)
}

// HashPassword generates a bcrypt hash of the password.
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash compares a password with a hash.
func (s *Service) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Register registers a new tenant organization and owner user.
func (s *Service) Register(ctx context.Context, email, password, displayName, orgName string) (*auth.User, *auth.Organization, string, error) {
	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &auth.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	org := &auth.Organization{
		ID:        uuid.New().String(),
		Name:      orgName,
		Slug:      orgName,
		Tier:      "pro",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	member := &auth.OrgMember{
		OrgID:    org.ID,
		UserID:   user.ID,
		Role:     auth.RoleOwner,
		JoinedAt: time.Now(),
	}

	if s.authRepo != nil {
		if err := s.authRepo.CreateUser(ctx, user); err != nil {
			return nil, nil, "", fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.authRepo.CreateOrg(ctx, org); err != nil {
			return nil, nil, "", fmt.Errorf("failed to create org: %w", err)
		}
		if err := s.authRepo.AddOrgMember(ctx, member); err != nil {
			return nil, nil, "", fmt.Errorf("failed to add org member: %w", err)
		}
	}

	token, err := s.GenerateUserToken(user, member)
	if err != nil {
		return nil, nil, "", err
	}

	_ = s.RecordAuditEvent(ctx, org.ID, user.ID, "user", "org.registered", "org:"+org.ID, nil)

	return user, org, token, nil
}

// Authenticate validates user credentials and returns a user, org member, and JWT token.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*auth.User, *auth.OrgMember, string, error) {
	if s.authRepo == nil {
		// Fallback for standalone demo mode
		userID := "usr_" + hex.EncodeToString([]byte(email))[:12]
		orgID := "org_demo_101"
		role := auth.RoleAdmin

		token, err := s.GenerateToken(userID, orgID, string(role))
		if err != nil {
			return nil, nil, "", err
		}
		user := &auth.User{ID: userID, Email: email, DisplayName: "AI Ops Admin", IsActive: true}
		member := &auth.OrgMember{OrgID: orgID, UserID: userID, Role: role}
		return user, member, token, nil
	}

	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, "", fmt.Errorf("invalid credentials")
	}

	if !s.CheckPasswordHash(password, user.PasswordHash) {
		return nil, nil, "", fmt.Errorf("invalid credentials")
	}

	orgID := "org_demo_101"
	member, err := s.authRepo.GetOrgMember(ctx, orgID, user.ID)
	if err != nil {
		// Default to owner/admin if single org
		member = &auth.OrgMember{OrgID: orgID, UserID: user.ID, Role: auth.RoleOwner, JoinedAt: time.Now()}
	}

	token, err := s.GenerateUserToken(user, member)
	if err != nil {
		return nil, nil, "", err
	}

	now := time.Now()
	user.LastLoginAt = &now

	_ = s.RecordAuditEvent(ctx, member.OrgID, user.ID, "user", "user.login", "user:"+user.ID, nil)

	return user, member, token, nil
}

// CreateAPIKey generates a machine API key and stores its SHA-256 hash.
func (s *Service) CreateAPIKey(ctx context.Context, orgID, createdBy, name string) (*auth.APIKey, string, error) {
	rawBytes := make([]byte, 16)
	_, _ = rand.Read(rawBytes)
	secretKey := "obsygnal_sec_" + hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(secretKey))
	keyHashStr := hex.EncodeToString(hash[:])

	apiKey := &auth.APIKey{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Name:      name,
		KeyHash:   keyHashStr,
		Prefix:    secretKey[:16] + "...",
		Role:      auth.RoleOperator,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	if s.apiKeyRepo != nil {
		if err := s.apiKeyRepo.CreateAPIKey(ctx, apiKey); err != nil {
			return nil, "", fmt.Errorf("failed to save API key: %w", err)
		}
	}

	_ = s.RecordAuditEvent(ctx, orgID, createdBy, "user", "apikey.created", "apikey:"+apiKey.ID, map[string]string{"name": name})

	return apiKey, secretKey, nil
}

// ListAPIKeys returns active API keys for an org.
func (s *Service) ListAPIKeys(ctx context.Context, orgID string) ([]auth.APIKey, error) {
	if s.apiKeyRepo != nil {
		return s.apiKeyRepo.ListAPIKeys(ctx, orgID)
	}
	return []auth.APIKey{
		{ID: "key-01", OrgID: orgID, Name: "K8s-Production-Cluster-Agent", Prefix: "obsygnal_sec_a8b9...", Role: auth.RoleOperator, CreatedAt: time.Now().Add(-72 * time.Hour)},
		{ID: "key-02", OrgID: orgID, Name: "Ray-Cluster-Collector", Prefix: "obsygnal_sec_f4e1...", Role: auth.RoleOperator, CreatedAt: time.Now().Add(-24 * time.Hour)},
	}, nil
}

// RevokeAPIKey revokes an API key.
func (s *Service) RevokeAPIKey(ctx context.Context, id, orgID string) error {
	if s.apiKeyRepo != nil {
		if err := s.apiKeyRepo.RevokeAPIKey(ctx, id, orgID); err != nil {
			return err
		}
	}
	_ = s.RecordAuditEvent(ctx, orgID, "admin", "user", "apikey.revoked", "apikey:"+id, nil)
	return nil
}

// RecordAuditEvent logs a compliance audit event.
func (s *Service) RecordAuditEvent(ctx context.Context, orgID, actorID, actorType, action, resource string, metadata map[string]string) error {
	if s.auditRepo == nil {
		return nil
	}
	evt := &auth.AuditEvent{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		ActorID:   actorID,
		ActorType: actorType,
		Action:    action,
		Resource:  resource,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	return s.auditRepo.InsertAuditEvent(ctx, evt)
}

// ListAuditEvents retrieves audit logs.
func (s *Service) ListAuditEvents(ctx context.Context, orgID string, limit int) ([]auth.AuditEvent, error) {
	if s.auditRepo != nil {
		return s.auditRepo.ListAuditEvents(ctx, orgID, limit)
	}
	return []auth.AuditEvent{
		{ID: uuid.New().String(), OrgID: orgID, ActorID: "admin@obsygnal.io", ActorType: "user", Action: "recommendation.applied", Resource: "rec-idle-gpu-01", CreatedAt: time.Now().Add(-10 * time.Minute)},
		{ID: uuid.New().String(), OrgID: orgID, ActorID: "obsygnal_sec_a8b9", ActorType: "api_key", Action: "telemetry.batch_ingest", Resource: "cluster-sim-01", CreatedAt: time.Now().Add(-1 * time.Minute)},
	}, nil
}

