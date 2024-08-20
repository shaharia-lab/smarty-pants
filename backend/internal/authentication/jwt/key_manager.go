package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const (
	defaultKeySize = 2048 // RSA-256 typically uses 2048-bit keys
)

// Storage interface for database operations
type Storage interface {
	GetKeyPair() (privateKey, publicKey []byte, err error)
	UpdateKeyPair(privateKey, publicKey []byte) error
}

// KeyManager handles RSA key pair operations
type KeyManager struct {
	storage Storage
	logger  *logrus.Logger
}

// NewKeyManager creates a new KeyManager instance
func NewKeyManager(storage Storage, logger *logrus.Logger) *KeyManager {
	return &KeyManager{
		storage: storage,
		logger:  logger,
	}
}

// GetKeyPair retrieves the key pair from storage or generates a new one if not found
func (km *KeyManager) GetKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Try to retrieve the key pair from storage
	privateKeyBytes, publicKeyBytes, err := km.storage.GetKeyPair()
	if err == nil {
		km.logger.Info("Retrieved existing key pair from storage")
		privateKey, err := ParseRSAPrivateKey(privateKeyBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		publicKey, err := ParseRSAPublicKey(publicKeyBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse public key: %v", err)
		}
		return privateKey, publicKey, nil
	}

	// If not found or error occurred, generate a new key pair
	km.logger.Info("No existing key pair found, generating new one")
	privateKey, err := km.generateKeyPair()
	if err != nil {
		km.logger.WithError(err).Error("Failed to generate new key pair")
		return nil, nil, err
	}

	// Convert keys to DER format for storage
	privateKeyBytes = x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		km.logger.WithError(err).Error("Failed to marshal public key")
		return nil, nil, err
	}

	// Store the new key pair
	err = km.storage.UpdateKeyPair(privateKeyBytes, publicKeyBytes)
	if err != nil {
		km.logger.WithError(err).Error("Failed to store new key pair")
		return nil, nil, err
	}

	km.logger.Info("Generated and stored new key pair")
	return privateKey, &privateKey.PublicKey, nil
}

// generateKeyPair creates a new RSA key pair
func (km *KeyManager) generateKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, defaultKeySize)
}

// GetJWTSigningMethod returns the JWT signing method (always RS256)
func (km *KeyManager) GetJWTSigningMethod() jwt.SigningMethod {
	return jwt.SigningMethodRS256
}

// ParseRSAPrivateKey parses a DER encoded private key
func ParseRSAPrivateKey(derBytes []byte) (*rsa.PrivateKey, error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(derBytes)
	if err != nil {
		// Try PKCS8 format if PKCS1 fails
		key, err := x509.ParsePKCS8PrivateKey(derBytes)
		if err != nil {
			return nil, err
		}
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return privateKey, nil
	}
	return privateKey, nil
}

// ParseRSAPublicKey parses a DER encoded public key
func ParseRSAPublicKey(derBytes []byte) (*rsa.PublicKey, error) {
	publicKey, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, err
	}
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPublicKey, nil
}
