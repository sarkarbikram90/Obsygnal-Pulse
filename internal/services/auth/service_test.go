package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sarkarbikram90/Obsygnal-Pulse/internal/pkg/logger"
	authsvc "github.com/sarkarbikram90/Obsygnal-Pulse/internal/services/auth"
)

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	log := logger.New("test")
	svc := authsvc.NewService("test-secret-key-32-bytes-long-string", 1*time.Hour, log)

	userID := "usr_101"
	orgID := "org_202"
	role := "admin"

	tokenStr, err := svc.GenerateToken(userID, orgID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	claims, err := svc.ValidateToken(tokenStr)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, orgID, claims.OrgID)
	assert.Equal(t, role, string(claims.Role))
}

func TestAuthService_PasswordHashing(t *testing.T) {
	log := logger.New("test")
	svc := authsvc.NewService("test-secret", 1*time.Hour, log)

	rawPassword := "SecurePassword123!"
	hash, err := svc.HashPassword(rawPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.True(t, svc.CheckPasswordHash(rawPassword, hash))
	assert.False(t, svc.CheckPasswordHash("WrongPassword!", hash))
}

func TestAuthService_InvalidToken(t *testing.T) {
	log := logger.New("test")
	svc := authsvc.NewService("test-secret-key", 1*time.Hour, log)

	_, err := svc.ValidateToken("invalid.token.string")
	assert.Error(t, err)
}
