// Package system provides the system information and health check endpoints.
package system

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// App represents the application name
type App struct {
	Name string `json:"name"`
}

// Settings represents the application settings
type Settings struct {
	AuthEnabled    bool     `json:"auth_enabled"`
	OAuthProviders []string `json:"oauth_providers"`
}

// Info represents the system information
type Info struct {
	Version  string   `json:"version"`
	App      App      `json:"app"`
	Settings Settings `json:"settings"`
}

// Manager represents the system manager
type Manager struct {
	logger     *logrus.Logger
	storage    storage.Storage
	appVersion string
	systemInfo Info
}

// NewManager creates a new system manager
func NewManager(logger *logrus.Logger, systemInfo Info) *Manager {
	return &Manager{
		logger:     logger,
		systemInfo: systemInfo,
	}
}

// RegisterRoutes registers the search routes
func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/system", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			util.SendSuccessResponse(w, http.StatusOK, types.GenerateResponseMsg{Message: "Pong"}, m.logger, nil)
		})

		r.Route("/probes", func(r chi.Router) {
			r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
				util.SendSuccessResponse(w, http.StatusOK, types.GenerateResponseMsg{Message: "I am alive"}, m.logger, nil)
			})

			r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
				util.SendSuccessResponse(w, http.StatusOK, types.GenerateResponseMsg{Message: "I am ready"}, m.logger, nil)
			})
		})

		r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
			util.SendSuccessResponse(w, http.StatusOK, m.systemInfo, m.logger, nil)
		})
	})
}
