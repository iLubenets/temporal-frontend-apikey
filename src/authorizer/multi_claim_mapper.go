package authorizer

import "go.temporal.io/server/common/authorization"

// MultiClaimMapper enable multiple claim mappers at the same time
type MultiClaimMapper struct {
	claimMappers []authorization.ClaimMapper
}

// NewMultiClaimMapper creates a new MultiClaimMapper
func NewMultiClaimMapper() *MultiClaimMapper {
	return &MultiClaimMapper{claimMappers: []authorization.ClaimMapper{}}
}

// Add new claim mapper to the chain
func (m *MultiClaimMapper) Add(claimMapper authorization.ClaimMapper) {
	m.claimMappers = append(m.claimMappers, claimMapper)
}

// GetClaims converts authorization info of a subject into Temporal claims (permissions) for authorization
func (m *MultiClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	for _, cm := range m.claimMappers {
		claims, err := cm.GetClaims(authInfo)
		if err != nil {
			return claims, err
		}
		if claims == nil || (claims.System == authorization.RoleUndefined && len(claims.Namespaces) == 0) {
			continue
		}
		return claims, nil
	}
	return &authorization.Claims{}, nil
}
