package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// MyCustomClaims represents the structure of your custom claims
type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	keyManager *KeyManager
	logger     *logrus.Logger
}

// NewJWTManager creates a new JWTManager with the given KeyManager and logger
func NewJWTManager(keyManager *KeyManager, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		keyManager: keyManager,
		logger:     logger,
	}
}

// IssueToken creates and signs a new JWT token
func (m *JWTManager) IssueToken(foo string, subject string, audience []string, expiration time.Duration) (string, error) {
	privateKey, _, err := m.keyManager.GetKeyPair()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get private key")
		return "", fmt.Errorf("failed to get private key: %v", err)
	}

	claims := MyCustomClaims{
		foo,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app-name",
			Subject:   subject,
			ID:        fmt.Sprintf("%d", time.Now().Unix()),
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		m.logger.WithError(err).Error("Failed to sign token")
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	m.logger.Info("Token issued successfully")
	return signedToken, nil
}

// ValidateToken verifies the given token and returns the claims if valid
func (m *JWTManager) ValidateToken(tokenString string) (*MyCustomClaims, error) {
	_, publicKey, err := m.keyManager.GetKeyPair()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get public key")
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to parse token")
		return nil, err
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		m.logger.Info("Token validated successfully")
		return claims, nil
	}

	m.logger.Error("Invalid token")
	return nil, errors.New("invalid token")
}
