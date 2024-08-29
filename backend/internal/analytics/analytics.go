// Package analytics provides the analytics API
package analytics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// Analytics represents the analytics API
type Analytics struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

// NewManager creates a new Analytics API with the given storage, logger and ACLManager
func NewManager(storage storage.Storage, logger *logrus.Logger, aclManager auth.ACLManager) *Analytics {
	return &Analytics{
		storage:    storage,
		logger:     logger,
		aclManager: aclManager,
	}
}

// RegisterRoutes registers the API handlers with the given ServeMux
func (a *Analytics) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/analytics", func(r chi.Router) {
		r.Get("/overview", a.getAnalyticsOverview)
	})
}

// getAnalyticsOverview returns the analytics overview
func (a *Analytics) getAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	routeCtx := r.Context()
	if !a.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "analytics") {
		return
	}

	ctx, span := observability.StartSpan(routeCtx, "api.getAnalyticsOverview")
	defer span.End()

	overview, err := a.storage.GetAnalyticsOverview(ctx)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get analytics overview from storage")

		span.RecordError(err)
		span.SetStatus(http.StatusInternalServerError, "failed to get analytics overview")

		util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to get analytics overview", err))
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, overview, a.logger, nil)
}
