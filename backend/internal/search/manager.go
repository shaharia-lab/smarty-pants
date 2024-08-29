package search

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	searchSystem System
	logger       *logrus.Logger
}

func NewManager(searchSystem System, logger *logrus.Logger) *Manager {
	return &Manager{
		searchSystem: searchSystem,
		logger:       logger,
	}
}

func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/search", func(r chi.Router) {
		r.Post("/", AddSearchHandler(m.searchSystem, m.logger))
	})
}

func AddSearchHandler(searchSystem System, logging *logrus.Logger) http.HandlerFunc {
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
