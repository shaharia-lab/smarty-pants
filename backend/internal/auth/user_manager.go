package auth

import (
	"context"

	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
)

type UserManager struct {
	storage storage.Storage
}

func NewUserManager(storage storage.Storage) *UserManager {
	return &UserManager{storage: storage}
}

func (um *UserManager) CreateUser(ctx context.Context, name, email, status string) (*types.User, error) {
	user := &types.User{
		Name:   name,
		Email:  email,
		Status: status,
	}

	err := um.storage.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (um *UserManager) GetUser(ctx context.Context, uuid string) (*types.User, error) {
	return um.storage.GetUser(ctx, uuid)
}

func (um *UserManager) UpdateUserStatus(ctx context.Context, uuid string, status string) error {
	return um.storage.UpdateUserStatus(ctx, uuid, status)
}

func (um *UserManager) ActivateUser(ctx context.Context, uuid string) error {
	return um.UpdateUserStatus(ctx, uuid, "active")
}

func (um *UserManager) DeactivateUser(ctx context.Context, uuid string) error {
	return um.UpdateUserStatus(ctx, uuid, "inactive")
}
