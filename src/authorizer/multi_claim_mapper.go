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
	m.logger.Info("auth: claim-mapper registered", tag.Name(claimMapperName))
}

// GetClaims converts authorization info of a subject into Temporal claims (permissions) for authorization
func (m *MultiClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	for name, cm := range m.claimMappers {
		claims, err := cm.GetClaims(authInfo)
		if err != nil {
			m.logger.Warn("auth: claim-mapper error", tag.Name(name), tag.Error(err))
			continue
		}
		if !hasClaims(claims) {
			m.logger.Debug("auth: claim-mapper skipped: no claims recognized", tag.Name(name))
			continue
		}
		m.logger.Info("auth: claim-mapper selected and permissions identified",
			tag.Name(name), tag.NewStringTag("claims", fmt.Sprintf("sys:%v,ns:%v", claims.System, claims.Namespaces)))
		return claims, nil
	}
	m.logger.Warn("auth: no claim-mapper recognized the credentials")
	return &authorization.Claims{}, nil
}
