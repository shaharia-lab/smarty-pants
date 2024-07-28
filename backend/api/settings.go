package api

import (
	"encoding/json"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

func updateSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.updateSettingsHandler")
		defer span.End()

		var settings types.Settings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			SendErrorResponse(w, http.StatusBadRequest, "Failed to decode request body", logging, nil)
			return
		}

		err := st.UpdateSettings(ctx, settings)
		if err != nil {
			logging.WithError(err).Error("Failed to update settings")
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to update settings", logging, nil)
			return
		}

		SendSuccessResponse(w, http.StatusOK, settings, logging, nil)
	}
}

func getSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.getSettingsHandler")
		defer span.End()

		settingsDB, err := st.GetSettings(ctx)
		if err != nil {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch settings", logging, nil)
			return
		}

		SendSuccessResponse(w, http.StatusOK, settingsDB, logging, nil)
	}
}
