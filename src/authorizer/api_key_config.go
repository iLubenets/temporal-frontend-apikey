package authorizer

import (
	"fmt"
	"os"
	"strings"

	"go.temporal.io/server/common/authorization"
)

type APIKeyConfig struct {
	keys map[string]*authorization.Claims
}

func NewApiKeyConfig() *APIKeyConfig {
	return &APIKeyConfig{
		keys: make(map[string]*authorization.Claims),
	}
}

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

func (c *APIKeyConfig) GetByApiKey(apiKey string) *authorization.Claims {
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
