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
			sendJSONError(w, validationError.Error(), http.StatusBadRequest)
		} else {
			sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		}
		return
	}

	if provider.UUID == uuid.Nil {
		provider.UUID = uuid.New()
	}

	err = h.storage.CreateEmbeddingProvider(r.Context(), provider)
	if err != nil {
		sendJSONError(w, "Failed to create embedding provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	util.SendSuccessResponse(w, http.StatusCreated, provider, h.logger, nil)
}

func (h *EmbeddingManager) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.Error(types.InvalidUUIDMessage, "error", err)
		http.Error(w, types.InvalidUUIDMessage, http.StatusBadRequest)
		return
	}

	var provider types.EmbeddingProviderConfig
	err = json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	provider.UUID = id

	err = h.storage.UpdateEmbeddingProvider(r.Context(), provider)
	if err != nil {
		h.logger.Error("Failed to update embedding provider", "error", err)
		http.Error(w, "Failed to update embedding provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(provider)
}

func (h *EmbeddingManager) deleteProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		h.logger.Error(types.InvalidUUIDMessage, "error", err)
		http.Error(w, types.InvalidUUIDMessage, http.StatusBadRequest)
		return
	}

	err = h.storage.DeleteEmbeddingProvider(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete embedding provider", "error", err)
		http.Error(w, "Failed to delete embedding provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EmbeddingManager) getProviderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		sendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
		return
	}

	provider, err := h.storage.GetEmbeddingProvider(r.Context(), id)
	if err != nil {
		if errors.Is(err, types.ErrEmbeddingProviderNotFound) {
			sendJSONError(w, "Embedding provider not found", http.StatusNotFound)
		} else {
			h.logger.WithError(err).Error("Failed to get embedding provider")
			sendJSONError(w, "Failed to get embedding provider", http.StatusInternalServerError)
		}
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
		sendJSONError(w, "Failed to get embedding providers: "+err.Error(), http.StatusInternalServerError)
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
		sendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
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
		sendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
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

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
