package llm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/datasource"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	storage storage.Storage
	logger  *logrus.Logger
}

func NewManager(storage storage.Storage, logger *logrus.Logger) *Manager {
	return &Manager{
		storage: storage,
		logger:  logger,
	}
}

func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/llm-provider", func(r chi.Router) {
		r.Post("/", AddLLMProviderHandler(m.storage, m.logger))
		r.Route("/{uuid}", func(r chi.Router) {
			r.Delete("/", DeleteLLMProviderHandler(m.storage, m.logger))
			r.Get("/", GetLLMProviderHandler(m.storage, m.logger))
			r.Put("/", UpdateLLMProviderHandler(m.storage, m.logger))
			r.Put("/activate", SetActiveLLMProviderHandler(m.storage, m.logger))
			r.Put("/deactivate", SetDisableLLMProviderHandler(m.storage, m.logger))
		})
		r.Get("/", GetLLMProvidersHandler(m.storage, m.logger))
	})
}

func AddLLMProviderHandler(s storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var provider types.LLMProviderConfig
		err := json.NewDecoder(r.Body).Decode(&provider)
		if err != nil {
			var validationError *types.ValidationError
			if errors.As(err, &validationError) {
				datasource.SendJSONError(w, validationError.Error(), http.StatusBadRequest)
			} else {
				datasource.SendJSONError(w, "Invalid request body", http.StatusBadRequest)
			}
			return
		}

		if provider.UUID == uuid.Nil {
			provider.UUID = uuid.New()
		}

		provider.Status = string(types.LLMProviderStatusInactive)

		err = s.CreateLLMProvider(r.Context(), provider)
		if err != nil {
			datasource.SendJSONError(w, "Failed to create embedding provider: "+err.Error(), http.StatusInternalServerError)
			return
		}

		util.SendSuccessResponse(w, http.StatusCreated, provider, logging, nil)
	}
}

func UpdateLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error(types.InvalidUUIDMessage, "error", err)
			http.Error(w, types.InvalidUUIDMessage, http.StatusBadRequest)
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

func DeleteLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.Error(types.InvalidUUIDMessage, "error", err)
			http.Error(w, types.InvalidUUIDMessage, http.StatusBadRequest)
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

func GetLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			datasource.SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		provider, err := s.GetLLMProvider(r.Context(), id)
		if err != nil {
			if errors.Is(err, types.ErrLLMProviderNotFound) {
				datasource.SendJSONError(w, "Embedding provider not found", http.StatusNotFound)
			} else {
				l.WithError(err).Error("Failed to get embedding provider")
				datasource.SendJSONError(w, "Failed to get embedding provider", http.StatusInternalServerError)
			}
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, provider, l, nil)
	}
}

func GetLLMProvidersHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
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
			datasource.SendJSONError(w, "Failed to get embedding providers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if len(providers.LLMProviders) == 0 {
			providers.LLMProviders = []types.LLMProviderConfig{}
		}

		util.SendSuccessResponse(w, http.StatusOK, providers, l, nil)
	}
}

func SetActiveLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.WithError(err).Error(types.InvalidUUIDMessage)
			datasource.SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		err = s.SetActiveLLMProvider(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to set active LLM provider")
			datasource.SendJSONError(w, "Failed to set active LLM provider", http.StatusInternalServerError)
			return
		}

		l.WithField("llm_provider_id", id).Info("LLM provider activated successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "LLM provider activated successfully"})
	}
}

func SetDisableLLMProviderHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil {
			l.WithError(err).Error(types.InvalidUUIDMessage)
			datasource.SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		err = s.SetDisableLLMProvider(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to set deactivate LLM provider")
			datasource.SendJSONError(w, "Failed to set deactivate LLM provider", http.StatusInternalServerError)
			return
		}

		l.WithField("llm_provider_id", id).Info("LLM provider has been deactivated successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "LLM provider has been deactivated successfully"})
	}
}
