// Package auth provides JWT token management functionality
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the structure of your custom claims
type JWTClaims struct {
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	keyManager  *KeyManager
	userManager *UserManager
	logger      *logrus.Logger
}

// NewJWTManager creates a new JWTManager with the given KeyManager, UserManager and logger
func NewJWTManager(keyManager *KeyManager, userManager *UserManager, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		keyManager:  keyManager,
		userManager: userManager,
		logger:      logger,
	}
}

// IssueTokenForUser creates and signs a new JWT token for a user
func (m *JWTManager) IssueTokenForUser(ctx context.Context, userUUID uuid.UUID, audience []string, expiration time.Duration) (string, error) {
	m.logger.WithFields(logrus.Fields{
		"userUUID":   userUUID,
		"audience":   audience,
		"expiration": expiration,
	}).Debug("Attempting to issue new token for user")

	// Verify user exists and is active
	user, err := m.userManager.GetUser(ctx, userUUID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get user")
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.Status != "active" {
		m.logger.WithField("userUUID", userUUID).Error("User is not active")
		return "", errors.New("user is not active")
	}

	privateKey, _, err := m.keyManager.GetKeyPair()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get private key")
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	currentTime := time.Now().UTC()

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			NotBefore: jwt.NewNumericDate(currentTime),
			Issuer:    "smarty-pants",
			Subject:   user.UUID.String(),
			ID:        fmt.Sprintf("%d", currentTime.Unix()),
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		m.logger.WithError(err).Error("Failed to sign token")
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"tokenID":  claims.ID,
		"userUUID": user.UUID,
	}).Info("Token issued successfully for user")
	m.logger.WithFields(logrus.Fields{
		"tokenID":   claims.ID,
		"expiresAt": claims.ExpiresAt,
	}).Debug("Token details")
	return signedToken, nil
}

// ValidateToken verifies the given token and returns the claims if valid
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	m.logger.Debug("Attempting to validate token")

	_, publicKey, err := m.keyManager.GetKeyPair()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get public key")
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to parse token")
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		m.logger.Info("Token validated successfully")
		m.logger.WithFields(logrus.Fields{
			"subject":   claims.Subject,
			"issuer":    claims.Issuer,
			"expiresAt": claims.ExpiresAt,
		}).Debug("Validated token details")
		return claims, nil
	}

	m.logger.Error("Invalid token")
	return nil, errors.New("invalid token")
}
