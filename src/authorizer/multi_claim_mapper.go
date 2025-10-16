package authorizer

import (
	"fmt"

	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// MultiClaimMapper enable multiple claim mappers at the same time
type MultiClaimMapper struct {
	logger       logpkg.Logger
	claimMappers map[string]authorization.ClaimMapper
}

// NewMultiClaimMapper creates a new MultiClaimMapper
func NewMultiClaimMapper(logger logpkg.Logger) *MultiClaimMapper {
	return &MultiClaimMapper{logger: logger, claimMappers: map[string]authorization.ClaimMapper{}}
}

// Add new claim mapper to the chain
func (m *MultiClaimMapper) Add(claimMapperName string, claimMapper authorization.ClaimMapper) {
	m.claimMappers[claimMapperName] = claimMapper
	m.logger.Info("adding claim mapper", tag.Name(claimMapperName))
}

// GetClaims converts authorization info of a subject into Temporal claims (permissions) for authorization
func (m *MultiClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	for name, cm := range m.claimMappers {
		claims, err := cm.GetClaims(authInfo)
		if err != nil {
			m.logger.Warn(fmt.Sprintf("failed to get claims authInfo:%v", authInfo), tag.Name(name), tag.Error(err))
			return claims, err
		}
		if claims == nil || (claims.System == authorization.RoleUndefined && len(claims.Namespaces) == 0) {
			m.logger.Info("claim mapper didn't recognize the token", tag.Name(name))
			continue
		}
		m.logger.Info("claim mapper identified the token ", tag.Name(name))
		return claims, nil
	}
	m.logger.Info("non of claim mappers recognized the token")
	return &authorization.Claims{}, nil
}
