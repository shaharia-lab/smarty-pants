// Package auth provides JWT token management functionality
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the structure of your custom claims
type JWTClaims struct {
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	keyManager        *KeyManager
	userManager       *UserManager
	logger            *logrus.Logger
	skipAuthEndpoints []string
}

// NewJWTManager creates a new JWTManager with the given KeyManager, UserManager and logger
func NewJWTManager(keyManager *KeyManager, userManager *UserManager, logger *logrus.Logger, skipAuthEndpoints []string) *JWTManager {
	return &JWTManager{
		keyManager:        keyManager,
		userManager:       userManager,
		logger:            logger,
		skipAuthEndpoints: skipAuthEndpoints,
	}
}

// IssueToken creates and signs a new JWT token for a user
func (m *JWTManager) IssueToken(ctx context.Context, userUUID uuid.UUID, audience []string, expiration time.Duration) (string, error) {
	m.logger.WithFields(logrus.Fields{
		"userUUID":   userUUID,
		"audience":   audience,
		"expiration": expiration,
	}).Debug("Attempting to issue new token for user")

	user, err := m.userManager.GetUser(ctx, userUUID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get user")
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.Status != types.UserStatusActive {
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

// AuthMiddleware is a middleware function that accepts multiple arguments
func (m *JWTManager) AuthMiddleware(authEnabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authEnabled {
				anonymousUser, err := m.userManager.GetAnonymousUser(r.Context())
				if err != nil {
					m.logger.WithError(err).Error("Failed to get anonymous user")
					util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Internal server error", Err: err.Error()})
					return
				}

				if anonymousUser == nil {
					m.logger.Error("Anonymous user is nil")
					util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Internal server error", Err: "Anonymous user doesn't exists in the system"})
					return
				}

				r.WithContext(context.WithValue(r.Context(), types.AuthenticatedUserCtxKey, anonymousUser))
				next.ServeHTTP(w, r)
				return
			}

			if m.isPathInSkipList(r.URL.Path) {
				r.WithContext(context.WithValue(r.Context(), types.AuthenticatedUserCtxKey, types.DefaultAnonymousUser()))
				next.ServeHTTP(w, r)
				return
			}

			m.processAccessToken(w, r, next)
		})
	}
}

func (m *JWTManager) isPathInSkipList(path string) bool {
	for _, skipPath := range m.skipAuthEndpoints {
		if path == skipPath {
			return true
		}
	}

	return false
}

func (m *JWTManager) processAccessToken(w http.ResponseWriter, r *http.Request, next http.Handler) {
	accessToken, err := m.resolveAccessTokenFromRequest(r)
	if err != nil {
		m.logger.WithError(err).Error("Failed to resolve access token from request header")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Un-Authorized", Err: err.Error()})
		return
	}

	if accessToken == "" {
		m.logger.Error("Access token is missing or empty")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Invalid credentials", Err: "Access token is missing"})
		return
	}

	jwtClaims, err := m.ValidateToken(accessToken)
	if err != nil {
		m.logger.WithError(err).Error("Failed to validate token")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Un-Authorized", Err: err.Error()})
		return
	}

	userUUID, err := uuid.Parse(jwtClaims.Subject)
	if err != nil {
		m.logger.WithError(err).Error("Failed to parse user UUID")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Un-Authorized", Err: "Invalid authentication"})
		return
	}

	user, err := m.userManager.GetUser(r.Context(), userUUID)
	if err != nil {
		m.logger.WithField("jwt_claim_subject", jwtClaims.Subject).WithError(err).Error("Failed to get user")
		util.SendAPIErrorResponse(w, http.StatusUnauthorized, &util.APIError{Message: "Un-Authorized", Err: "Invalid authentication"})
		return
	}

	m.logger.WithField("userUUID", user.UUID).Debug("User authenticated successfully. Setting user in request context")
	ctx := context.WithValue(r.Context(), types.AuthenticatedUserCtxKey, user)
	r = r.WithContext(ctx)

	next.ServeHTTP(w, r)
}

func (m *JWTManager) resolveAccessTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}
