// Package embedding provides the embedding provider management functionalities
package embedding

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

// Manager is the embedding provider manager
type Manager struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

// NewEmbeddingManager creates a new embedding manager
func NewEmbeddingManager(s storage.Storage, l *logrus.Logger, aclManager auth.ACLManager) *Manager {
	return &Manager{
		storage:    s,
		logger:     l,
		aclManager: aclManager,
	}
}

// RegisterRoutes registers the embedding provider routes
func (h *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/embedding-provider", func(r chi.Router) {
		r.Post("/", h.addProviderHandler)
		r.Route("/{uuid}", func(r chi.Router) {
			r.Delete("/", h.deleteProviderHandler)
			r.Get("/", h.getProviderHandler)
			r.Put("/", h.updateProviderHandler)
			r.Put("/activate", h.setActiveProviderHandler)
			r.Put("/deactivate", h.setDeactivateProviderHandler)
		})
		r.Get("/", h.getEmbeddingProviders)
	})
}

func (h *Manager) addProviderHandler(w http.ResponseWriter, r *http.Request) {
	var provider types.EmbeddingProviderConfig
	err := json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		var validationError *types.ValidationError
		if errors.As(err, &validationError) {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Validation failed to add embedding provider", Err: err.Error()})
			return
		}

		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Invalid request body", Err: err.Error()})
		return
	}

	if provider.UUID == uuid.Nil {
		provider.UUID = uuid.New()
	}

	err = h.storage.CreateEmbeddingProvider(r.Context(), provider)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to create embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusCreated, provider, h.logger, nil)
}

func (h *Manager) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !h.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsEmbeddingProviderUpdate, nil) {
		return
	}

	resolvedProvider, err := h.getEmbeddingProviderFromRequest(w, r)
	if err != nil {
		return
	}

	var provider types.EmbeddingProviderConfig
	err = json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Failed to decode request body", Err: err.Error()})
		return
	}

	provider.UUID = resolvedProvider.UUID

	err = h.storage.UpdateEmbeddingProvider(r.Context(), provider)
	if err != nil {
		h.logger.Error("Failed to update embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to update embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, h.logger, nil)
}

func (h *Manager) deleteProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !h.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsEmbeddingProviderDelete, nil) {
		return
	}

	provider, err := h.getEmbeddingProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = h.storage.DeleteEmbeddingProvider(r.Context(), provider.UUID)
	if err != nil {
		h.logger.Error("Failed to delete embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to delete embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusNoContent, nil, h.logger, nil)
}

func (h *Manager) getProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !h.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsEmbeddingProviderGet, nil) {
		return
	}

	provider, err := h.getEmbeddingProviderFromRequest(w, r)
	if err != nil {
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, h.logger, nil)
}

func (h *Manager) getEmbeddingProviderFromRequest(w http.ResponseWriter, r *http.Request) (*types.EmbeddingProviderConfig, error) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return nil, err
	}

	provider, err := h.storage.GetEmbeddingProvider(r.Context(), id)
	if err != nil {
		if errors.Is(err, types.ErrEmbeddingProviderNotFound) {
			util.SendAPIErrorResponse(w, http.StatusNotFound, &util.APIError{Message: "Embedding provider not found", Err: err.Error()})
			return nil, err
		}

		h.logger.WithError(err).Error("Failed to get embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to get embedding provider", Err: err.Error()})
		return nil, err
	}
	return provider, nil
}

func (h *Manager) getEmbeddingProviders(w http.ResponseWriter, r *http.Request) {
	var filter types.EmbeddingProviderFilter
	var option types.EmbeddingProviderFilterOption

	filter.Status = r.URL.Query().Get("status")

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	option.Page = page

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage < 1 {
		perPage = 10
	}
	option.Limit = perPage

	providers, err := h.storage.GetAllEmbeddingProviders(r.Context(), filter, option)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get embedding providers")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to get embedding providers", Err: err.Error()})
		return
	}

	if len(providers.EmbeddingProviders) == 0 {
		providers.EmbeddingProviders = []types.EmbeddingProviderConfig{}
	}

	util.SendSuccessResponse(w, http.StatusOK, providers, h.logger, nil)
}

func (h *Manager) setActiveProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !h.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsEmbeddingProviderActivate, nil) {
		return
	}

	provider, err := h.getEmbeddingProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = h.storage.SetActiveEmbeddingProvider(r.Context(), provider.UUID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to set active embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: "Failed to set active embedding provider",
			Err:     err.Error(),
		})
		return
	}

	h.logger.WithField("embedding_provider_id", provider.UUID).Info("Embedding provider activated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Embedding provider activated successfully"}, h.logger, nil)
}

func (h *Manager) setDeactivateProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !h.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsEmbeddingProviderActivate, nil) {
		return
	}

	provider, err := h.getEmbeddingProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = h.storage.SetDisableEmbeddingProvider(r.Context(), provider.UUID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to deactivate embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: "Failed to deactivate embedding provider",
			Err:     err.Error(),
		})
		return
	}

	h.logger.WithField("datasource_id", provider.UUID).Info("Embedding provider has been deactivated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Embedding provider has been deactivated successfully"}, h.logger, nil)
}
