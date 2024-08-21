package api

import (
	"encoding/json"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

func updateSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
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

func getSettingsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
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
