package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/log"
)

func TestPermissionToRole(t *testing.T) {
	assert.Equal(t, authorization.RoleReader, permissionToRole("read"))
	assert.Equal(t, authorization.RoleReader, permissionToRole("READ"))
	assert.Equal(t, authorization.RoleWriter, permissionToRole("write"))
	assert.Equal(t, authorization.RoleWorker, permissionToRole("worker"))
	assert.Equal(t, authorization.RoleAdmin, permissionToRole("admin"))
	assert.Equal(t, authorization.RoleUndefined, permissionToRole("unknown"))
}

func TestParseApiKeysString_Success(t *testing.T) {
	keys, err := parseAPIKeysString("app1:write:ns1; admin:*:should-not-parse; admin:admin:* ; worker:worker:ns2 ;  ")
	require.NoError(t, err)

	// app1 key -> namespace role
	c1 := keys["app1"]
	require.NotNil(t, c1)
	assert.Equal(t, "app1", c1.Subject)
	assert.Equal(t, authorization.RoleWriter, c1.Namespaces["ns1"])
	assert.Equal(t, authorization.RoleUndefined, c1.System)

	// admin key -> system role
	c2 := keys["admin"]
	require.NotNil(t, c2)
	assert.Equal(t, authorization.RoleAdmin, c2.System)
	assert.Empty(t, c2.Namespaces)

	// worker key -> namespace role
	c3 := keys["worker"]
	require.NotNil(t, c3)
	assert.Equal(t, authorization.RoleWorker, c3.Namespaces["ns2"])
}

func TestParseApiKeysString_Invalid(t *testing.T) {
	// wrong parts
	_, err := parseAPIKeysString("bad:format")
	require.Error(t, err)

	// empty sections are skipped, but malformed entries error
	_, err = parseAPIKeysString("ok:read:ns; badentry")
	require.Error(t, err)
}

func TestAPIKeyClaimMapper_GetClaims_TokenVariants(t *testing.T) {
	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper("valid:write:ns;admin:admin:*", logger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		authToken     string
		expectSubject string
		validate      func(*testing.T, *authorization.Claims)
	}{
		{"raw token", "valid", "valid", func(t *testing.T, c *authorization.Claims) {
			assert.Equal(t, authorization.RoleWriter, c.Namespaces["ns"])
		}},
		{"bearer lower", "bearer valid", "valid", func(t *testing.T, c *authorization.Claims) {
			assert.Equal(t, authorization.RoleWriter, c.Namespaces["ns"])
		}},
		{"bearer upper", "Bearer   valid", "valid", func(t *testing.T, c *authorization.Claims) {
			assert.Equal(t, authorization.RoleWriter, c.Namespaces["ns"])
		}},
		{"admin wildcard", "Bearer admin", "admin", func(t *testing.T, c *authorization.Claims) {
			assert.Equal(t, authorization.RoleAdmin, c.System)
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := mapper.GetClaims(&authorization.AuthInfo{AuthToken: tc.authToken})
			require.NoError(t, err)
			require.NotNil(t, claims)
			assert.Equal(t, tc.expectSubject, claims.Subject)
			if tc.validate != nil {
				tc.validate(t, claims)
			}
		})
	}
}

func TestAPIKeyClaimMapper_GetClaims_Nones(t *testing.T) {
	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper("k:read:ns", logger)
	require.NoError(t, err)

	// nil auth info
	claims, err := mapper.GetClaims(nil)
	require.NoError(t, err)
	assert.Nil(t, claims)

	// empty token
	claims, err = mapper.GetClaims(&authorization.AuthInfo{AuthToken: "   "})
	require.NoError(t, err)
	assert.Nil(t, claims)

	// non-bearer scheme should be ignored (let other mappers handle)
	claims, err = mapper.GetClaims(&authorization.AuthInfo{AuthToken: "Basic abc"})
	require.NoError(t, err)
	assert.Nil(t, claims)

	// unknown api key
	claims, err = mapper.GetClaims(&authorization.AuthInfo{AuthToken: "Bearer nope"})
	require.NoError(t, err)
	assert.Nil(t, claims)
}
