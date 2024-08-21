package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

type contextKey string

const userContextKey contextKey = "user_details"

// UserManager is a manager for user operations.
type UserManager struct {
	storage storage.Storage
	logger  *logrus.Logger
}

// NewUserManager creates a new instance of UserManager with the given storage and logger.
func NewUserManager(storage storage.Storage, logger *logrus.Logger) *UserManager {
	return &UserManager{
		storage: storage,
		logger:  logger,
	}
}

// CreateUser creates a new user with the given name, email, and status.
func (um *UserManager) CreateUser(ctx context.Context, name, email string, status types.UserStatus) (*types.User, error) {
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

// GetUser fetches the user details from the storage.
func (um *UserManager) GetUser(ctx context.Context, uuid uuid.UUID) (*types.User, error) {
	return um.storage.GetUser(ctx, uuid)
}

// UpdateUserStatus updates the status of the user with the given UUID.
func (um *UserManager) UpdateUserStatus(ctx context.Context, uuid uuid.UUID, status types.UserStatus) error {
	return um.storage.UpdateUserStatus(ctx, uuid, status)
}

// ActivateUser sets the user status to active.
func (um *UserManager) ActivateUser(ctx context.Context, uuid uuid.UUID) error {
	return um.UpdateUserStatus(ctx, uuid, types.UserStatusActive)
}

// DeactivateUser sets the user status to inactive.
func (um *UserManager) DeactivateUser(ctx context.Context, uuid uuid.UUID) error {
	return um.UpdateUserStatus(ctx, uuid, types.UserStatusInactive)
}

// ResolveUserFromRequest is a middleware that extracts the user UUID from the request and fetches the user details from the storage.
func (um *UserManager) ResolveUserFromRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		um.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"raw_query":   r.URL.RawQuery,
			"request_uri": r.RequestURI,
		}).Debug("ResolveUserFromRequest called")

		rctx := chi.RouteContext(r.Context())
		if rctx == nil {
			um.logger.Error("No route context found")
			util.SendErrorResponse(w, http.StatusInternalServerError, "Internal server error", um.logger, nil)
			return
		}

		userUUID := rctx.URLParam("uuid")
		um.logger.WithField("uuid", userUUID).Debug("Extracted UUID from request")

		if userUUID == "" {
			um.logger.Error("Empty UUID parameter")
			util.SendErrorResponse(w, http.StatusBadRequest, "Invalid user UUID", um.logger, nil)
			return
		}

		parsedUUID, err := uuid.Parse(userUUID)
		if err != nil {
			um.logger.WithError(err).Error("Failed to parse UUID")
			util.SendErrorResponse(w, http.StatusBadRequest, "Invalid user UUID", um.logger, nil)
			return
		}

		user, err := um.GetUser(r.Context(), parsedUUID)
		if err != nil {
			if errors.Is(err, types.UserNotFoundError) {
				util.SendErrorResponse(w, http.StatusNotFound, "User not found", um.logger, nil)
			} else {
				util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get user", um.logger, nil)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RegisterRoutes registers the user management routes.
func (um *UserManager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(um.ResolveUserFromRequest)
			r.Get("/{uuid}", um.handleGetUser)
			r.Put("/{uuid}/activate", um.handleActivateUser)
			r.Put("/{uuid}/deactivate", um.handleDeactivateUser)
			r.Put("/{uuid}/status", um.handleUpdateUserStatus)
		})
	})
}

func (um *UserManager) handleGetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*types.User)
	util.SendSuccessResponse(w, http.StatusOK, user, um.logger, nil)
}

func (um *UserManager) handleActivateUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*types.User)
	err := um.ActivateUser(r.Context(), user.UUID)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to activate user", um.logger, nil)
		return
	}
	util.SendSuccessResponse(w, http.StatusOK, nil, um.logger, nil)
}

func (um *UserManager) handleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*types.User)
	err := um.DeactivateUser(r.Context(), user.UUID)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to deactivate user", um.logger, nil)
		return
	}
	util.SendSuccessResponse(w, http.StatusOK, nil, um.logger, nil)
}

func (um *UserManager) handleUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*types.User)
	var statusUpdate struct {
		Status types.UserStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		util.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body", um.logger, nil)
		return
	}

	err := um.UpdateUserStatus(r.Context(), user.UUID, statusUpdate.Status)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update user status", um.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, nil, um.logger, nil)
}
