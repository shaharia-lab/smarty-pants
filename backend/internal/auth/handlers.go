package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type OAuthHandlers struct {
	oauthManager *OAuthManager
	logger       *logrus.Logger
}

func NewOAuthHandlers(oauthManager *OAuthManager, logger *logrus.Logger) *OAuthHandlers {
	return &OAuthHandlers{
		oauthManager: oauthManager,
		logger:       logger,
	}
}

func (h *OAuthHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/auth/{provider}/login", h.HandleLogin)
	r.Get("/auth/{provider}/callback", h.HandleCallback)
}

func (h *OAuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	provider := Provider(chi.URLParam(r, "provider"))
	state, _ := generateRandomState() // Implement this function to generate a random state

	url, err := h.oauthManager.GetAuthURL(provider, state)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get auth URL")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the state in a secure session or cookie
	// ...

	http.Redirect(w, r, url, http.StatusFound)
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *OAuthHandlers) HandleCallback(w http.ResponseWriter, r *http.Request) {
	provider := Provider(chi.URLParam(r, "provider"))
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	h.logger.WithField("state", state).Info("Need to handle state")

	user, err := h.oauthManager.HandleCallback(r.Context(), provider, code)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle OAuth callback")
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Create a session for the authenticated user
	// ...

	response := map[string]interface{}{
		"message": "Authentication successful",
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
