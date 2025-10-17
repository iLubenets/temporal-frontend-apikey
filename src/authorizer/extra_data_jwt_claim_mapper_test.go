package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/log"
)

type mockClaimMapper struct {
	claims *authorization.Claims
	err    error
}

func (m *mockClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	return m.claims, m.err
}

func TestNewExtraDataJWTClamMapper(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{}

	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	require.NotNil(t, mapper)
	assert.IsType(t, &extraDataJWTClamMapper{}, mapper)
}

func TestExtraDataJWTClamMapper_GetClaims_NilAuthInfo(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{Subject: "test"},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	claims, err := mapper.GetClaims(nil)

	require.NoError(t, err)
	assert.Nil(t, claims)
}

func TestExtraDataJWTClamMapper_GetClaims_NoExtraData(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{Subject: "test"},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer some-token",
		ExtraData: "",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	assert.Nil(t, claims)
}

func TestExtraDataJWTClamMapper_GetClaims_ExtraDataNotJWT(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{Subject: "test"},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer some-token",
		ExtraData: "not-a-jwt", // No dots, doesn't look like JWT
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	assert.Nil(t, claims)
}

func TestExtraDataJWTClamMapper_GetClaims_WithJWTInExtraData(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{
			Subject: "user123",
			System:  authorization.RoleReader,
		},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user123", claims.Subject)
	assert.Equal(t, authorization.RoleReader, claims.System)
}

func TestExtraDataJWTClamMapper_GetClaims_WithBearerPrefixInExtraData(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{
			Subject: "user456",
			System:  authorization.RoleWriter,
		},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user456", claims.Subject)
}

func TestExtraDataJWTClamMapper_GetClaims_WithBearerLowercase(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{
			Subject: "user789",
		},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user789", claims.Subject)
}

func TestExtraDataJWTClamMapper_GetClaims_NoWorkaroundWhenHasNamespaces(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{
			Subject:    "user-with-ns",
			System:     authorization.RoleUndefined,
			Namespaces: map[string]authorization.Role{"ns1": authorization.RoleReader},
		},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "user-with-ns", claims.Subject)
	// Should NOT apply workaround since has namespaces
	assert.Equal(t, authorization.RoleUndefined, claims.System)
}

func TestExtraDataJWTClamMapper_GetClaims_NoWorkaroundWhenHasSystemRole(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		claims: &authorization.Claims{
			Subject:    "admin-user",
			System:     authorization.RoleWriter,
			Namespaces: map[string]authorization.Role{},
		},
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.NoError(t, err)
	require.NotNil(t, claims)
	// Should NOT apply workaround since already has system role
	assert.Equal(t, authorization.RoleWriter, claims.System)
}

func TestExtraDataJWTClamMapper_GetClaims_PropagatesError(t *testing.T) {
	logger := log.NewTestLogger()
	mockMapper := &mockClaimMapper{
		err: assert.AnError,
	}
	mapper := NewExtraDataJWTClamMapper(mockMapper, logger)

	authInfo := &authorization.AuthInfo{
		AuthToken: "Bearer original-token",
		ExtraData: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	claims, err := mapper.GetClaims(authInfo)

	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestLooksLikeJWT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid JWT format",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expected: true,
		},
		{
			name:     "simple three-part token",
			input:    "part1.part2.part3",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "single part",
			input:    "singlepart",
			expected: false,
		},
		{
			name:     "two parts",
			input:    "part1.part2",
			expected: false,
		},
		{
			name:     "four parts",
			input:    "part1.part2.part3.part4",
			expected: false,
		},
		{
			name:     "just dots",
			input:    "..",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := looksLikeJWT(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
