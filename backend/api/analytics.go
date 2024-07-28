// Package api provides an API for the application.
package api

import (
	"net/http"

	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/storage"
	"github.com/shaharia-lab/smarty-pants-ai/internal/util"
	"github.com/sirupsen/logrus"
)

// getAnalyticsOverview returns a handler function that fetches analytics overview from storage
func getAnalyticsOverview(st storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeCtx := r.Context()

		ctx, span := observability.StartSpan(routeCtx, "api.GetAnalyticsOverview")
		defer span.End()

		overview, err := st.GetAnalyticsOverview(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get analytics overview from storage")

			span.RecordError(err)
			span.SetStatus(http.StatusInternalServerError, "failed to get analytics overview")

			SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to get analytics overview", err))
			return
		}

		SendSuccessResponse(w, http.StatusOK, overview, logger, nil)
	}
}
