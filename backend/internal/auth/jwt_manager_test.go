package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewJWTManager(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	mockUserManager := NewUserManager(mockStorage, l)
	keyManager := NewKeyManager(mockStorage, l)

	jwtManager := NewJWTManager(keyManager, mockUserManager, l, []string{})

	assert.NotNil(t, jwtManager)
	assert.Equal(t, keyManager, jwtManager.keyManager)
	assert.Equal(t, mockUserManager, jwtManager.userManager)
	assert.Equal(t, l, jwtManager.logger)
}

func TestIssueTokenForUser(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	mockUserManager := NewUserManager(mockStorage, l)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, mockUserManager, l, []string{})

	tests := []struct {
		name        string
		userUUID    uuid.UUID
		audience    []string
		expiration  time.Duration
		mockUser    *types.User
		mockError   error
		expectError bool
	}{
		{
			name:       "Valid active user",
			userUUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			audience:   []string{"web", "mobile"},
			expiration: time.Hour,
			mockUser: &types.User{
				UUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:  "user@example.com",
				Status: "active",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:       "Inactive user",
			userUUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			audience:   []string{"web"},
			expiration: time.Hour,
			mockUser: &types.User{
				UUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:  "inactive@example.com",
				Status: "inactive",
			},
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "Non-existent user",
			userUUID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			audience:    []string{"web"},
			expiration:  time.Hour,
			mockUser:    nil,
			mockError:   errors.New("user not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.On("GetUser", mock.Anything, tt.userUUID).Return(tt.mockUser, tt.mockError).Once()

			token, err := jwtManager.IssueToken(context.Background(), tt.userUUID, tt.audience, tt.expiration)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify the token
				parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return &privateKey.PublicKey, nil
				})

				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(*JWTClaims)
				assert.True(t, ok)
				assert.Equal(t, tt.mockUser.UUID.String(), claims.Subject)
				assert.Equal(t, tt.userUUID.String(), claims.Subject)
				assert.ElementsMatch(t, tt.audience, claims.Audience)
				assert.WithinDuration(t, time.Now().Add(tt.expiration), claims.ExpiresAt.Time, 5*time.Second)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	mockUserManager := NewUserManager(mockStorage, l)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, mockUserManager, l, []string{})

	validUser := &types.User{
		UUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Email:  "user@example.com",
		Status: "active",
	}
	mockStorage.On("GetUser", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(validUser, nil)

	validToken, err := jwtManager.IssueToken(context.Background(), uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), []string{"web"}, time.Hour)
	assert.NoError(t, err)

	expiredClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			Subject:   "user@example.com",
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString(privateKey)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		token            string
		expectError      bool
		expectedSubject  string
		expectedAudience []string
		expectedUserUUID string
	}{
		{
			name:             "Valid token",
			token:            validToken,
			expectError:      false,
			expectedSubject:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000").String(),
			expectedAudience: []string{"web"},
			expectedUserUUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000").String(),
		},
		{
			name:        "Expired token",
			token:       expiredTokenString,
			expectError: true,
		},
		{
			name:        "Invalid token",
			token:       "invalid.token.string",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectedSubject, claims.Subject)
				assert.ElementsMatch(t, tt.expectedAudience, claims.Audience)
				assert.Equal(t, tt.expectedUserUUID, claims.Subject)
			}
		})
	}
}

func TestJWTManagerWithKeyManagerErrors(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	mockUserManager := NewUserManager(mockStorage, l)

	mockStorage.On("GetKeyPair").Return([]byte(nil), []byte(nil), errors.New("key manager error"))
	mockStorage.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(errors.New("update key pair error"))

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, mockUserManager, l, []string{})

	validUser := &types.User{
		UUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Email:  "user@example.com",
		Status: "active",
	}
	mockStorage.On("GetUser", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(validUser, nil)

	t.Run("IssueToken with KeyManager error", func(t *testing.T) {
		token, err := jwtManager.IssueToken(context.Background(), uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), []string{"web"}, time.Hour)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "failed to get private key")
	})

	t.Run("ValidateToken with KeyManager error", func(t *testing.T) {
		claims, err := jwtManager.ValidateToken("some.token.string")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "failed to get public key")
	})

	mockStorage.AssertCalled(t, "GetKeyPair")
	mockStorage.AssertExpectations(t)
}

