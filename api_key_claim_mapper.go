package main

import (
	"fmt"
	"strings"

	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// APIKey holds key info
type APIKey struct {
	Key       string
	Role      string
	Namespace string // "*" = all namespaces
}

// APIKeyClaimMapper implements authorization.ClaimMapper only
type APIKeyClaimMapper struct {
	logger logpkg.Logger
	keys   map[string]*APIKey
}

func NewAPIKeyClaimMapper(apiKeys string, logger logpkg.Logger) (*APIKeyClaimMapper, error) {
	keys := make(map[string]*APIKey)
	for _, entry := range strings.Split(apiKeys, ";") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 || parts[0] == "" {
			return nil, fmt.Errorf("invalid key format: %s", entry)
		}
		keys[parts[0]] = &APIKey{Key: parts[0], Role: parts[1], Namespace: parts[2]}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no valid keys")
	}

	logger.Info("API key claim-mapper initialized", tag.NewInt("keys", len(keys)))
	return &APIKeyClaimMapper{logger: logger, keys: keys}, nil
}

// GetClaims extracts API key from Authorization header and maps to Claims.
func (m *APIKeyClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	if authInfo == nil {
		return nil, fmt.Errorf("missing auth info")
	}
	token := strings.TrimSpace(authInfo.AuthToken)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	if token == "" {
		// No token => no claims. Default authorizer will deny (except health checks).
		return nil, nil
	}
	keyInfo, ok := m.keys[token]
	if !ok {
		return nil, fmt.Errorf("invalid key")
	}
	// Map role
	role := roleFromString(keyInfo.Role)
	claims := &authorization.Claims{Subject: keyInfo.Key, Namespaces: map[string]authorization.Role{}}
	if keyInfo.Namespace == "*" {
		claims.System = role
	} else {
		claims.Namespaces[keyInfo.Namespace] = role
	}
	return claims, nil
}

func roleFromString(s string) authorization.Role {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "worker":
		return authorization.RoleWorker
	case "reader", "read":
		return authorization.RoleReader
	case "writer", "write":
		return authorization.RoleWriter
	case "admin":
		return authorization.RoleAdmin
	default:
		return authorization.RoleUndefined
	}
}
