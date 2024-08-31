package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

var roleHierarchy = map[types.UserRole][]types.UserRole{
	types.UserRoleAdmin:     {types.UserRoleUser, types.UserRoleDeveloper},
	types.UserRoleUser:      {},
	types.UserRoleDeveloper: {types.UserRoleUser},
}

type ACLManager struct {
	logger     *logrus.Logger
	enableAuth bool
}

func NewACLManager(logger *logrus.Logger, enableAuth bool) ACLManager {
	return ACLManager{
		logger:     logger,
		enableAuth: enableAuth,
	}
}

// IsAllowed checks if the user has the required role to access the resource.
func (m *ACLManager) IsAllowed(w http.ResponseWriter, r *http.Request, requiredRole types.UserRole, operationType string, _ interface{}) bool {
	if !m.enableAuth {
		return true
	}

	user, err := m.getUserFromContext(r.Context())
	if err != nil {
		m.logger.WithError(err).Error("Failed to get user role from context")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Un-Authorized", Err: "No authenticated user"})
		return false
	}

	if m.hasRoleOrHigher(user.Roles, requiredRole) {
		return true
	}

	m.logger.WithFields(logrus.Fields{
		"user_roles":    user.Roles,
		"required_role": requiredRole,
		"resource":      operationType,
	}).Warn("Access denied due to insufficient role")

	util.SendAPIErrorResponse(
		w,
		http.StatusForbidden,
		&util.APIError{
			Message: "You are not authorized to access this resource",
			Err:     fmt.Sprintf("Only %s is allowed to access this resource", requiredRole),
		},
	)
	return false
}

// getUserFromContext retrieves the user from the context.
func (m *ACLManager) getUserFromContext(ctx context.Context) (*types.User, error) {
	user, ok := ctx.Value(types.AuthenticatedUserCtxKey).(*types.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// hasRoleOrHigher checks if any of the user's roles are equal to or higher than the required role.
func (m *ACLManager) hasRoleOrHigher(userRoles []types.UserRole, requiredRole types.UserRole) bool {
	for _, userRole := range userRoles {
		if userRole == requiredRole || m.isHigherRole(userRole, requiredRole) {
			return true
		}
	}
	return false
}

// isHigherRole checks if a given user role is higher or equal in the hierarchy compared to the required role.
func (m *ACLManager) isHigherRole(userRole, requiredRole types.UserRole) bool {
	higherRoles, exists := roleHierarchy[userRole]
	if !exists {
		return false
	}

	for _, role := range higherRoles {
		if role == requiredRole {
			return true
		}
		if m.isHigherRole(role, requiredRole) {
			return true
		}
	}

	return false
}