func TestAuthMiddleware(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	mockUserManager := NewUserManager(mockStorage, l)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil).Maybe()

	keyManager := NewKeyManager(mockStorage, l)
	skipAuthEndpoints := []string{"/api/v1/public", "/api/v1/analytics/overview"}
	jwtManager := NewJWTManager(keyManager, mockUserManager, l, skipAuthEndpoints)

	validUser := &types.User{
		UUID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Email:  "user@example.com",
		Status: types.UserStatusActive,
		Roles:  []types.UserRole{types.UserRoleUser},
	}

	mockStorage.On("GetUser", mock.Anything, validUser.UUID).Return(validUser, nil).Maybe()

	validToken, err := jwtManager.IssueToken(context.Background(), validUser.UUID, []string{"web"}, time.Hour)
	assert.NoError(t, err)

	anonymousUser := &types.User{
		UUID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		Name:   "Anonymous User",
		Email:  "anonymous@example.com",
		Status: types.UserStatusActive,
		Roles:  []types.UserRole{types.UserRoleAdmin},
	}

	tests := []struct {
		name           string
		authEnabled    bool
		path           string
		setupRequest   func(*http.Request)
		expectedStatus int
		expectedUser   *types.User
		expectedError  *util.APIError
	}{
		{
			name:        "Auth disabled",
			authEnabled: false,
			setupRequest: func(req *http.Request) {
				// No token needed
			},
			expectedStatus: http.StatusOK,
			expectedUser: &types.User{
				UUID:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
				Name:   "Anonymous User",
				Email:  "anonymous@example.com",
				Status: types.UserStatusActive,
				Roles:  []types.UserRole{types.UserRoleAdmin},
			},
		},
		{
			name:        "Auth enabled, valid token",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+validToken)
			},
			expectedStatus: http.StatusOK,
			expectedUser:   validUser,
		},
		{
			name:        "Auth enabled, no Authorization header",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				// No Authorization header
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "authorization header is missing",
			},
		},
		{
			name:        "Auth enabled, empty Authorization header",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "authorization header is missing",
			},
		},
		{
			name:        "Auth enabled, invalid Authorization header format",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "InvalidFormat")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "authorization header format must be Bearer {token}",
			},
		},
		{
			name:        "Auth enabled, no space after Bearer",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer"+validToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "authorization header format must be Bearer {token}",
			},
		},
		{
			name:        "Auth enabled, invalid token",
			authEnabled: true,
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalidtoken")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "failed to parse token: token is malformed: token contains an invalid number of segments",
			},
		},
		{
			name:        "Auth enabled, skip path, no token",
			authEnabled: true,
			path:        "/api/v1/public",
			setupRequest: func(req *http.Request) {
				// No token provided
			},
			expectedStatus: http.StatusOK,
			expectedUser:   anonymousUser,
		},
		{
			name:        "Auth enabled, skip path, with valid token",
			authEnabled: true,
			path:        "/api/v1/analytics/overview",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+validToken)
			},
			expectedStatus: http.StatusOK,
			expectedUser:   anonymousUser, // Should still be anonymous due to skip path
		},
		{
			name:        "Auth enabled, non-skip path, no token",
			authEnabled: true,
			path:        "/api/v1/protected",
			setupRequest: func(req *http.Request) {
				// No token provided
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "authorization header is missing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)

			tt.setupRequest(req)

			rr := httptest.NewRecorder()

			handler := jwtManager.AuthMiddleware(tt.authEnabled)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedUser != nil {
					user := r.Context().Value(AuthenticatedUserCtxKey).(*types.User)
					assert.Equal(t, tt.expectedUser.UUID, user.UUID)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
					assert.Equal(t, tt.expectedUser.Status, user.Status)
					assert.ElementsMatch(t, tt.expectedUser.Roles, user.Roles)
				}
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != nil {
				var apiError util.APIError
				err := json.Unmarshal(rr.Body.Bytes(), &apiError)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError.Message, apiError.Message)
				assert.Equal(t, tt.expectedError.Err, apiError.Err)
			}
		})
	}
}

func TestAuthMiddlewareWithErrors(t *testing.T) {
	l := logger.NoOpsLogger()
	validUserUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	tests := []struct {
		name           string
		setupMocks     func(*storage.StorageMock) string // Return the token to be used in the test
		expectedStatus int
		expectedError  *util.APIError
	}{
		{
			name: "Error getting key pair",
			setupMocks: func(mockStorage *storage.StorageMock) string {
				mockStorage.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(errors.New("key pair error")).Once()
				mockStorage.On("GetKeyPair").Return([]byte{}, []byte{}, errors.New("key pair error")).Once()
				return "validtokenformat"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "failed to get public key: failed to store new key pair: key pair error",
			},
		},
		{
			name: "Invalid token",
			setupMocks: func(mockStorage *storage.StorageMock) string {
				privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
				privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
				publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
				mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil).Once()
				return "invalidtoken"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "failed to parse token: token is malformed: token contains an invalid number of segments",
			},
		},
		{
			name: "Error getting user",
			setupMocks: func(mockStorage *storage.StorageMock) string {
				privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: validUserUUID.String(),
					},
				})
				validToken, _ := token.SignedString(privateKey)

				privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
				publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
				mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil).Once()
				mockStorage.On("GetUser", mock.Anything, validUserUUID).Return(nil, errors.New("user not found")).Once()

				return validToken
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError: &util.APIError{
				Message: "Un-Authorized",
				Err:     "Invalid authentication",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			mockUserManager := NewUserManager(mockStorage, l)
			keyManager := NewKeyManager(mockStorage, l)
			jwtManager := NewJWTManager(keyManager, mockUserManager, l, []string{})

			token := tt.setupMocks(mockStorage)

			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			req.Header.Set("Authorization", "Bearer "+token)

			rr := httptest.NewRecorder()

			handler := jwtManager.AuthMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Handler should not be called")
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var apiError util.APIError
			err = json.Unmarshal(rr.Body.Bytes(), &apiError)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError.Message, apiError.Message)
			assert.Contains(t, apiError.Err, tt.expectedError.Err, "Error message should contain expected error")

			mockStorage.AssertExpectations(t)
		})
	}
}
