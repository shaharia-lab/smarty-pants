package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
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

func (um *UserManager) GetPaginatedUsers(ctx context.Context, filter types.UserFilter, option types.UserFilterOption) (types.PaginatedUsers, error) {
	return um.storage.GetPaginatedUsers(ctx, filter, option)
}

// ResolveUserFromRequest is a middleware that extracts the user UUID from the request and fetches the user details from the storage.
func (um *UserManager) ResolveUserFromRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := observability.StartSpan(r.Context(), "auth.UserManager.ResolveUserFromRequest")
		defer span.End()

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

func (um *UserManager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/", um.handleListUsers)

		r.Group(func(r chi.Router) {
			r.Use(um.ResolveUserFromRequest)
			r.Get("/{uuid}", um.handleGetUser)
			r.Put("/{uuid}/activate", um.handleActivateUser)
			r.Put("/{uuid}/deactivate", um.handleDeactivateUser)
		})
	})
}

func (um *UserManager) handleListUsers(w http.ResponseWriter, r *http.Request) {
	_, span := observability.StartSpan(r.Context(), "auth.UserManager.handleListUsers")
	defer span.End()

	filter := types.UserFilter{
		NameContains:  r.URL.Query().Get("name"),
		EmailContains: r.URL.Query().Get("email"),
		Status:        types.UserStatus(r.URL.Query().Get("status")),
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage < 1 {
		perPage = 10 // Default to 10 per page
	}

	option := types.UserFilterOption{
		Page:    page,
		PerPage: perPage,
	}

	paginatedUsers, err := um.GetPaginatedUsers(r.Context(), filter, option)
	if err != nil {
		um.logger.WithError(err).Error("Failed to get paginated users")
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve users", um.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, paginatedUsers, um.logger, nil)
}

func (um *UserManager) handleGetUser(w http.ResponseWriter, r *http.Request) {
	_, span := observability.StartSpan(r.Context(), "auth.UserManager.handleGetUser")
	defer span.End()

	user := r.Context().Value(userContextKey).(*types.User)
	util.SendSuccessResponse(w, http.StatusOK, user, um.logger, nil)
}

func (um *UserManager) handleActivateUser(w http.ResponseWriter, r *http.Request) {
	_, span := observability.StartSpan(r.Context(), "auth.UserManager.handleActivateUser")
	defer span.End()

	user := r.Context().Value(userContextKey).(*types.User)
	err := um.ActivateUser(r.Context(), user.UUID)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to activate user", um.logger, nil)
		return
	}
	util.SendSuccessResponse(w, http.StatusOK, nil, um.logger, nil)
}

func (um *UserManager) handleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	_, span := observability.StartSpan(r.Context(), "auth.UserManager.handleDeactivateUser")
	defer span.End()

	user := r.Context().Value(userContextKey).(*types.User)
	err := um.DeactivateUser(r.Context(), user.UUID)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to deactivate user", um.logger, nil)
		return
	}
	util.SendSuccessResponse(w, http.StatusOK, nil, um.logger, nil)
}
