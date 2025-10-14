package authorizer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
)

func TestNewApiKeyConfig(t *testing.T) {
	cfg := NewApiKeyConfig()
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.keys)
	assert.Equal(t, 0, len(cfg.keys))
}

func TestRoleFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected authorization.Role
	}{
		{"worker", "worker", authorization.RoleWorker},
		{"Worker uppercase", "Worker", authorization.RoleWorker},
		{"WORKER uppercase", "WORKER", authorization.RoleWorker},
		{"worker with spaces", "  worker  ", authorization.RoleWorker},
		{"reader", "reader", authorization.RoleReader},
		{"Reader uppercase", "Reader", authorization.RoleReader},
		{"writer", "writer", authorization.RoleWriter},
		{"Writer uppercase", "Writer", authorization.RoleWriter},
		{"admin", "admin", authorization.RoleAdmin},
		{"Admin uppercase", "Admin", authorization.RoleAdmin},
		{"ADMIN uppercase", "ADMIN", authorization.RoleAdmin},
		{"undefined empty", "", authorization.RoleUndefined},
		{"undefined invalid", "invalid", authorization.RoleUndefined},
		{"undefined unknown", "unknown-role", authorization.RoleUndefined},
		{"read should be undefined", "read", authorization.RoleUndefined},
		{"write should be undefined", "write", authorization.RoleUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIKeyConfig_LoadEnv_Success(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedKeys   int
		validateClaims func(t *testing.T, cfg *APIKeyConfig)
	}{
		{
			name:         "single key with namespace",
			envValue:     "app1-key:writer:app1-namespace",
			expectedKeys: 1,
			validateClaims: func(t *testing.T, cfg *APIKeyConfig) {
				claims := cfg.GetByApiKey("app1-key")
				require.NotNil(t, claims)
				assert.Equal(t, "app1-key", claims.Subject)
				assert.Equal(t, authorization.RoleWriter, claims.Namespaces["app1-namespace"])
				assert.Equal(t, authorization.RoleUndefined, claims.System)
			},
		},
		{
			name:         "single key with wildcard namespace",
			envValue:     "admin-key:admin:*",
			expectedKeys: 1,
			validateClaims: func(t *testing.T, cfg *APIKeyConfig) {
				claims := cfg.GetByApiKey("admin-key")
				require.NotNil(t, claims)
				assert.Equal(t, "admin-key", claims.Subject)
				assert.Equal(t, authorization.RoleAdmin, claims.System)
				assert.Equal(t, 0, len(claims.Namespaces))
			},
		},
		{
			name:         "multiple keys",
			envValue:     "key1:reader:ns1;key2:writer:ns2;key3:admin:*",
			expectedKeys: 3,
			validateClaims: func(t *testing.T, cfg *APIKeyConfig) {
				claims1 := cfg.GetByApiKey("key1")
				require.NotNil(t, claims1)
				assert.Equal(t, authorization.RoleReader, claims1.Namespaces["ns1"])

				claims2 := cfg.GetByApiKey("key2")
				require.NotNil(t, claims2)
				assert.Equal(t, authorization.RoleWriter, claims2.Namespaces["ns2"])

				claims3 := cfg.GetByApiKey("key3")
				require.NotNil(t, claims3)
				assert.Equal(t, authorization.RoleAdmin, claims3.System)
			},
		},
		{
			name:         "key with worker role",
			envValue:     "worker-key:worker:my-namespace",
			expectedKeys: 1,
			validateClaims: func(t *testing.T, cfg *APIKeyConfig) {
				claims := cfg.GetByApiKey("worker-key")
				require.NotNil(t, claims)
				assert.Equal(t, authorization.RoleWorker, claims.Namespaces["my-namespace"])
			},
		},
		{
			name:         "keys with whitespace",
			envValue:     "  key1:reader:ns1  ;  key2:writer:ns2  ",
			expectedKeys: 2,
			validateClaims: func(t *testing.T, cfg *APIKeyConfig) {
				claims1 := cfg.GetByApiKey("key1")
				require.NotNil(t, claims1)

				claims2 := cfg.GetByApiKey("key2")
				require.NotNil(t, claims2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEMPORAL_API_KEYS", tt.envValue)
			defer os.Unsetenv("TEMPORAL_API_KEYS")

			cfg := NewApiKeyConfig()
			err := cfg.LoadEnv()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedKeys, len(cfg.keys))

			if tt.validateClaims != nil {
				tt.validateClaims(t, cfg)
			}
		})
	}
}

func TestAPIKeyConfig_LoadEnv_Error(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedError string
	}{
		{
			name:          "empty env",
			envValue:      "",
			expectedError: "invalid key",
		},
		{
			name:          "invalid format - too few parts",
			envValue:      "key:role",
			expectedError: "invalid key",
		},
		{
			name:          "invalid format - too many parts",
			envValue:      "key:role:namespace:extra",
			expectedError: "invalid key",
		},
		{
			name:          "empty key",
			envValue:      ":role:namespace",
			expectedError: "invalid key format",
		},
		{
			name:          "empty role",
			envValue:      "key::namespace",
			expectedError: "invalid key format",
		},
		{
			name:          "empty namespace",
			envValue:      "key:role:",
			expectedError: "invalid key format",
		},
		{
			name:          "multiple entries with one invalid",
			envValue:      "key1:role:namespace;invalid",
			expectedError: "invalid key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEMPORAL_API_KEYS", tt.envValue)
			defer os.Unsetenv("TEMPORAL_API_KEYS")

			cfg := NewApiKeyConfig()
			err := cfg.LoadEnv()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestAPIKeyConfig_GetByApiKey(t *testing.T) {
	os.Setenv("TEMPORAL_API_KEYS", "key1:reader:ns1;key2:writer:ns2")
	defer os.Unsetenv("TEMPORAL_API_KEYS")

	cfg := NewApiKeyConfig()
	err := cfg.LoadEnv()
	require.NoError(t, err)

	t.Run("existing key", func(t *testing.T) {
		claims := cfg.GetByApiKey("key1")
		assert.NotNil(t, claims)
		assert.Equal(t, "key1", claims.Subject)
	})

	t.Run("non-existing key", func(t *testing.T) {
		claims := cfg.GetByApiKey("non-existing-key")
		assert.Nil(t, claims)
	})

	t.Run("empty key", func(t *testing.T) {
		claims := cfg.GetByApiKey("")
		assert.Nil(t, claims)
	})
}
