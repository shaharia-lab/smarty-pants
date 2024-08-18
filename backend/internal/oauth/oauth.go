package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Provider string

var ErrUserNotFound = errors.New("user not found")

const (
	Google Provider = "google"
	// Add other providers here in the future
)

type OAuthManager struct {
	configs map[Provider]*oauth2.Config
	storage Storage
	logger  *logrus.Logger
}

type Storage interface {
	SaveUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveOAuthToken(ctx context.Context, token *OAuthToken) error
	GetOAuthToken(ctx context.Context, userID int64, provider Provider) (*OAuthToken, error)
}

var TokenNotFound = errors.New("token not found")

type User struct {
	ID             int64
	Name           string
	Email          string
	AuthProvider   Provider
	AuthProviderID string
}

type OAuthToken struct {
	UserID       int64
	Provider     Provider
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func NewOAuthManager(storage Storage, logger *logrus.Logger) *OAuthManager {
	return &OAuthManager{
		configs: make(map[Provider]*oauth2.Config),
		storage: storage,
		logger:  logger,
	}
}

func (m *OAuthManager) RegisterProvider(provider Provider, clientID, clientSecret, redirectURL string) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}

	switch provider {
	case Google:
		config.Endpoint = google.Endpoint
	// Add cases for other providers here
	default:
		m.logger.Errorf("Unsupported OAuth provider: %s", provider)
		return
	}

	m.configs[provider] = config
}

func (m *OAuthManager) GetAuthURL(provider Provider, state string) (string, error) {
	config, ok := m.configs[provider]
	if !ok {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
	return config.AuthCodeURL(state), nil
}

func (m *OAuthManager) HandleCallback(ctx context.Context, provider Provider, code string) (*User, error) {
	config, ok := m.configs[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	m.logger.WithField("code", code).Info("Handling OAuth callback")
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	client := config.Client(ctx, token)
	userInfo, err := m.getUserInfo(ctx, client, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from provider: %w", err)
	}

	user, err := m.storage.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			user = &User{
				Name:           userInfo.Name,
				Email:          userInfo.Email,
				AuthProvider:   provider,
				AuthProviderID: userInfo.ID,
			}
			err = m.storage.SaveUser(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("failed to save user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get user from storage: %w", err)
		}
	}

	oauthToken := &OAuthToken{
		UserID:       user.ID,
		Provider:     provider,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
	}
	err = m.storage.SaveOAuthToken(ctx, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to save OAuth token: %w", err)
	}

	return user, nil
}

func (m *OAuthManager) getUserInfo(ctx context.Context, client *http.Client, provider Provider) (*UserInfo, error) {
	switch provider {
	case Google:
		return getGoogleUserInfo(ctx, client)
	// Add cases for other providers here
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

type UserInfo struct {
	ID    string
	Name  string
	Email string
}

func getGoogleUserInfo(ctx context.Context, client *http.Client) (*UserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Sub   string `json:"sub"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:    userInfo.Sub,
		Name:  userInfo.Name,
		Email: userInfo.Email,
	}, nil
}
