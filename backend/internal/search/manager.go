// Package search provides a search system for the application.
package search

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
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
		r.Post("/", addSearchHandler(m.searchSystem, m.logger))
	})
}

func addSearchHandler(searchSystem System, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logging.Error("Failed to read request body: ", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var searchReq Request
		if err := json.Unmarshal(body, &searchReq); err != nil {
			logging.Error("Failed to unmarshal request body: ", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		results, err := searchSystem.SearchDocument(ctx, searchReq)
		if err != nil {
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, results, logging, nil)
	}
}
