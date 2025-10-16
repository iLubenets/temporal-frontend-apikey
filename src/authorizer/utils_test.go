package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/server/common/authorization"
)

func TestHasClaims_NilClaims(t *testing.T) {
	result := hasClaims(nil)
	assert.False(t, result)
}

func TestHasClaims_EmptyClaims(t *testing.T) {
	claims := &authorization.Claims{
		System:     authorization.RoleUndefined,
		Namespaces: map[string]authorization.Role{},
	}
	result := hasClaims(claims)
	assert.False(t, result)
}

func TestHasClaims_WithSystemRole(t *testing.T) {
	claims := &authorization.Claims{
		System:     authorization.RoleAdmin,
		Namespaces: map[string]authorization.Role{},
	}
	result := hasClaims(claims)
	assert.True(t, result)
}

func TestHasClaims_WithNamespaces(t *testing.T) {
	claims := &authorization.Claims{
		System: authorization.RoleUndefined,
		Namespaces: map[string]authorization.Role{
			"namespace1": authorization.RoleReader,
		},
	}
	result := hasClaims(claims)
	assert.True(t, result)
}

func TestHasClaims_WithBothSystemAndNamespaces(t *testing.T) {
	claims := &authorization.Claims{
		System: authorization.RoleWriter,
		Namespaces: map[string]authorization.Role{
			"namespace1": authorization.RoleReader,
			"namespace2": authorization.RoleWriter,
		},
	}
	result := hasClaims(claims)
	assert.True(t, result)
}

func TestHasClaims_NilNamespacesMap(t *testing.T) {
	claims := &authorization.Claims{
		System:     authorization.RoleUndefined,
		Namespaces: nil,
	}
	result := hasClaims(claims)
	assert.False(t, result)
}

func TestHasClaims_WithSystemRoleAndNilNamespaces(t *testing.T) {
	claims := &authorization.Claims{
		System:     authorization.RoleReader,
		Namespaces: nil,
	}
	result := hasClaims(claims)
	assert.True(t, result)
}
