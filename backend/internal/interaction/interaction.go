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

// MessageRequest represents a request to send a message
type MessageRequest struct {
	Query string `json:"query"`
}

// MessageResponse represents a response containing a message
type MessageResponse struct {
	Response string `json:"response"`
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
		r.Post("/", createInteractionHandler(m.storage, m.logger))
		r.Get("/", getInteractionsHandler(m.logger))
		r.Route("/{uuid}", func(r chi.Router) {
			r.Get("/", getInteractionHandler(m.logger))
			r.Post("/message", sendMessageHandler(m.searchSystem, m.storage, m.logger))
		})
	})

	r.Post("/message", sendMessageHandler(m.searchSystem, m.storage, m.logger))
}

func createInteractionHandler(st storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.createInteractionHandler")
		defer span.End()

		var req MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		interaction, err := st.CreateInteraction(ctx, types.Interaction{
			UUID:  uuid.New(),
			Query: req.Query,
			Conversations: []types.Conversation{
				{Role: types.InteractionRoleUser, Text: req.Query},
			},
		})

		if err != nil {
			logger.WithError(err).Error("Failed to create interaction")
			util.SendErrorResponse(w, http.StatusInternalServerError, "failed to create interaction", logger, nil)
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, interaction, logger, nil)
	}
}

func getInteractionsHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := InteractionsResponse{
			Interactions: []InteractionSummary{
				{UUID: uuid.New().String(), Title: "Sample query 1"},
				{UUID: uuid.New().String(), Title: "Sample query 2"},
			},
			Limit:   1,
			PerPage: 10,
		}

		util.SendSuccessResponse(w, http.StatusOK, response, logger, nil)
	}
}

func getInteractionHandler(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		util.SendSuccessResponse(w, http.StatusOK, interaction, logger, nil)
	}
}

func sendMessageHandler(searchSystem search.System, st storage.Storage, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.sendMessageHandler")
		defer span.End()

		var req MessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.SendErrorResponse(w, http.StatusBadRequest, "invalid request", logger, nil)
			return
		}

		llmProvider, err := llm.InitializeLLMProvider(ctx, st, logger)
		if err != nil {
			logger.WithError(err).Error("Failed to initialize LLM provider")
			util.SendErrorResponse(w, http.StatusInternalServerError, "failed to initialize LLM provider", logger, nil)
			return
		}

		llmContexts, err := searchSystem.GenerateLLMContext(ctx, search.Request{Query: req.Query})
		if err != nil {
			logger.WithError(err).Error("Failed to generate LLM contexts")
			util.SendErrorResponse(w, http.StatusInternalServerError, "failed to generate LLM contexts", logger, nil)
			return
		}
		span.SetAttributes(
			attribute.Int("no_of_search_results_for_context", len(llmContexts)),
		)

		promptGenerator := llm.NewPromptGenerator(logger, llm.PromptTemplate{Template: llm.DefaultPromptTemplate})
		prompt, err := promptGenerator.GeneratePrompt(llm.PromptTemplateData{
			Query:                 req.Query,
			Documents:             llmContexts,
			ConversationHistories: []types.Conversation{},
		})

		if err != nil {
			logger.WithError(err).Error("Failed to generate prompt")
			util.SendErrorResponse(w, http.StatusInternalServerError, "failed to generate prompt", logger, nil)
			return
		}

		llmResponse, err := llmProvider.GetResponse(prompt)
		if err != nil {
			logger.WithError(err).Error("Failed to get response from LLM provider")
			util.SendErrorResponse(w, http.StatusInternalServerError, "failed to get response from LLM provider", logger, nil)
			return
		}

		response := MessageResponse{
			Response: llmResponse,
		}

		logger.Infof("LLM response: %s", llmResponse)
		util.SendSuccessResponse(w, http.StatusOK, response, logger, nil)
	}
}
