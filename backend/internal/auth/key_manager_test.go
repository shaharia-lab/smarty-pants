package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewKeyManager(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()

	km := NewKeyManager(mockStorage, l)

	assert.NotNil(t, km)
	assert.Equal(t, mockStorage, km.storage)
	assert.Equal(t, l, km.logger)
}

func TestGetKeyPair(t *testing.T) {
	t.Run("Existing key pair in storage", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		assert.NoError(t, err)

		mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)

		km := NewKeyManager(mockStorage, l)

		retrievedPrivateKey, retrievedPublicKey, err := km.GetKeyPair()

		assert.NoError(t, err)
		assert.NotNil(t, retrievedPrivateKey)
		assert.NotNil(t, retrievedPublicKey)
		assert.Equal(t, privateKey, retrievedPrivateKey)
		assert.Equal(t, &privateKey.PublicKey, retrievedPublicKey)

		mockStorage.AssertCalled(t, "GetKeyPair")
	})

	t.Run("No existing key pair, generate new one", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		mockStorage.On("GetKeyPair").Return([]byte(nil), []byte(nil), errors.New("key pair not found"))
		mockStorage.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(nil)

		km := NewKeyManager(mockStorage, l)

		privateKey, publicKey, err := km.GetKeyPair()

		assert.NoError(t, err)
		assert.NotNil(t, privateKey)
		assert.NotNil(t, publicKey)

		mockStorage.AssertCalled(t, "GetKeyPair")
		mockStorage.AssertCalled(t, "UpdateKeyPair", mock.Anything, mock.Anything)
	})

	t.Run("Error parsing existing private key", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		mockStorage.On("GetKeyPair").Return([]byte("invalid private key"), []byte("valid public key"), nil)

		km := NewKeyManager(mockStorage, l)

		privateKey, publicKey, err := km.GetKeyPair()

		assert.Error(t, err)
		assert.Nil(t, privateKey)
		assert.Nil(t, publicKey)
		assert.Contains(t, err.Error(), "failed to parse private key")

		mockStorage.AssertCalled(t, "GetKeyPair")
	})

	t.Run("Error parsing existing public key", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

		mockStorage.On("GetKeyPair").Return(privateKeyBytes, []byte("invalid public key"), nil)

		km := NewKeyManager(mockStorage, l)

		retrievedPrivateKey, retrievedPublicKey, err := km.GetKeyPair()

		assert.Error(t, err)
		assert.Nil(t, retrievedPrivateKey)
		assert.Nil(t, retrievedPublicKey)
		assert.Contains(t, err.Error(), "failed to parse public key")

		mockStorage.AssertCalled(t, "GetKeyPair")
	})

	t.Run("Error generating new key pair", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		mockStorage.On("GetKeyPair").Return([]byte(nil), []byte(nil), errors.New("key pair not found"))

		// Create a custom KeyManager with a mocked generateKeyPair function
		km := &KeyManager{
			storage: mockStorage,
			logger:  l,
		}

		// Override the generateKeyPair method for this test
		originalGenerateKeyPair := generateKeyPair
		defer func() { generateKeyPair = originalGenerateKeyPair }()
		generateKeyPair = func(random io.Reader, bits int) (*rsa.PrivateKey, error) {
			return nil, errors.New("failed to generate key pair")
		}

		privateKey, publicKey, err := km.GetKeyPair()

		assert.Error(t, err)
		assert.Nil(t, privateKey)
		assert.Nil(t, publicKey)
		assert.Contains(t, err.Error(), "failed to generate new key pair")

		mockStorage.AssertCalled(t, "GetKeyPair")
		mockStorage.AssertNotCalled(t, "UpdateKeyPair")
	})

	t.Run("Error storing new key pair", func(t *testing.T) {
		mockStorage := new(storage.StorageMock)
		l := logger.NoOpsLogger()

		mockStorage.On("GetKeyPair").Return([]byte(nil), []byte(nil), errors.New("key pair not found"))
		mockStorage.On("UpdateKeyPair", mock.Anything, mock.Anything).Return(errors.New("failed to store key pair"))

		km := NewKeyManager(mockStorage, l)

		privateKey, publicKey, err := km.GetKeyPair()

		assert.Error(t, err)
		assert.Nil(t, privateKey)
		assert.Nil(t, publicKey)
		assert.Contains(t, err.Error(), "failed to store new key pair")

		mockStorage.AssertCalled(t, "GetKeyPair")
		mockStorage.AssertCalled(t, "UpdateKeyPair", mock.Anything, mock.Anything)
	})
}

func TestGetJWTSigningMethod(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	l := logger.NoOpsLogger()

	km := NewKeyManager(mockStorage, l)

	signingMethod := km.GetJWTSigningMethod()

	assert.Equal(t, jwt.SigningMethodRS256, signingMethod)
}

func TestParseRSAPrivateKey(t *testing.T) {
	t.Run("Valid PKCS1 private key", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

		parsedKey, err := ParseRSAPrivateKey(privateKeyBytes)

		assert.NoError(t, err)
		assert.NotNil(t, parsedKey)
		assert.Equal(t, privateKey, parsedKey)
	})

	t.Run("Valid PKCS8 private key", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		assert.NoError(t, err)

		parsedKey, err := ParseRSAPrivateKey(privateKeyBytes)

		assert.NoError(t, err)
		assert.NotNil(t, parsedKey)
		assert.Equal(t, privateKey, parsedKey)
	})

	t.Run("Invalid private key", func(t *testing.T) {
		invalidKeyBytes := []byte("invalid key")

		parsedKey, err := ParseRSAPrivateKey(invalidKeyBytes)

		assert.Error(t, err)
		assert.Nil(t, parsedKey)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})
}

func TestParseRSAPublicKey(t *testing.T) {
	t.Run("Valid public key", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		assert.NoError(t, err)

		parsedKey, err := ParseRSAPublicKey(publicKeyBytes)

		assert.NoError(t, err)
		assert.NotNil(t, parsedKey)
		assert.Equal(t, &privateKey.PublicKey, parsedKey)
	})

	t.Run("Invalid public key", func(t *testing.T) {
		invalidKeyBytes := []byte("invalid key")

		parsedKey, err := ParseRSAPublicKey(invalidKeyBytes)

		assert.Error(t, err)
		assert.Nil(t, parsedKey)
		assert.Contains(t, err.Error(), "failed to parse public key")
	})
}
