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
	storage storage.Storage
	logger  *logrus.Logger
}

// NewManager creates a new settings API with the given storage and logger
func NewManager(storage storage.Storage, logger *logrus.Logger, manager auth.ACLManager) *Manager {
	return &Manager{
		storage: storage,
		logger:  logger,
	}
}

// RegisterRoutes registers the API handlers with the given ServeMux
func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/settings", func(r chi.Router) {
		r.Put("/", UpdateSettingsHandler(m.storage, m.logger))
		r.Get("/", GetSettingsHandler(m.storage, m.logger))
	})
}

// UpdateSettingsHandler updates the settings
func UpdateSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.updateSettingsHandler")
		defer span.End()

		var settings types.Settings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			util.SendErrorResponse(w, http.StatusBadRequest, "Failed to decode request body", logging, nil)
			return
		}

		err := st.UpdateSettings(ctx, settings)
		if err != nil {
			logging.WithError(err).Error("Failed to update settings")
			util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update settings", logging, nil)
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, settings, logging, nil)
	}
}

// GetSettingsHandler returns the settings
func GetSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.getSettingsHandler")
		defer span.End()

		settingsDB, err := st.GetSettings(ctx)
		if err != nil {
			util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch settings", logging, nil)
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, settingsDB, logging, nil)
	}
}
