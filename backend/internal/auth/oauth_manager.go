package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
)

// OAuthProvider is an interface for OAuth providers
type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCodeForToken(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, token string) (*UserInfo, error)
}

// OAuthManager handles OAuth operations
type OAuthManager struct {
	providers   map[string]OAuthProvider
	userManager *UserManager
	jwtManager  *JWTManager
	logger      *logrus.Logger
	stateStore  map[string]stateInfo
}

// stateInfo stores the provider and timestamp for a given state
type stateInfo struct {
	provider  string
	timestamp time.Time
}

// NewOAuthManager creates a new OAuthManager with the given providers, UserManager, JWTManager, and logger
func NewOAuthManager(providers map[string]OAuthProvider, userManager *UserManager, jwtManager *JWTManager, logger *logrus.Logger) *OAuthManager {
	return &OAuthManager{
		providers:   providers,
		userManager: userManager,
		jwtManager:  jwtManager,
		logger:      logger,
		stateStore:  make(map[string]stateInfo),
	}
}

// AuthFlowRequest represents the request body for initiating an OAuth flow
type AuthFlowRequest struct {
	AuthFlow struct {
		Provider   string `json:"provider"`
		CurrentURL string `json:"current_url,omitempty"`
	} `json:"auth_flow"`
}

// AuthFlowResponse represents the response body for initiating an OAuth flow
type AuthFlowResponse struct {
	AuthFlow struct {
		Provider        string `json:"provider"`
		AuthRedirectURL string `json:"auth_redirect_url"`
		State           string `json:"state"`
	} `json:"auth_flow"`
}

// AuthCodeRequest represents the request body for handling an OAuth code
type AuthCodeRequest struct {
	AuthFlow struct {
		Provider string `json:"provider"`
		AuthCode string `json:"auth_code"`
		State    string `json:"state"`
	} `json:"auth_flow"`
}

// AuthTokenResponse represents the response body for handling an OAuth code
type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// InitiateAuthFlow initiates an OAuth flow with the given provider
func (om *OAuthManager) InitiateAuthFlow(w http.ResponseWriter, r *http.Request) {
	var req AuthFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		om.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	provider, ok := om.providers[req.AuthFlow.Provider]
	if !ok {
		om.logger.WithField("provider", req.AuthFlow.Provider).Error("Unsupported provider")
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	state := uuid.New().String()
	om.stateStore[state] = stateInfo{
		provider:  req.AuthFlow.Provider,
		timestamp: time.Now(),
	}

	authURL := provider.GetAuthURL(state)

	resp := AuthFlowResponse{
		AuthFlow: struct {
			Provider        string `json:"provider"`
			AuthRedirectURL string `json:"auth_redirect_url"`
			State           string `json:"state"`
		}{
			Provider:        req.AuthFlow.Provider,
			AuthRedirectURL: authURL,
			State:           state,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleAuthCode handles an OAuth code and returns an access token
func (om *OAuthManager) HandleAuthCode(w http.ResponseWriter, r *http.Request) {
	var req AuthCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		om.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stateInfo, ok := om.stateStore[req.AuthFlow.State]
	if !ok {
		om.logger.WithField("state", req.AuthFlow.State).Error("Invalid state")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	delete(om.stateStore, req.AuthFlow.State)

	provider, ok := om.providers[stateInfo.provider]
	if !ok {
		om.logger.WithField("provider", stateInfo.provider).Error("Invalid provider")
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	token, err := provider.ExchangeCodeForToken(r.Context(), req.AuthFlow.AuthCode)
	if err != nil {
		om.logger.WithError(err).Error("Failed to exchange OAuth code")
		http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		return
	}

	userInfo, err := provider.GetUserInfo(r.Context(), token)
	if err != nil {
		om.logger.WithError(err).Error("Failed to get user info")
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	user, err := om.userManager.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		if errors.Is(err, types.UserNotFoundError) {
			user, err = om.userManager.CreateUser(r.Context(), userInfo.Name, userInfo.Email, "active")
			if err != nil {
				om.logger.WithError(err).Error("Failed to create user")
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			om.logger.WithError(err).Error("Failed to get user")
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
			return
		}
	}

	jwtToken, err := om.jwtManager.IssueTokenForUser(r.Context(), user.UUID, []string{"user"}, 24*time.Hour)
	if err != nil {
		om.logger.WithError(err).Error("Failed to issue JWT token")
		http.Error(w, "Failed to issue token", http.StatusInternalServerError)
		return
	}

	resp := AuthTokenResponse{
		AccessToken: jwtToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterRoutes registers the OAuth routes
func (om *OAuthManager) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/auth/initiate", om.InitiateAuthFlow)
	r.Post("/api/v1/auth/callback", om.HandleAuthCode)
}

// UserInfo represents the structure of user info from an OAuth provider
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
