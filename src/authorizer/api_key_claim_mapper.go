package authorizer

import (
	"fmt"
	"strings"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
)

// apiKeyClaimMapper implements authorization.ClaimMapper only
type apiKeyClaimMapper struct {
	logger logpkg.Logger
	keys   map[string]*authorization.Claims
}

// NewAPIKeyClaimMapper creates a new apiKeyClaimMapper with the given logger and loads API key configuration from environment.
func NewAPIKeyClaimMapper(apiKeysString string, logger logpkg.Logger) (authorization.ClaimMapper, error) {
	keys, err := parseAPIKeysString(apiKeysString)
	if err != nil {
		return nil, err
	}
	logger.Info("API key claim-mapper initialized")
	return &apiKeyClaimMapper{logger: logger, keys: keys}, nil
}

// GetClaims extracts API key from Authorization header and maps to Claims.
func (m *apiKeyClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	if authInfo == nil || authInfo.AuthToken == "" {
		return nil, nil
	}
	parts := strings.SplitN(authInfo.AuthToken, " ", 2)
	if len(parts) != 2 {
		return nil, serviceerror.NewPermissionDenied("unexpected authorization token format", "")
	}
	if !strings.EqualFold(parts[0], authorizationBearer) {
		return nil, serviceerror.NewPermissionDenied("unexpected name in authorization token", "")
	}
	key := strings.TrimSpace(parts[1])

	if claims, ok := m.keys[key]; ok {
		return claims, nil
	}
	return nil, nil
}

func parseAPIKeysString(apiKeysStr string) (map[string]*authorization.Claims, error) {
	keys := make(map[string]*authorization.Claims)
	for _, entry := range strings.Split(apiKeysStr, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 {
			return keys, fmt.Errorf("invalid key [%.*s...] format - expected <key>:<role>:<namespace>", 3, entry)
		}
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return keys, fmt.Errorf("invalid key format: [<key>(len:%d):<role>(val:%s):<namespace>(val:%s)]", len(parts[0]), parts[1], parts[2])
		}

		apiKey := parts[0]
		role := permissionToRole(parts[1])
		namespace := parts[2]
		claim := &authorization.Claims{
			Subject:    apiKey,
			Namespaces: map[string]authorization.Role{},
		}
		if namespace == "*" {
			claim.System = role
		} else {
			claim.Namespaces[namespace] = role
		}

		keys[apiKey] = claim
	}

	return keys, nil
}
