package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

func addEmbeddingProviderHandler(s storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = s.CreateEmbeddingProvider(r.Context(), provider)
		if err != nil {
			sendJSONError(w, "Failed to create embedding provider: "+err.Error(), http.StatusInternalServerError)
			return
		}

		util.SendSuccessResponse(w, http.StatusCreated, provider, logging, nil)
	}
}

func updateEmbeddingProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error(invalidUUIDMsg, "error", err)
			http.Error(w, invalidUUIDMsg, http.StatusBadRequest)
			return
		}

		var provider types.EmbeddingProviderConfig
		err = json.NewDecoder(r.Body).Decode(&provider)
		if err != nil {
			l.Error("Failed to decode request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		provider.UUID = id

		err = s.UpdateEmbeddingProvider(r.Context(), provider)
		if err != nil {
			l.Error("Failed to update embedding provider", "error", err)
			http.Error(w, "Failed to update embedding provider", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(provider)
	}
}

func deleteEmbeddingProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error(invalidUUIDMsg, "error", err)
			http.Error(w, invalidUUIDMsg, http.StatusBadRequest)
			return
		}

		err = s.DeleteEmbeddingProvider(r.Context(), id)
		if err != nil {
			l.Error("Failed to delete embedding provider", "error", err)
			http.Error(w, "Failed to delete embedding provider", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getEmbeddingProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			sendJSONError(w, invalidUUIDMsg, http.StatusBadRequest)
			return
		}

		provider, err := s.GetEmbeddingProvider(r.Context(), id)
		if err != nil {
			if errors.Is(err, types.ErrEmbeddingProviderNotFound) {
				sendJSONError(w, "Embedding provider not found", http.StatusNotFound)
			} else {
				l.WithError(err).Error("Failed to get embedding provider")
				sendJSONError(w, "Failed to get embedding provider", http.StatusInternalServerError)
			}
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, provider, l, nil)
	}
}

func getEmbeddingProvidersHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		providers, err := s.GetAllEmbeddingProviders(r.Context(), filter, option)
		if err != nil {
			l.WithError(err).Error("Failed to get embedding providers")
			sendJSONError(w, "Failed to get embedding providers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if len(providers.EmbeddingProviders) == 0 {
			providers.EmbeddingProviders = []types.EmbeddingProviderConfig{}
		}

		util.SendSuccessResponse(w, http.StatusOK, providers, l, nil)
	}
}

func setActiveEmbeddingProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.WithError(err).Error(invalidUUIDMsg)
			sendJSONError(w, invalidUUIDMsg, http.StatusBadRequest)
			return
		}

		err = s.SetActiveEmbeddingProvider(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to set active embedding provider")
			util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
				Message: "Failed to set active embedding provider",
				Err:     err.Error(),
			})
			return
		}

		l.WithField("embedding_provider_id", id).Info("Embedding provider activated successfully")
		util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Embedding provider activated successfully"}, l, nil)
	}
}
