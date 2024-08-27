package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type UserRole string
type ContextKey string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"

	UserRoleAdmin     UserRole = "admin"
	UserRoleUser      UserRole = "user"
	UserRoleDeveloper UserRole = "developer"

	AuthenticatedUserCtxKey            = "authenticated_user"
	UserDetailsCtxKey       ContextKey = "user_details"
)

type UserFilter struct {
	NameContains  string     `json:"name"`
	EmailContains string     `json:"email"`
	Status        UserStatus `json:"status"`
	Roles         []UserRole `json:"roles"`
}

type UserFilterOption struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type PaginatedUsers struct {
	Users      []User `json:"users"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

type User struct {
	UUID      uuid.UUID  `json:"uuid"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status"`
	Roles     []UserRole `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

var anonymousUserData = User{
	UUID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	Name:      "Anonymous User",
	Email:     "user@example.com",
	Status:    UserStatusActive,
	Roles:     []UserRole{UserRoleAdmin},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

// DefaultAnonymousUser returns a pointer to a copy of the anonymous User
func DefaultAnonymousUser() *User {
	userCopy := anonymousUserData
	return &userCopy
}

// GetAuthenticatedUser safely retrieves the authenticated user from the context
// If the user is not found or is nil, it returns the anonymous user
func GetAuthenticatedUser(ctx context.Context) *User {
	if ctx == nil {
		return DefaultAnonymousUser()
	}

	value := ctx.Value(AuthenticatedUserCtxKey)
	if value == nil {
		return DefaultAnonymousUser()
	}

	user, ok := value.(*User)
	if !ok || user == nil {
		return DefaultAnonymousUser()
	}

	return user
}
