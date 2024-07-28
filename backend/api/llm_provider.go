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
	"github.com/sirupsen/logrus"
)

func addLLMProviderHandler(s storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var provider types.LLMProviderConfig
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

		provider.Status = string(types.LLMProviderStatusInactive)

		err = s.CreateLLMProvider(r.Context(), provider)
		if err != nil {
			sendJSONError(w, "Failed to create embedding provider: "+err.Error(), http.StatusInternalServerError)
			return
		}

		SendSuccessResponse(w, http.StatusCreated, provider, logging, nil)
	}
}

func updateLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error("Invalid UUID", "error", err)
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		var provider types.LLMProviderConfig
		err = json.NewDecoder(r.Body).Decode(&provider)
		if err != nil {
			l.Error("Failed to decode request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		provider.UUID = id

		err = s.UpdateLLMProvider(r.Context(), provider)
		if err != nil {
			l.Error("Failed to update embedding provider", "error", err)
			http.Error(w, "Failed to update embedding provider", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(provider)
	}
}

func deleteLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error("Invalid UUID", "error", err)
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		err = s.DeleteLLMProvider(r.Context(), id)
		if err != nil {
			l.Error("Failed to delete embedding provider", "error", err)
			http.Error(w, "Failed to delete embedding provider", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			sendJSONError(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		provider, err := s.GetLLMProvider(r.Context(), id)
		if err != nil {
			if errors.Is(err, types.ErrLLMProviderNotFound) {
				sendJSONError(w, "Embedding provider not found", http.StatusNotFound)
			} else {
				l.WithError(err).Error("Failed to get embedding provider")
				sendJSONError(w, "Failed to get embedding provider", http.StatusInternalServerError)
			}
			return
		}

		SendSuccessResponse(w, http.StatusOK, provider, l, nil)
	}
}

func getLLMProvidersHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var filter types.LLMProviderFilter
		var option types.LLMProviderFilterOption

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

		providers, err := s.GetAllLLMProviders(r.Context(), filter, option)
		if err != nil {
			l.WithError(err).Error("Failed to get embedding providers")
			sendJSONError(w, "Failed to get embedding providers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if len(providers.LLMProviders) == 0 {
			providers.LLMProviders = []types.LLMProviderConfig{}
		}

		SendSuccessResponse(w, http.StatusOK, providers, l, nil)
	}
}

func setActiveLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.WithError(err).Error("Invalid UUID")
			sendJSONError(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		err = s.SetActiveLLMProvider(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to set active LLM provider")
			sendJSONError(w, "Failed to set active LLM provider", http.StatusInternalServerError)
			return
		}

		l.WithField("llm_provider_id", id).Info("LLM provider activated successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "LLM provider activated successfully"})
	}
}

func setDisableLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.WithError(err).Error("Invalid UUID")
			sendJSONError(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		err = s.SetDisableLLMProvider(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to set deactivate LLM provider")
			sendJSONError(w, "Failed to set deactivate LLM provider", http.StatusInternalServerError)
			return
		}

		l.WithField("llm_provider_id", id).Info("LLM provider has been deactivated successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "LLM provider has been deactivated successfully"})
	}
}
