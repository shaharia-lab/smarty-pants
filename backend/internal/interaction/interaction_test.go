package interaction

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/shaharia-lab/smarty-pants/backend/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateInteractionHandler(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      string
		expectedStatus int
		expectedQuery  string
		setupMock      func(*storage.StorageMock)
	}{
		{
			name:           "Success",
			inputBody:      `{"role": "user", "text": "Test query"}`,
			expectedStatus: http.StatusOK,
			expectedQuery:  "Test query",
			setupMock: func(st *storage.StorageMock) {
				st.On("CreateInteraction", mock.Anything, mock.Anything).Return(types.Interaction{
					UUID:          uuid.New(),
					Query:         "Test query",
					Conversations: nil,
				}, nil)
			},
		},
		{
			name:           "Bad Request - Invalid JSON",
			inputBody:      `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			expectedQuery:  "",
			setupMock:      func(st *storage.StorageMock) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observability.InitTracer(context.Background(), "smarty-pants-ai", logger.NoOpsLogger(), &config.Config{TracingEnabled: false})

			st := new(storage.StorageMock)
			l := logger.NoOpsLogger()
			ss := search.NewSearchSystem(l, st)
			aclManager := auth.NewACLManager(l, false)

			tt.setupMock(st)

			m := NewManager(st, l, ss, aclManager)
			handler := http.HandlerFunc(m.createInteractionHandler)

			req, err := http.NewRequest("POST", "/interactions", bytes.NewBufferString(tt.inputBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response types.Interaction
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.UUID)
				assert.Equal(t, tt.expectedQuery, response.Query)
			}

			st.AssertExpectations(t)
		})
	}
}

func TestGetInteractionsHandler(t *testing.T) {
	st := new(storage.StorageMock)
	st.On("GetAllInteractions", mock.Anything, 1, 10).Return(&types.PaginatedInteractions{
		Interactions: []types.Interaction{
			{
				UUID:          uuid.UUID{},
				Query:         "",
				Conversations: nil,
				CreatedAt:     time.Time{},
			},
			{
				UUID:          uuid.UUID{},
				Query:         "",
				Conversations: nil,
				CreatedAt:     time.Time{},
			},
		},
		Total:      2,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}, nil)

	l := logger.NoOpsLogger()
	ss := search.NewSearchSystem(l, st)
	aclManager := auth.NewACLManager(l, false)

	m := NewManager(st, l, ss, aclManager)
	handler := http.HandlerFunc(m.getInteractionsHandler)

	req, err := http.NewRequest("GET", "/interactions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response types.PaginatedInteractions
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Interactions, 2)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 10, response.PerPage)
}

func TestGetInteractionHandler(t *testing.T) {
	st := new(storage.StorageMock)
	st.On("GetInteraction", mock.Anything, mock.Anything).Return(types.Interaction{
		UUID:  uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
		Query: "Sample query",
		Conversations: []types.Conversation{
			{
				UUID:      uuid.UUID{},
				Role:      "",
				Text:      "",
				CreatedAt: time.Time{},
			},
			{
				UUID:      uuid.UUID{},
				Role:      "",
				Text:      "",
				CreatedAt: time.Time{},
			},
		},
		CreatedAt: time.Time{},
	}, nil)
	l := logger.NoOpsLogger()
	ss := search.NewSearchSystem(l, st)
	aclManager := auth.NewACLManager(l, false)

	m := NewManager(st, l, ss, aclManager)
	handler := http.HandlerFunc(m.getInteractionHandler)

	router := chi.NewRouter()
	router.Get("/interactions/{uuid}", handler)

	uuidParsed := uuid.MustParse("12345678-1234-1234-1234-1234567890ab")

	req, err := http.NewRequest("GET", "/interactions/"+uuidParsed.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response types.Interaction
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uuidParsed, response.UUID)
	assert.Equal(t, "Sample query", response.Query)
	assert.Len(t, response.Conversations, 2)
}

func TestSendMessageHandler(t *testing.T) {
	tests := []struct {
		name             string
		inputBody        string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "Success",
			inputBody:        `{"text": "Test message", "role": "user"}`,
			expectedStatus:   http.StatusOK,
			expectedResponse: "Thank you for your message",
		},
		{
			name:             "Bad Request - Invalid JSON",
			inputBody:        `{invalid json}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, _ = observability.InitTracer(ctx, "smarty-pants-ai", logger.NoOpsLogger(), &config.Config{})

			sm := new(storage.StorageMock)
			userMessageMatcher := mock.MatchedBy(func(ctx context.Context) bool {
				return true
			})

			// Set up expectations with the custom matchers
			sm.On("AddConversation", userMessageMatcher, mock.Anything, "user", "Test message").Return(types.Conversation{}, nil).Once()
			sm.On("AddConversation", userMessageMatcher, mock.Anything, "system", "Thank you for your message").Return(types.Conversation{
				UUID:      uuid.UUID{},
				Role:      "system",
				Text:      "Thank you for your message",
				CreatedAt: time.Time{},
			}, nil).Once()

			sm.On("GetAllLLMProviders",
				mock.Anything,
				types.LLMProviderFilter{Status: "active"},
				types.LLMProviderFilterOption{Limit: 1, Page: 1}).Return(
				&types.PaginatedLLMProviders{
					LLMProviders: []types.LLMProviderConfig{
						{
							UUID:     uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
							Name:     "test_llm",
							Provider: types.LLMProviderTypeNoOps,
							Configuration: &types.NoOpLLMProviderSettings{
								ResponseToReturn: "Thank you for your message",
							},
							Status: "active",
						},
					},
					Total:      1,
					Page:       1,
					PerPage:    1,
					TotalPages: 1,
				}, nil)

			sm.On("GetAllEmbeddingProviders",
				mock.Anything,
				types.EmbeddingProviderFilter{Status: "active"},
				types.EmbeddingProviderFilterOption{Limit: 1, Page: 1}).Return(
				&types.PaginatedEmbeddingProviders{
					EmbeddingProviders: []types.EmbeddingProviderConfig{
						{
							UUID:     uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
							Name:     "test_embedding",
							Provider: types.EmbeddingProviderTypeNoOp,
							Configuration: &types.NoOpSettings{ContentParts: []types.ContentPart{
								{
									Content:                   "Hi",
									Embedding:                 []float32{0.03},
									EmbeddingProviderUUID:     uuid.UUID{},
									EmbeddingPromptTotalToken: 0,
									GeneratedAt:               time.Time{},
								},
							}},
							Status: "active",
						},
					},
					Total:      1,
					Page:       1,
					PerPage:    1,
					TotalPages: 1,
				}, nil)

			sm.On("Search", mock.Anything, types.SearchConfig{
				QueryText:  "Test message",
				Embedding:  []float32{0.03},
				Status:     "",
				SourceType: "",
				Limit:      10,
				Page:       1,
			}).Return(&types.SearchResults{
				Documents: []types.SearchResultsDocument{
					{
						ContentPart:          "Content part 1",
						ContentPartID:        1,
						OriginalDocumentUUID: uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
						RelevantScore:        0.9,
					},
				},
				QueryText:    "Test message",
				Limit:        1,
				Page:         1,
				TotalPages:   1,
				TotalResults: 1,
			}, nil)

			em := new(embedding.EmbeddingProviderMock)
			em.On("GetEmbedding", mock.Anything, "Test message").Return([]types.ContentPart{
				{
					Content:                   "Content part 1",
					Embedding:                 []float32{0.03},
					EmbeddingProviderUUID:     uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
					EmbeddingPromptTotalToken: 120,
					GeneratedAt:               time.Time{},
				},
			}, nil)

			searchSystemMock := search.NewSearchSystem(logger.NoOpsLogger(), sm)

			l := logger.NoOpsLogger()
			aclManager := auth.NewACLManager(l, false)

			m := NewManager(sm, l, searchSystemMock, aclManager)

			r := chi.NewRouter()
			r.Post("/interactions/{uuid}/message", m.sendMessageHandler)

			req, err := http.NewRequest("POST", "/interactions/"+uuid.New().String()+"/message", bytes.NewBufferString(tt.inputBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			done := make(chan bool)
			go func() {
				r.ServeHTTP(rr, req)
				done <- true
			}()

			select {
			case <-done:

			case <-time.After(15 * time.Second):
				t.Fatal("Handler took too long to respond")
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var c types.Conversation
				err = json.Unmarshal(rr.Body.Bytes(), &c)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, c.Text)
			}

			if tt.expectedStatus != http.StatusBadRequest {
				sm.AssertExpectations(t)
			}
		})
	}
}
