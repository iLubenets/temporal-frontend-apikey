package authorizer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/log"
)

func TestNewAPIKeyClaimMapper_Success(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "test-key:reader:test-namespace")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)

	require.NoError(t, err)
	assert.NotNil(t, mapper)
	assert.NotNil(t, mapper.logger)
	assert.NotNil(t, mapper.cfg)
}

func TestNewAPIKeyClaimMapper_Error(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)

	require.Error(t, err)
	assert.Nil(t, mapper)
	assert.Contains(t, err.Error(), "invalid key")
}

func TestAPIKeyClaimMapper_GetClaims_NilAuthInfo(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "test-key:reader:test-namespace")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)
	require.NoError(t, err)

	claims, err := mapper.GetClaims(nil)

	assert.Nil(t, claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing auth info")
}

func TestAPIKeyClaimMapper_GetClaims_EmptyToken(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "test-key:reader:test-namespace")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		authToken string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"bearer with no token", "Bearer "},
		{"bearer with whitespace", "Bearer    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authInfo := &authorization.AuthInfo{
				AuthToken: tt.authToken,
			}

			claims, err := mapper.GetClaims(authInfo)

			assert.Nil(t, claims)
			assert.NoError(t, err)
		})
	}
}

func TestAPIKeyClaimMapper_GetClaims_ValidKey(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "valid-key:writer:my-namespace;admin-key:admin:*")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)
	require.NoError(t, err)

	tests := []struct {
		name            string
		authToken       string
		expectedSubject string
		validateClaims  func(t *testing.T, claims *authorization.Claims)
	}{
		{
			name:            "direct key",
			authToken:       "valid-key",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
				assert.Equal(t, authorization.RoleUndefined, claims.System)
			},
		},
		{
			name:            "bearer token lowercase",
			authToken:       "bearer valid-key",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:            "bearer token uppercase",
			authToken:       "Bearer valid-key",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:            "bearer token mixed case",
			authToken:       "BeArEr valid-key",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:            "bearer token with extra spaces",
			authToken:       "Bearer   valid-key",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:            "token with leading/trailing spaces",
			authToken:       "  valid-key  ",
			expectedSubject: "valid-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:            "admin key with wildcard",
			authToken:       "admin-key",
			expectedSubject: "admin-key",
			validateClaims: func(t *testing.T, claims *authorization.Claims) {
				assert.Equal(t, authorization.RoleAdmin, claims.System)
				assert.Equal(t, 0, len(claims.Namespaces))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authInfo := &authorization.AuthInfo{
				AuthToken: tt.authToken,
			}

			claims, err := mapper.GetClaims(authInfo)

			require.NoError(t, err)
			require.NotNil(t, claims)
			assert.Equal(t, tt.expectedSubject, claims.Subject)

			if tt.validateClaims != nil {
				tt.validateClaims(t, claims)
			}
		})
	}
}

func TestAPIKeyClaimMapper_GetClaims_InvalidKey(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "valid-key:reader:test-namespace")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		authToken string
	}{
		{"non-existing key", "invalid-key"},
		{"bearer with invalid key", "Bearer invalid-key"},
		{"partial key", "valid"},
		{"key with suffix", "valid-key-extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authInfo := &authorization.AuthInfo{
				AuthToken: tt.authToken,
			}

			claims, err := mapper.GetClaims(authInfo)

			assert.Nil(t, claims)
			assert.NoError(t, err)
		})
	}
}

func TestAPIKeyClaimMapper_GetClaims_MultipleKeys(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "key1:reader:ns1;key2:writer:ns2;key3:worker:ns3")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	logger := log.NewTestLogger()
	mapper, err := NewAPIKeyClaimMapper(logger)
	require.NoError(t, err)

	// Test that each key returns its own claims
	authInfo1 := &authorization.AuthInfo{AuthToken: "key1"}
	claims1, err := mapper.GetClaims(authInfo1)
	require.NoError(t, err)
	require.NotNil(t, claims1)
	assert.Equal(t, "key1", claims1.Subject)
	assert.Equal(t, authorization.RoleReader, claims1.Namespaces["ns1"])

	authInfo2 := &authorization.AuthInfo{AuthToken: "key2"}
	claims2, err := mapper.GetClaims(authInfo2)
	require.NoError(t, err)
	require.NotNil(t, claims2)
	assert.Equal(t, "key2", claims2.Subject)
	assert.Equal(t, authorization.RoleWriter, claims2.Namespaces["ns2"])

	authInfo3 := &authorization.AuthInfo{AuthToken: "key3"}
	claims3, err := mapper.GetClaims(authInfo3)
	require.NoError(t, err)
	require.NotNil(t, claims3)
	assert.Equal(t, "key3", claims3.Subject)
	assert.Equal(t, authorization.RoleWorker, claims3.Namespaces["ns3"])
}
