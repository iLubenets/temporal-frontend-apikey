package authorizer

import (
	"fmt"
	"strings"

	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
)

const (
	authorizationBearer = "bearer"
	permissionRead      = "read"
	permissionWrite     = "write"
	permissionWorker    = "worker"
	permissionAdmin     = "admin"
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
	if authInfo == nil {
		return nil, nil
	}
	raw := strings.TrimSpace(authInfo.AuthToken)
	if raw == "" {
		return nil, nil
	}

	token := raw
	if idx := strings.IndexByte(raw, ' '); idx > 0 {
		scheme := strings.TrimSpace(raw[:idx])
		rest := strings.TrimSpace(raw[idx+1:])
		if strings.EqualFold(scheme, authorizationBearer) {
			token = rest
		} else {
			// Not an API-key scheme; skip so other mappers may handle
			return nil, nil
		}
	}

	if claims, ok := m.keys[token]; ok {
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

func permissionToRole(permission string) authorization.Role {
	switch strings.ToLower(permission) {
	case permissionRead:
		return authorization.RoleReader
	case permissionWrite:
		return authorization.RoleWriter
	case permissionAdmin:
		return authorization.RoleAdmin
	case permissionWorker:
		return authorization.RoleWorker
	}
	return authorization.RoleUndefined
}
