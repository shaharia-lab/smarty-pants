package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserManager_handleGetUser(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	um := NewUserManager(mockStorage, logger)

	userUUID := uuid.New()
	user := &types.User{
		UUID:      userUUID,
		Name:      "Test User",
		Email:     "test@example.com",
		Status:    types.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.On("GetUser", mock.Anything, userUUID).Return(user, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Use(um.ResolveUserFromRequest)
			r.Get("/{uuid}", um.handleGetUser)
		})
	})

	reqURL := fmt.Sprintf("/api/v1/users/%s", userUUID.String())
	req, _ := http.NewRequest("GET", reqURL, nil)

	// Use Chi's test context to simulate route parameters
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("uuid", userUUID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response types.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, user.UUID, response.UUID)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Status, response.Status)

	mockStorage.AssertExpectations(t)
}

func TestUserManager_handleActivateUser(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	um := NewUserManager(mockStorage, logger)

	userUUID := uuid.New()
	user := &types.User{
		UUID:   userUUID,
		Status: types.UserStatusInactive,
	}

	mockStorage.On("UpdateUserStatus", mock.Anything, userUUID, types.UserStatusActive).Return(nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Put("/users/{uuid}/activate", um.handleActivateUser)

	req, _ := http.NewRequest("PUT", "/users/"+userUUID.String()+"/activate", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	mockStorage.AssertExpectations(t)
}

func TestUserManager_handleDeactivateUser(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	um := NewUserManager(mockStorage, logger)

	userUUID := uuid.New()
	user := &types.User{
		UUID:   userUUID,
		Status: types.UserStatusActive,
	}

	mockStorage.On("UpdateUserStatus", mock.Anything, userUUID, types.UserStatusInactive).Return(nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Put("/users/{uuid}/deactivate", um.handleDeactivateUser)

	req, _ := http.NewRequest("PUT", "/users/"+userUUID.String()+"/deactivate", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	mockStorage.AssertExpectations(t)
}

func TestUserManager_ResolveUserFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*storage.StorageMock)
		inputUUID      string
		expectedStatus int
		expectedUser   *types.User
	}{
		{
			name: "Valid UUID and user found",
			setupMock: func(m *storage.StorageMock) {
				userUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				user := &types.User{UUID: userUUID, Name: "Test User", Email: "test@example.com", Status: types.UserStatusActive}
				m.On("GetUser", mock.Anything, userUUID).Return(user, nil)
			},
			inputUUID:      "11111111-1111-1111-1111-111111111111",
			expectedStatus: http.StatusOK,
			expectedUser:   &types.User{UUID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Test User", Email: "test@example.com", Status: types.UserStatusActive},
		},
		{
			name:           "Invalid UUID",
			setupMock:      func(m *storage.StorageMock) {},
			inputUUID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedUser:   nil,
		},
		{
			name: "User not found",
			setupMock: func(m *storage.StorageMock) {
				userUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				m.On("GetUser", mock.Anything, userUUID).Return(nil, types.UserNotFoundError)
			},
			inputUUID:      "22222222-2222-2222-2222-222222222222",
			expectedStatus: http.StatusNotFound,
			expectedUser:   nil,
		},
		{
			name: "Internal server error",
			setupMock: func(m *storage.StorageMock) {
				userUUID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
				m.On("GetUser", mock.Anything, userUUID).Return(nil, errors.New("internal error"))
			},
			inputUUID:      "33333333-3333-3333-3333-333333333333",
			expectedStatus: http.StatusInternalServerError,
			expectedUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			logger := logrus.New()
			logger.SetOutput(io.Discard) // Suppress log output during tests
			um := NewUserManager(mockStorage, logger)

			tt.setupMock(mockStorage)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resolvedUser, ok := r.Context().Value(userContextKey).(*types.User)
				if tt.expectedUser != nil {
					assert.True(t, ok)
					assert.Equal(t, tt.expectedUser, resolvedUser)
				} else {
					assert.False(t, ok)
				}
				w.WriteHeader(http.StatusOK)
			})

			middlewareChain := um.ResolveUserFromRequest(handler)

			req, _ := http.NewRequest("GET", "/users/"+tt.inputUUID, nil)
			rr := httptest.NewRecorder()

			// Set up Chi context
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.inputUUID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			middlewareChain.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUserManager_ResolveUserFromRequest_InvalidUUID(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	um := NewUserManager(mockStorage, logger)

	r := chi.NewRouter()
	r.Use(um.ResolveUserFromRequest)
	r.Get("/users/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		t.Error("This handler should not be called")
	})

	req, _ := http.NewRequest("GET", "/users/invalid-uuid", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserManager_handleListUsers(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	um := NewUserManager(mockStorage, logger)

	user1 := &types.User{
		UUID:      uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Status:    types.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := &types.User{
		UUID:      uuid.New(),
		Name:      "Jane Smith",
		Email:     "jane@example.com",
		Status:    types.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockPaginatedUsers := types.PaginatedUsers{
		Users:      []types.User{*user1, *user2},
		Total:      2,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}

	mockStorage.On("GetPaginatedUsers", mock.Anything, mock.AnythingOfType("types.UserFilter"), mock.AnythingOfType("types.UserFilterOption")).
		Return(mockPaginatedUsers, nil)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/", um.handleListUsers)
		})
	})

	testCases := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedUsers  int
	}{
		{
			name:           "Default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedUsers:  2,
		},
		{
			name:           "With pagination",
			queryParams:    "?page=1&per_page=5",
			expectedStatus: http.StatusOK,
			expectedUsers:  2,
		},
		{
			name:           "With filtering",
			queryParams:    "?name=John&email=example.com&status=active",
			expectedStatus: http.StatusOK,
			expectedUsers:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/v1/users%s", tc.queryParams)
			req, _ := http.NewRequest("GET", reqURL, nil)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response types.PaginatedUsers
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUsers, len(response.Users))
			assert.Equal(t, mockPaginatedUsers.Total, response.Total)
			assert.Equal(t, mockPaginatedUsers.Page, response.Page)
			assert.Equal(t, mockPaginatedUsers.PerPage, response.PerPage)
			assert.Equal(t, mockPaginatedUsers.TotalPages, response.TotalPages)
		})
	}

	mockStorage.AssertExpectations(t)
}
