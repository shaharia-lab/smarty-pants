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

type EmbeddingManager struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

func NewEmbeddingManager(s storage.Storage, l *logrus.Logger, aclManager auth.ACLManager) *EmbeddingManager {
	return &EmbeddingManager{
		storage:    s,
		logger:     l,
		aclManager: aclManager,
	}
}

func (h *EmbeddingManager) RegisterRoutes(r chi.Router) {
	r.Route("/embedding-provider", func(r chi.Router) {
		r.Post("/", h.addProviderHandler)
		r.Route("/{uuid}", func(r chi.Router) {
			r.Delete("/", h.deleteProviderHandler)
			r.Get("/", h.getProviderHandler)
			r.Put("/", h.updateProviderHandler)
			r.Put("/activate", h.setActiveProviderHandler)
			r.Put("/deactivate", h.setDeactivateProviderHandler)
		})
		r.Get("/", h.GetEmbeddingProviders)
	})
}

func (h *EmbeddingManager) addProviderHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *EmbeddingManager) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.Error(types.InvalidUUIDMessage, "error", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return
	}

	var provider types.EmbeddingProviderConfig
	err = json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Failed to decode request body", Err: err.Error()})
		return
	}
	provider.UUID = id

	err = h.storage.UpdateEmbeddingProvider(r.Context(), provider)
	if err != nil {
		h.logger.Error("Failed to update embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to update embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, h.logger, nil)
}

func (h *EmbeddingManager) deleteProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.Error(types.InvalidUUIDMessage, "error", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return
	}

	err = h.storage.DeleteEmbeddingProvider(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to delete embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusNoContent, nil, h.logger, nil)
}

func (h *EmbeddingManager) getProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return
	}

	provider, err := h.storage.GetEmbeddingProvider(r.Context(), id)
	if err != nil {
		if errors.Is(err, types.ErrEmbeddingProviderNotFound) {
			util.SendAPIErrorResponse(w, http.StatusNotFound, &util.APIError{Message: "Embedding provider not found", Err: err.Error()})
			return
		}

		h.logger.WithError(err).Error("Failed to get embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to get embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, h.logger, nil)
}

func (h *EmbeddingManager) GetEmbeddingProviders(w http.ResponseWriter, r *http.Request) {
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

func (h *EmbeddingManager) setActiveProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return
	}

	err = h.storage.SetActiveEmbeddingProvider(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to set active embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: "Failed to set active embedding provider",
			Err:     err.Error(),
		})
		return
	}

	h.logger.WithField("embedding_provider_id", id).Info("Embedding provider activated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Embedding provider activated successfully"}, h.logger, nil)
}

func (h *EmbeddingManager) setDeactivateProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return
	}

	err = h.storage.SetDisableEmbeddingProvider(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to deactivate embedding provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: "Failed to deactivate embedding provider",
			Err:     err.Error(),
		})
		return
	}

	h.logger.WithField("datasource_id", id).Info("Embedding provider has been deactivated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Embedding provider has been deactivated successfully"}, h.logger, nil)
}
