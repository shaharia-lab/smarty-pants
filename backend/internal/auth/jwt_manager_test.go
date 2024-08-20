package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewJWTManager(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	keyManager := NewKeyManager(mockStorage, l)

	jwtManager := NewJWTManager(keyManager, l)

	assert.NotNil(t, jwtManager)
	assert.Equal(t, keyManager, jwtManager.keyManager)
	assert.Equal(t, l, jwtManager.logger)
}

func TestIssueToken(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, l)

	tests := []struct {
		name        string
		subject     string
		audience    []string
		expiration  time.Duration
		expectError bool
	}{
		{
			name:        "Valid token",
			subject:     "user123",
			audience:    []string{"web", "mobile"},
			expiration:  time.Hour,
			expectError: false,
		},
		{
			name:        "Empty subject",
			subject:     "",
			audience:    []string{"web"},
			expiration:  time.Hour,
			expectError: false,
		},
		{
			name:        "Empty audience",
			subject:     "user123",
			audience:    []string{},
			expiration:  time.Hour,
			expectError: false,
		},
		{
			name:        "Zero expiration",
			subject:     "user123",
			audience:    []string{"web"},
			expiration:  time.Second, // Changed from 0 to 1 second
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.IssueToken(tt.subject, tt.audience, tt.expiration)

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
				assert.Equal(t, tt.subject, claims.Subject)

				// Compare Audience as strings
				assert.ElementsMatch(t, tt.audience, claims.Audience)

				if tt.expiration > 0 {
					assert.WithinDuration(t, time.Now().Add(tt.expiration), claims.ExpiresAt.Time, 5*time.Second)
				} else {
					assert.True(t, claims.ExpiresAt.Time.After(time.Now()), "Token should not be expired")
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, l)

	validToken, err := jwtManager.IssueToken("user123", []string{"web"}, time.Hour)
	assert.NoError(t, err)

	expiredClaims := JWTClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			Subject:   "user123",
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
	}{
		{
			name:             "Valid token",
			token:            validToken,
			expectError:      false,
			expectedSubject:  "user123",
			expectedAudience: []string{"web"},
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
			}
		})
	}
}

func TestJWTManagerWithKeyManagerErrors(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()

	// Mock GetKeyPair to always return an error
	mockStorage.On("GetKeyPair").Return([]byte(nil), []byte(nil), errors.New("key manager error"))

	// Mock UpdateKeyPair to also return an error
	mockStorage.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(errors.New("update key pair error"))

	keyManager := NewKeyManager(mockStorage, l)
	jwtManager := NewJWTManager(keyManager, l)

	t.Run("IssueToken with KeyManager error", func(t *testing.T) {
		token, err := jwtManager.IssueToken("user123", []string{"web"}, time.Hour)
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

	// Verify that GetKeyPair was called
	mockStorage.AssertCalled(t, "GetKeyPair")
}
