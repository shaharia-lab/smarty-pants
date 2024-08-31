// Package settings provides the settings API
package settings

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// Manager represents the settings API
type Manager struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

// NewManager creates a new settings API with the given storage and logger
func NewManager(storage storage.Storage, logger *logrus.Logger, aclManager auth.ACLManager) *Manager {
	return &Manager{
		storage:    storage,
		logger:     logger,
		aclManager: aclManager,
	}
}

// RegisterRoutes registers the API handlers with the given ServeMux
func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/settings", func(r chi.Router) {
		r.Put("/", m.updateSettingsHandler)
		r.Get("/", m.getSettingsHandler)
	})
}

func (m *Manager) updateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.StartSpan(r.Context(), "api.updateSettingsHandler")
	defer span.End()

	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APiAccessOpsSettingsUpdate, nil) {
		return
	}

	var settings types.Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		util.SendErrorResponse(w, http.StatusBadRequest, "Failed to decode request body", m.logger, nil)
		return
	}

	err := m.storage.UpdateSettings(ctx, settings)
	if err != nil {
		m.logger.WithError(err).Error("Failed to update settings")
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update settings", m.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, settings, m.logger, nil)
}

func (m *Manager) getSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := observability.StartSpan(r.Context(), "api.getSettingsHandler")
	defer span.End()

	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APiAccessOpsSettingsGet, nil) {
		return
	}

	settingsDB, err := m.storage.GetSettings(ctx)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch settings", m.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, settingsDB, m.logger, nil)
}
