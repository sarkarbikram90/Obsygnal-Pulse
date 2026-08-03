package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_HasPermission(t *testing.T) {
	assert.True(t, RoleAdmin.HasPermission(PermManageUsers))
	assert.False(t, RoleViewer.HasPermission(PermManageUsers))
	assert.True(t, RoleOwner.HasPermission(PermManageOrg))
	assert.False(t, Role("unknown").HasPermission(PermViewDashboard))
}
