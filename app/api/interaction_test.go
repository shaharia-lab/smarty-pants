package api

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
	"github.com/shaharia-lab/smarty-pants/internal/config"
	"github.com/shaharia-lab/smarty-pants/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/internal/logger"
	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/search"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
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
			inputBody:      `{"query": "Test query"}`,
			expectedStatus: http.StatusOK,
			expectedQuery:  "Test query",
			setupMock: func(st *storage.StorageMock) {
				st.On("CreateInteraction", mock.Anything, mock.MatchedBy(func(i types.Interaction) bool {
					return i.Query == "Test query" &&
						len(i.Conversations) == 1 &&
						i.Conversations[0].Role == types.InteractionRoleUser &&
						i.Conversations[0].Text == "Test query" &&
						i.UUID != uuid.Nil
				})).Return(types.Interaction{
					UUID:  uuid.New(),
					Query: "Test query",
					Conversations: []types.Conversation{
						{Role: types.InteractionRoleUser, Text: "Test query"},
					},
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

			tt.setupMock(st)

			handler := createInteractionHandler(st, logger.NoOpsLogger())

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
				assert.Len(t, response.Conversations, 1)
				assert.Equal(t, types.InteractionRoleUser, response.Conversations[0].Role)
				assert.Equal(t, tt.expectedQuery, response.Conversations[0].Text)
			}

			st.AssertExpectations(t)
		})
	}
}

func TestGetInteractionsHandler(t *testing.T) {
	handler := getInteractionsHandler(logger.NoOpsLogger())

	req, err := http.NewRequest("GET", "/interactions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response InteractionsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Interactions, 2)
	assert.Equal(t, 1, response.Limit)
	assert.Equal(t, 10, response.PerPage)
}

func TestGetInteractionHandler(t *testing.T) {
	handler := getInteractionHandler(logger.NoOpsLogger())

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
	assert.Len(t, response.Conversations, 3)
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
			inputBody:        `{"query": "Test message"}`,
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
			handler := sendMessageHandler(searchSystemMock, sm, logger.NoOpsLogger())

			req, err := http.NewRequest("POST", "/interactions/123/message", bytes.NewBufferString(tt.inputBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			done := make(chan bool)
			go func() {
				handler.ServeHTTP(rr, req)
				done <- true
			}()

			select {
			case <-done:

			case <-time.After(15 * time.Second):
				t.Fatal("Handler took too long to respond")
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response MessageResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, response.Response)
			}
		})
	}
}
