package authorizer

import (
	"fmt"
	"os"
	"strings"

	"go.temporal.io/server/common/authorization"
)

// APIKeyConfig manages API key to claims mappings loaded from environment variables.
type APIKeyConfig struct {
	keys map[string]*authorization.Claims
}

// NewAPIKeyConfig creates a new APIKeyConfig instance with an empty key map.
func NewAPIKeyConfig() *APIKeyConfig {
	return &APIKeyConfig{
		keys: make(map[string]*authorization.Claims),
	}
}

// LoadEnv loads API keys from TEMPORAL_API_KEYS environment variable.
// Expected format: "key:role:namespace;key2:role2:namespace2" where role is reader/writer/worker/admin and namespace can be * for system-wide access.
func (c *APIKeyConfig) LoadEnv() error {
	// app1-key:writer:app1-namespace
	apiKeysEnv := os.Getenv("TEMPORAL_API_KEYS")

	for _, entry := range strings.Split(apiKeysEnv, ";") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 {
			return fmt.Errorf("invalid key [%.*s...] format - expected <key>:<role>:<namespace>", 3, entry)
		}
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return fmt.Errorf("invalid key format: [<key>(len:%d):<role>(val:%s):<namespace>(val:%s)]", len(parts[0]), parts[1], parts[2])
		}

		apiKey := parts[0]
		role := roleFromString(parts[1])
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

		c.keys[apiKey] = claim
	}
	if len(c.keys) == 0 {
		return fmt.Errorf("no valid keys")
	}

	return nil
}

// GetByAPIKey retrieves authorization claims for the given API key, returns nil if not found.
func (c *APIKeyConfig) GetByAPIKey(apiKey string) *authorization.Claims {
	if claim, ok := c.keys[apiKey]; ok {
		return claim
	}
	return nil
}

func roleFromString(s string) authorization.Role {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "worker":
		return authorization.RoleWorker
	case "reader":
		return authorization.RoleReader
	case "writer":
		return authorization.RoleWriter
	case "admin":
		return authorization.RoleAdmin
	default:
		return authorization.RoleUndefined
	}
}
