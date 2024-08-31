// Package search provides a search system for the application.
package search

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// Manager represents the search manager
type Manager struct {
	searchSystem System
	logger       *logrus.Logger
	aclManager   auth.ACLManager
}

// NewManager creates a new search manager
func NewManager(searchSystem System, logger *logrus.Logger, aclManager auth.ACLManager) *Manager {
	return &Manager{
		searchSystem: searchSystem,
		logger:       logger,
		aclManager:   aclManager,
	}
}

// RegisterRoutes registers the search routes
func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/search", func(r chi.Router) {
		r.Post("/", m.addSearchHandler)
	})
}

func (m *Manager) addSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, span := observability.StartSpan(ctx, "api.addSearchHandler")
	defer span.End()

	if !m.aclManager.IsAllowed(w, r, types.UserRoleUser, types.APiAccessOpsSearchSearch, nil) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		m.logger.Error("Failed to read request body: ", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Failed to read request body", Err: err.Error()})
		return
	}

	var searchReq Request
	if err := json.Unmarshal(body, &searchReq); err != nil {
		m.logger.Error("Failed to unmarshal request body: ", err)
		util.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body", m.logger, nil)
		return
	}

	results, err := m.searchSystem.SearchDocument(ctx, searchReq)
	if err != nil {
		m.logger.WithError(err).Error("Failed to search document")
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to search document", m.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, results, m.logger, nil)
}
