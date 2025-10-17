package authorizer

import (
	"strings"

	"go.temporal.io/server/common/authorization"
	logpkg "go.temporal.io/server/common/log"
)

type extraDataJWTClamMapper struct {
	defaultJWTClaimMapper authorization.ClaimMapper
	logger                logpkg.Logger
}

// NewExtraDataJWTClamMapper using defaultJWTClaimMapper to check AuthInfo.ExtraData
func NewExtraDataJWTClamMapper(defaultJWTClaimMapper authorization.ClaimMapper, logger logpkg.Logger) authorization.ClaimMapper {
	return &extraDataJWTClamMapper{defaultJWTClaimMapper: defaultJWTClaimMapper, logger: logger}
}

// GetClaims check "Authorization-Extras" header if present
func (m *extraDataJWTClamMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	if authInfo != nil && looksLikeJWT(authInfo.ExtraData) {
		authToken := authInfo.AuthToken
		extraData := authInfo.ExtraData
		// switch AuthToken<->ExtraData
		alt := *authInfo
		if strings.HasPrefix(strings.ToLower(extraData), authorizationBearer) {
			alt.AuthToken = extraData
		} else {
			alt.AuthToken = authorizationBearer + " " + extraData
		}
		alt.ExtraData = authToken
		claims, err := m.defaultJWTClaimMapper.GetClaims(&alt)
		return claims, err
	}
	return nil, nil
}

func looksLikeJWT(t string) bool {
	// very light check: three dot-separated parts
	return strings.Count(t, ".") == 2
}
