package authorizer

import (
	"strings"

	"go.temporal.io/server/common/authorization"
)

const (
	authorizationBearer = "bearer"

	permissionRead   = "read"
	permissionWrite  = "write"
	permissionWorker = "worker"
	permissionAdmin  = "admin"
)

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

func hasClaims(c *authorization.Claims) bool {
	return c != nil && (c.System != authorization.RoleUndefined || len(c.Namespaces) > 0)
}
