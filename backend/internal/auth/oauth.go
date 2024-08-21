package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthManager struct {
	config      *oauth2.Config
	userManager *UserManager
	jwtManager  *JWTManager
	logger      *logrus.Logger
	stateStore  map[string]time.Time // Simple in-memory store for states
}

func NewOAuthManager(clientID, clientSecret, redirectURL string, userManager *UserManager, jwtManager *JWTManager, logger *logrus.Logger) *OAuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthManager{
		config:      config,
		userManager: userManager,
		jwtManager:  jwtManager,
		logger:      logger,
		stateStore:  make(map[string]time.Time),
	}
}

type AuthFlowRequest struct {
	AuthFlow struct {
		Provider   string `json:"provider"`
		CurrentURL string `json:"current_url,omitempty"`
	} `json:"auth_flow"`
}

type AuthFlowResponse struct {
	AuthFlow struct {
		Provider        string `json:"provider"`
		AuthRedirectURL string `json:"auth_redirect_url"`
		State           string `json:"state"`
	} `json:"auth_flow"`
}

type AuthCodeRequest struct {
	AuthFlow struct {
		Provider string `json:"provider"`
		AuthCode string `json:"auth_code"`
		State    string `json:"state"`
	} `json:"auth_flow"`
}

type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (om *OAuthManager) InitiateAuthFlow(w http.ResponseWriter, r *http.Request) {
	var req AuthFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		om.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AuthFlow.Provider != "google" {
		om.logger.WithField("provider", req.AuthFlow.Provider).Error("Unsupported provider")
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	state := uuid.New().String()
	om.stateStore[state] = time.Now() // Store state with timestamp

	authURL := om.config.AuthCodeURL(state)

	resp := AuthFlowResponse{
		AuthFlow: struct {
			Provider        string `json:"provider"`
			AuthRedirectURL string `json:"auth_redirect_url"`
			State           string `json:"state"`
		}{
			Provider:        "google",
			AuthRedirectURL: authURL,
			State:           state,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (om *OAuthManager) HandleAuthCode(w http.ResponseWriter, r *http.Request) {
	var req AuthCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		om.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AuthFlow.Provider != "google" {
		om.logger.WithField("provider", req.AuthFlow.Provider).Error("Unsupported provider")
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	// Validate state
	if _, ok := om.stateStore[req.AuthFlow.State]; !ok {
		om.logger.WithField("state", req.AuthFlow.State).Error("Invalid state")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	delete(om.stateStore, req.AuthFlow.State) // Remove used state

	token, err := om.config.Exchange(r.Context(), req.AuthFlow.AuthCode)
	if err != nil {
		om.logger.WithError(err).Error("Failed to exchange OAuth code")
		http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		return
	}

	userInfo, err := om.getUserInfo(r.Context(), token.AccessToken)
	if err != nil {
		om.logger.WithError(err).Error("Failed to get user info")
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	user, err := om.userManager.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		if errors.Is(err, types.UserNotFoundError) {
			// User doesn't exist, create a new one
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

	// Issue JWT token
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

func (om *OAuthManager) getUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %v", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}

	return &userInfo, nil
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (om *OAuthManager) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/auth/initiate", om.InitiateAuthFlow)
	r.Post("/api/v1/auth/callback", om.HandleAuthCode)
}
