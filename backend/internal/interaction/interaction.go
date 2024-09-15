// Package interaction provides the interaction management functionalities
package interaction

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/llm"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

// InteractionSummary represents a summary of an interaction
type InteractionSummary struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
}

// InteractionsResponse represents a response containing a list of interactions
type InteractionsResponse struct {
	Interactions []InteractionSummary `json:"interactions"`
	Limit        int                  `json:"limit"`
	PerPage      int                  `json:"per_page"`
}

// Manager is the interaction manager
type Manager struct {
	storage      storage.Storage
	logger       *logrus.Logger
	searchSystem search.System
	aclManager   auth.ACLManager
}

// NewManager creates a new interaction manager
func NewManager(storage storage.Storage, logger *logrus.Logger, searchSystem search.System, aclManager auth.ACLManager) *Manager {
	return &Manager{
		storage:      storage,
		logger:       logger,
		searchSystem: searchSystem,
		aclManager:   aclManager,
	}
}

// RegisterRoutes registers the interaction routes
func (m *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/interactions", func(r chi.Router) {
		r.Post("/", m.createInteractionHandler)
		r.Get("/", m.getInteractionsHandler)
		r.Route("/{uuid}", func(r chi.Router) {
			r.Get("/", m.getInteractionHandler)
			r.Post("/message", m.sendMessageHandler)
		})
	})

	r.Post("/message", m.sendMessageHandler)
}

func (m *Manager) createInteractionHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleUser, types.APIAccessOpsInteractionCreate, nil) {
		return
	}

	ctx, span := observability.StartSpan(r.Context(), "api.createInteractionHandler")
	defer span.End()

	var c types.Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	interaction, err := m.storage.CreateInteraction(ctx, types.Interaction{
		UUID:  uuid.New(),
		Query: c.Text,
		Conversations: []types.Conversation{
			{Role: types.InteractionRoleUser, Text: c.Text, UUID: uuid.New(), CreatedAt: c.CreatedAt},
		},
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to create interaction")
		util.SendErrorResponse(w, http.StatusInternalServerError, "failed to create interaction", m.logger, nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, interaction, m.logger, nil)
}

func (m *Manager) getInteractionsHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleUser, types.APIAccessOpsInteractionsGet, nil) {
		return
	}

	response := InteractionsResponse{
		Interactions: []InteractionSummary{
			{UUID: uuid.New().String(), Title: "Sample query 1"},
			{UUID: uuid.New().String(), Title: "Sample query 2"},
		},
		Limit:   1,
		PerPage: 10,
	}

	util.SendSuccessResponse(w, http.StatusOK, response, m.logger, nil)
}

func (m *Manager) getInteractionHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleUser, types.APIAccessOpsInteractionGet, nil) {
		return
	}

	interactionUUID := chi.URLParam(r, "uuid")

	interaction := types.Interaction{
		UUID:  uuid.MustParse(interactionUUID),
		Query: "Sample query",
		Conversations: []types.Conversation{
			{Role: "system", Text: "Hello, how may I help you today?"},
			{Role: "user", Text: "Sample user message"},
			{Role: "system", Text: "Sample system response"},
		},
	}

	util.SendSuccessResponse(w, http.StatusOK, interaction, m.logger, nil)
}

func (m *Manager) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if !m.aclManager.IsAllowed(w, r, types.UserRoleUser, types.APIAccessOpsMessageSend, nil) {
		return
	}

	ctx, span := observability.StartSpan(r.Context(), "api.sendMessageHandler")
	defer span.End()

	var c types.Conversation
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		util.SendErrorResponse(w, http.StatusBadRequest, "invalid request", m.logger, nil)
		return
	}

	llmProvider, err := llm.InitializeLLMProvider(ctx, m.storage, m.logger)
	if err != nil {
		m.logger.WithError(err).Error("Failed to initialize LLM provider")
		util.SendErrorResponse(w, http.StatusInternalServerError, "failed to initialize LLM provider", m.logger, nil)
		return
	}

	llmContexts, err := m.searchSystem.GenerateLLMContext(ctx, search.Request{Query: c.Text})
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate LLM contexts")
		util.SendErrorResponse(w, http.StatusInternalServerError, "failed to generate LLM contexts", m.logger, nil)
		return
	}
	span.SetAttributes(
		attribute.Int("no_of_search_results_for_context", len(llmContexts)),
	)

	promptGenerator := llm.NewPromptGenerator(m.logger, llm.PromptTemplate{Template: llm.DefaultPromptTemplate})
	prompt, err := promptGenerator.GeneratePrompt(llm.PromptTemplateData{
		Query:                 c.Text,
		Documents:             llmContexts,
		ConversationHistories: []types.Conversation{},
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to generate prompt")
		util.SendErrorResponse(w, http.StatusInternalServerError, "failed to generate prompt", m.logger, nil)
		return
	}

	llmResponse, err := llmProvider.GetResponse(prompt)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get response from LLM provider")
		util.SendErrorResponse(w, http.StatusInternalServerError, "failed to get response from LLM provider", m.logger, nil)
		return
	}

	m.logger.Infof("LLM response: %s", llmResponse)
	util.SendSuccessResponse(w, http.StatusOK, types.Conversation{
		Role: "system",
		Text: llmResponse,
	}, m.logger, nil)
}
