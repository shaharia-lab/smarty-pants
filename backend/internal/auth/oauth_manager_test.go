// File: oauth_manager_test.go

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	logger2 "github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OAuthProvider
type MockOAuthProvider struct {
	mock.Mock
}

func (m *MockOAuthProvider) GetAuthURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func (m *MockOAuthProvider) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

func (m *MockOAuthProvider) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*UserInfo), args.Error(1)
}

func TestInitiateAuthFlow(t *testing.T) {
	mockProvider := new(MockOAuthProvider)
	storageMock := new(storage.StorageMock)
	logger := logger2.NoOpsLogger()
	mockUserManager := NewUserManager(storageMock, logger)
	jwtManager := NewJWTManager(NewKeyManager(storageMock, logger), mockUserManager, logger)

	providers := map[string]OAuthProvider{
		"google": mockProvider,
	}

	oauthManager := NewOAuthManager(providers, mockUserManager, jwtManager, logger)

	mockProvider.On("GetAuthURL", mock.AnythingOfType("string")).Return("https://example.com/auth")

	reqBody := AuthFlowRequest{
		AuthFlow: struct {
			Provider   string `json:"provider"`
			CurrentURL string `json:"current_url,omitempty"`
		}{
			Provider: "google",
		},
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/initiate", bytes.NewReader(reqBodyBytes))
	rr := httptest.NewRecorder()

	oauthManager.InitiateAuthFlow(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AuthFlowResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "google", resp.AuthFlow.Provider)
	assert.Equal(t, "https://example.com/auth", resp.AuthFlow.AuthRedirectURL)
	assert.NotEmpty(t, resp.AuthFlow.State)

	mockProvider.AssertExpectations(t)
}

func TestHandleAuthCode(t *testing.T) {
	mockProvider := new(MockOAuthProvider)
	storageMock := new(storage.StorageMock)
	storageMock.On("GetUser", mock.Anything, mock.Anything).Return(&types.User{
		UUID:   uuid.New(),
		Email:  "test@example.com",
		Name:   "Test User",
		Status: types.UserStatusActive,
	}, nil)
	storageMock.On("GetKeyPair").Return(nil, nil, errors.New("no key pair found"))
	storageMock.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(nil)

	storageMock.On("GetPaginatedUsers", mock.Anything, mock.AnythingOfType("types.UserFilter"), mock.AnythingOfType("types.UserFilterOption")).
		Return(func(ctx context.Context, filter types.UserFilter, option types.UserFilterOption) types.PaginatedUsers {
			return types.PaginatedUsers{
				Users: []types.User{
					{
						UUID:  uuid.New(),
						Email: "test@example.com",
						Name:  "Test User",
					},
				},
				Total:      1,
				Page:       option.Page,
				PerPage:    option.PerPage,
				TotalPages: 1,
			}
		}, nil)

	logger := logger2.NoOpsLogger()
	mockUserManager := NewUserManager(storageMock, logger)
	keyManager := NewKeyManager(storageMock, logger)
	jwtManager := NewJWTManager(keyManager, mockUserManager, logger)

	providers := map[string]OAuthProvider{
		"google": mockProvider,
	}

	oauthManager := NewOAuthManager(providers, mockUserManager, jwtManager, logger)

	// Set up the state
	state := uuid.New().String()
	oauthManager.stateStore[state] = stateInfo{
		provider:  "google",
		timestamp: time.Now(),
	}

	mockProvider.On("ExchangeCodeForToken", mock.Anything, "test_code").Return("access_token", nil)
	mockProvider.On("GetUserInfo", mock.Anything, "access_token").Return(&UserInfo{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Test User",
	}, nil)

	reqBody := AuthCodeRequest{
		AuthFlow: struct {
			Provider string `json:"provider"`
			AuthCode string `json:"auth_code"`
			State    string `json:"state"`
		}{
			Provider: "google",
			AuthCode: "test_code",
			State:    state,
		},
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/callback", bytes.NewReader(reqBodyBytes))
	rr := httptest.NewRecorder()

	oauthManager.HandleAuthCode(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp AuthTokenResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(strings.Split(resp.AccessToken, ".")))

	mockProvider.AssertExpectations(t)
	storageMock.AssertExpectations(t)
}

func TestRegisterRoutes(t *testing.T) {
	mockProvider := new(MockOAuthProvider)
	storageMock := new(storage.StorageMock)
	logger := logger2.NoOpsLogger()
	mockUserManager := NewUserManager(storageMock, logger)
	mockJWTManager := NewJWTManager(NewKeyManager(storageMock, logger), mockUserManager, logger)

	providers := map[string]OAuthProvider{
		"google": mockProvider,
	}

	oauthManager := NewOAuthManager(providers, mockUserManager, mockJWTManager, logger)

	r := chi.NewRouter()
	oauthManager.RegisterRoutes(r)

	assert.NotNil(t, r.Match(chi.NewRouteContext(), "POST", "/api/v1/auth/initiate"))
	assert.NotNil(t, r.Match(chi.NewRouteContext(), "POST", "/api/v1/auth/callback"))
}
