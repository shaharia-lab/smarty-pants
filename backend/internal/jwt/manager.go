package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// Config contains the JWT-related configuration
type Config struct {
	RSAPrivateKey     string `envconfig:"JWT_RSA_PRIVATE_KEY"`
	RSAPublicKey      string `envconfig:"JWT_RSA_PUBLIC_KEY"`
	RSAPrivateKeyFile string `envconfig:"JWT_RSA_PRIVATE_KEY_FILE"`
	RSAPublicKeyFile  string `envconfig:"JWT_RSA_PUBLIC_KEY_FILE"`
}

// MyCustomClaims represents the structure of your custom claims
type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     *Config
	logger     *logrus.Logger
}

// NewJWTManager creates a new JWTManager with the given configuration and logger
func NewJWTManager(config *Config, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		config: config,
		logger: logger,
	}
}

// LoadKeys loads RSA keys from the configuration
func (m *JWTManager) LoadKeys() error {
	var err error

	// Try to load private key
	if m.config.RSAPrivateKey != "" {
		m.privateKey, err = parseRSAPrivateKeyFromPEM([]byte(m.config.RSAPrivateKey))
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse private key from config")
			return fmt.Errorf("failed to parse private key: %v", err)
		}
	} else if m.config.RSAPrivateKeyFile != "" {
		privateKeyBytes, err := os.ReadFile(m.config.RSAPrivateKeyFile)
		if err != nil {
			m.logger.WithError(err).Error("Failed to read private key file")
			return fmt.Errorf("failed to read private key file: %v", err)
		}
		m.privateKey, err = parseRSAPrivateKeyFromPEM(privateKeyBytes)
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse private key from file")
			return fmt.Errorf("failed to parse private key from file: %v", err)
		}
	} else {
		m.logger.Error("No private key or private key file provided")
		return errors.New("no private key or private key file provided")
	}

	// Try to load public key
	if m.config.RSAPublicKey != "" {
		m.publicKey, err = parseRSAPublicKeyFromPEM([]byte(m.config.RSAPublicKey))
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse public key from config")
			return fmt.Errorf("failed to parse public key: %v", err)
		}
	} else if m.config.RSAPublicKeyFile != "" {
		publicKeyBytes, err := os.ReadFile(m.config.RSAPublicKeyFile)
		if err != nil {
			m.logger.WithError(err).Error("Failed to read public key file")
			return fmt.Errorf("failed to read public key file: %v", err)
		}
		m.publicKey, err = parseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			m.logger.WithError(err).Error("Failed to parse public key from file")
			return fmt.Errorf("failed to parse public key from file: %v", err)
		}
	} else {
		m.logger.Error("No public key or public key file provided")
		return errors.New("no public key or public key file provided")
	}

	m.logger.Info("RSA keys loaded successfully")
	return nil
}

func parseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func parseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// IssueToken creates and signs a new JWT token
func (m *JWTManager) IssueToken(foo string, subject string, audience []string, expiration time.Duration) (string, error) {
	if m.privateKey == nil {
		m.logger.Error("Private key not loaded")
		return "", errors.New("private key not loaded")
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
	signedToken, err := token.SignedString(m.privateKey)
	if err != nil {
		m.logger.WithError(err).Error("Failed to sign token")
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	m.logger.Info("Token issued successfully")
	return signedToken, nil
}

// ValidateToken verifies the given token and returns the claims if valid
func (m *JWTManager) ValidateToken(tokenString string) (*MyCustomClaims, error) {
	if m.publicKey == nil {
		m.logger.Error("Public key not loaded")
		return nil, errors.New("public key not loaded")
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
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
	return nil, fmt.Errorf("invalid token")
}
