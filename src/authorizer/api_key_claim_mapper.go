package authorizer

import (
	"fmt"
	"strings"

	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
)

// APIKeyClaimMapper implements authorization.ClaimMapper only
type APIKeyClaimMapper struct {
	logger logpkg.Logger
	cfg    *APIKeyConfig
}

func NewAPIKeyClaimMapper(logger logpkg.Logger) (*APIKeyClaimMapper, error) {
	apiKeyCfg := NewApiKeyConfig()
	if err := apiKeyCfg.LoadEnv(); err != nil {
		return nil, err
	}
	logger.Info("API key claim-mapper initialized")
	return &APIKeyClaimMapper{logger: logger, cfg: apiKeyCfg}, nil
}

// GetClaims extracts API key from Authorization header and maps to Claims.
func (m *APIKeyClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	if authInfo == nil {
		return nil, fmt.Errorf("missing auth info")
	}
	apiKey := strings.TrimSpace(authInfo.AuthToken)
	if strings.HasPrefix(strings.ToLower(apiKey), "bearer ") {
		apiKey = strings.TrimSpace(apiKey[len("bearer "):])
	}
	if apiKey == "" {
		// No token => no claims. Default authorizer will deny (except health checks).
		return nil, nil
	}

	claims := m.cfg.GetByApiKey(apiKey)
	if claims == nil {
		return nil, nil
	}

	return claims, nil
}
