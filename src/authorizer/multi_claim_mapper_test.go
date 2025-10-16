package authorizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/log"
)

type fakeMapper struct {
	claims *authorization.Claims
	err    error
}

func (f fakeMapper) GetClaims(a *authorization.AuthInfo) (*authorization.Claims, error) {
	return f.claims, f.err
}

func TestMultiClaimMapper_OrderAndShortCircuit(t *testing.T) {
	logger := log.NewTestLogger()
	m := NewMultiClaimMapper(logger)
	// first returns empty -> continue
	m.Add("fakeMapper1", fakeMapper{claims: &authorization.Claims{}})
	// second returns real claims -> stop
	m.Add("fakeMapper2", fakeMapper{claims: &authorization.Claims{Subject: "ok", System: authorization.RoleAdmin}})

	claims, err := m.GetClaims(&authorization.AuthInfo{AuthToken: "x"})
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "ok", claims.Subject)
	assert.Equal(t, authorization.RoleAdmin, claims.System)
}

func TestMultiClaimMapper_ErrorPropagation(t *testing.T) {
	logger := log.NewTestLogger()
	m := NewMultiClaimMapper(logger)
	m.Add("fakeMapper1", fakeMapper{err: errors.New("boom")})
	m.Add("fakeMapper2", fakeMapper{claims: &authorization.Claims{Subject: "later"}})

	claims, err := m.GetClaims(&authorization.AuthInfo{AuthToken: "x"})
	assert.Nil(t, claims)
	require.Error(t, err)
}
