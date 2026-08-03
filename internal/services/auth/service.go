// Package auth implements authentication, authorization, and multi-tenancy services.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/domain/auth"
	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
)

// Service handles user authentication, session management, and authorization.
type Service struct {
	jwtSecret []byte
	jwtExpiry time.Duration
	log       *logger.Logger
}

// NewService creates a new auth service.
func NewService(jwtSecret string, jwtExpiry time.Duration, log *logger.Logger) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
		log:       log.WithComponent("auth"),
	}
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
