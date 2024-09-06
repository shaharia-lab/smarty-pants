package llm

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

type Manager struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

func NewManager(storage storage.Storage, logger *logrus.Logger, aclManager auth.ACLManager) *Manager {
	return &Manager{
		storage:    storage,
		logger:     logger,
		aclManager: aclManager,
	}
}

func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/llm-provider", func(r chi.Router) {
		r.Post("/", m.addLLMProviderHandler)
		r.Route("/{uuid}", func(r chi.Router) {
			r.Delete("/", m.deleteLLMProviderHandler)
			r.Get("/", m.getLLMProviderHandler)
			r.Put("/", m.updateLLMProviderHandler)
			r.Put("/activate", m.setActiveLLMProviderHandler)
			r.Put("/deactivate", m.setDisableLLMProviderHandler)
		})
		r.Get("/", m.getLLMProvidersHandler)
	})
}

func (m *Manager) addLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	var provider types.LLMProviderConfig
	err := json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		var validationError *types.ValidationError
		if errors.As(err, &validationError) {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: validationError.Error(), Err: validationError.Error()})
			return
		}

		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: "Invalid request body", Err: err.Error()})
		return
	}

	if provider.UUID == uuid.Nil {
		provider.UUID = uuid.New()
	}

	provider.Status = string(types.LLMProviderStatusInactive)

	err = m.storage.CreateLLMProvider(r.Context(), provider)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to create embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusCreated, provider, m.logger, nil)
}

func (m *Manager) updateLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProviderUpdate, nil) {
		return
	}

	provider, err := m.getLLMProviderFromRequest(w, r)
	if err != nil {
		return
	}

	var providerToUpdate types.LLMProviderConfig
	err = json.NewDecoder(r.Body).Decode(&provider)
	if err != nil {
		m.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	providerToUpdate.UUID = provider.UUID

	err = m.storage.UpdateLLMProvider(r.Context(), *provider)
	if err != nil {
		m.logger.Error("Failed to update embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to update embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, m.logger, nil)
}

func (m *Manager) getLLMProviderFromRequest(w http.ResponseWriter, r *http.Request) (*types.LLMProviderConfig, error) {
	providerUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err == nil && providerUUID == uuid.Nil {
		err = errors.New(types.InvalidUUIDMessage)
	}

	if err != nil {
		m.logger.WithError(err).Error("failed to parse provider UUID")
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{Message: types.InvalidUUIDMessage, Err: err.Error()})
		return nil, err
	}

	provider, err := m.storage.GetLLMProvider(r.Context(), providerUUID)
	if err != nil {
		if errors.Is(err, types.ErrLLMProviderNotFound) {
			util.SendAPIErrorResponse(w, http.StatusNotFound, &util.APIError{Message: "Embedding provider not found", Err: err.Error()})
			return nil, err
		}

		m.logger.Error("Failed to get embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to get embedding provider", Err: err.Error()})
		return nil, err
	}

	return provider, nil
}

func (m *Manager) deleteLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProviderDelete, nil) {
		return
	}

	provider, err := m.getLLMProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = m.storage.DeleteLLMProvider(r.Context(), provider.UUID)
	if err != nil {
		m.logger.Error("Failed to delete embedding provider", "error", err)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to delete embedding provider", Err: err.Error()})
		return
	}

	util.SendSuccessResponse(w, http.StatusNoContent, nil, m.logger, nil)
}

func (m *Manager) getLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProviderGet, nil) {
		return
	}

	provider, err := m.getLLMProviderFromRequest(w, r)
	if err != nil {
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, provider, m.logger, nil)
}

func (m *Manager) getLLMProvidersHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProvidersGet, nil) {
		return
	}

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

	providers, err := m.storage.GetAllLLMProviders(r.Context(), filter, option)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get embedding providers")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to get embedding providers", Err: err.Error()})
		return
	}

	if len(providers.LLMProviders) == 0 {
		providers.LLMProviders = []types.LLMProviderConfig{}
	}

	util.SendSuccessResponse(w, http.StatusOK, providers, m.logger, nil)
}

func (m *Manager) setActiveLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProviderActivate, nil) {
		return
	}

	provider, err := m.getLLMProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = m.storage.SetActiveLLMProvider(r.Context(), provider.UUID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to set active LLM provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to set active LLM provider", Err: err.Error()})
		return
	}

	m.logger.WithField("llm_provider_id", provider.UUID).Info("LLM provider activated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "LLM provider activated successfully"}, m.logger, nil)
}

func (m *Manager) setDisableLLMProviderHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleAdmin, types.APIAccessOpsLLMProviderDeactivate, nil) {
		return
	}

	provider, err := m.getLLMProviderFromRequest(w, r)
	if err != nil {
		return
	}

	err = m.storage.SetDisableLLMProvider(r.Context(), provider.UUID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to set deactivate LLM provider")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{Message: "Failed to set deactivate LLM provider", Err: err.Error()})
		return
	}

	m.logger.WithField("llm_provider_id", provider.UUID).Info("LLM provider has been deactivated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "LLM provider has been deactivated successfully"}, m.logger, nil)
}
