package authorizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
)

type fakeMapper struct {
	claims *authorization.Claims
	err    error
}

func (f fakeMapper) GetClaims(a *authorization.AuthInfo) (*authorization.Claims, error) {
	return f.claims, f.err
}

func TestMultiClaimMapper_OrderAndShortCircuit(t *testing.T) {
	m := NewMultiClaimMapper()
	// first returns empty -> continue
	m.Add(fakeMapper{claims: &authorization.Claims{}})
	// second returns real claims -> stop
	m.Add(fakeMapper{claims: &authorization.Claims{Subject: "ok", System: authorization.RoleAdmin}})

	claims, err := m.GetClaims(&authorization.AuthInfo{AuthToken: "x"})
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "ok", claims.Subject)
	assert.Equal(t, authorization.RoleAdmin, claims.System)
}

func TestMultiClaimMapper_ErrorPropagation(t *testing.T) {
	m := NewMultiClaimMapper()
	m.Add(fakeMapper{err: errors.New("boom")})
	m.Add(fakeMapper{claims: &authorization.Claims{Subject: "later"}})

	claims, err := m.GetClaims(&authorization.AuthInfo{AuthToken: "x"})
	assert.Nil(t, claims)
	require.Error(t, err)
}
