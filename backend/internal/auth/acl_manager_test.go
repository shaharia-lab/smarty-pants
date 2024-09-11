package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestACLManager_IsAllowed(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)

	tests := []struct {
		name         string
		enableAuth   bool
		userRoles    []types.UserRole
		requiredRole types.UserRole
		expected     bool
	}{
		{"Auth disabled", false, []types.UserRole{types.UserRoleUser}, types.UserRoleAdmin, true},
		{"Admin accessing admin resource", true, []types.UserRole{types.UserRoleAdmin}, types.UserRoleAdmin, true},
		{"Admin accessing user resource", true, []types.UserRole{types.UserRoleAdmin}, types.UserRoleUser, true},
		{"Admin accessing developer resource", true, []types.UserRole{types.UserRoleAdmin}, types.UserRoleDeveloper, true},
		{"Developer accessing developer resource", true, []types.UserRole{types.UserRoleDeveloper}, types.UserRoleDeveloper, true},
		{"Developer accessing user resource", true, []types.UserRole{types.UserRoleDeveloper}, types.UserRoleUser, true},
		{"Developer accessing admin resource", true, []types.UserRole{types.UserRoleDeveloper}, types.UserRoleAdmin, false},
		{"User accessing user resource", true, []types.UserRole{types.UserRoleUser}, types.UserRoleUser, true},
		{"User accessing developer resource", true, []types.UserRole{types.UserRoleUser}, types.UserRoleDeveloper, false},
		{"User accessing admin resource", true, []types.UserRole{types.UserRoleUser}, types.UserRoleAdmin, false},
		{"Multiple roles: Admin and User accessing admin resource", true, []types.UserRole{types.UserRoleAdmin, types.UserRoleUser}, types.UserRoleAdmin, true},
		{"Multiple roles: Developer and User accessing developer resource", true, []types.UserRole{types.UserRoleDeveloper, types.UserRoleUser}, types.UserRoleDeveloper, true},
		{"No user in context", true, nil, types.UserRoleUser, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aclManager := NewACLManager(logger, tt.enableAuth)

			// Create a mock request with a context
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.userRoles != nil {
				ctx := context.WithValue(req.Context(), types.AuthenticatedUserCtxKey, &types.User{Roles: tt.userRoles})
				req = req.WithContext(ctx)
			}

			// Create a mock response writer
			w := httptest.NewRecorder()

			// Call IsAllowed
			result := aclManager.IsAllowed(w, req, tt.requiredRole, "", "test-resource")

			// Assert the result
			assert.Equal(t, tt.expected, result, "IsAllowed returned unexpected result")

			// Check response status code for unauthorized/forbidden cases
			if !tt.expected && tt.enableAuth {
				if tt.userRoles == nil {
					assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected unauthorized status code")
				} else {
					assert.Equal(t, http.StatusForbidden, w.Code, "Expected forbidden status code")
				}
			}
		})
	}
}

func TestACLManager_getUserFromContext(t *testing.T) {
	logger := logrus.New()
	aclManager := NewACLManager(logger, true)

	tests := []struct {
		name        string
		contextUser *types.User
		expectError bool
	}{
		{"User in context", &types.User{Roles: []types.UserRole{types.UserRoleUser}}, false},
		{"No user in context", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.contextUser != nil {
				ctx = context.WithValue(ctx, types.AuthenticatedUserCtxKey, tt.contextUser)
			}

			user, err := aclManager.GetAuthenticatedUserFromContext(ctx)

			if tt.expectError {
				assert.Error(t, err, "Expected an error")
				assert.Nil(t, user, "User should be nil")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.contextUser, user, "Retrieved user doesn't match")
			}
		})
	}
}

func TestACLManager_hasRoleOrHigher(t *testing.T) {
	logger := logrus.New()
	aclManager := NewACLManager(logger, true)

	tests := []struct {
		name         string
		userRoles    []types.UserRole
		requiredRole types.UserRole
		expected     bool
	}{
		{"Admin has admin role", []types.UserRole{types.UserRoleAdmin}, types.UserRoleAdmin, true},
		{"Admin has user role", []types.UserRole{types.UserRoleAdmin}, types.UserRoleUser, true},
		{"Admin has developer role", []types.UserRole{types.UserRoleAdmin}, types.UserRoleDeveloper, true},
		{"Developer has developer role", []types.UserRole{types.UserRoleDeveloper}, types.UserRoleDeveloper, true},
		{"Developer has user role", []types.UserRole{types.UserRoleDeveloper}, types.UserRoleUser, true},
		{"Developer doesn't have admin role", []types.UserRole{types.UserRoleDeveloper}, types.UserRoleAdmin, false},
		{"User has user role", []types.UserRole{types.UserRoleUser}, types.UserRoleUser, true},
		{"User doesn't have developer role", []types.UserRole{types.UserRoleUser}, types.UserRoleDeveloper, false},
		{"User doesn't have admin role", []types.UserRole{types.UserRoleUser}, types.UserRoleAdmin, false},
		{"Multiple roles: Admin and User", []types.UserRole{types.UserRoleAdmin, types.UserRoleUser}, types.UserRoleAdmin, true},
		{"Multiple roles: Developer and User", []types.UserRole{types.UserRoleDeveloper, types.UserRoleUser}, types.UserRoleDeveloper, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aclManager.hasRoleOrHigher(tt.userRoles, tt.requiredRole)
			assert.Equal(t, tt.expected, result, "hasRoleOrHigher returned unexpected result")
		})
	}
}
