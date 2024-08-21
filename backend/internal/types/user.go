package types

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type UserRole string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"

	UserRoleAdmin     UserRole = "admin"
	UserRoleUser      UserRole = "user"
	UserRoleDeveloper UserRole = "developer"
)

type User struct {
	UUID      uuid.UUID  `json:"uuid"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Status    UserStatus `json:"status"`
	Roles     []UserRole `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

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
