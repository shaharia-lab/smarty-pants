package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants-ai/internal/config"
	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIProvider_GetEmbedding(t *testing.T) {
	tests := []struct {
		name           string
		inputText      string
		mockResponse   OpenAIResponse
		mockStatusCode int
		expectError    bool
	}{
		{
			name:      "Successful embedding",
			inputText: "Test input",
			mockResponse: OpenAIResponse{
				Object: "list",
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{
					{
						Object:    "embedding",
						Index:     0,
						Embedding: []float32{0.1, 0.2, 0.3},
					},
				},
				Model: "text-embedding-3-small",
				Usage: struct {
					PromptTokens int32 `json:"prompt_tokens"`
					TotalTokens  int32 `json:"total_tokens"`
				}{
					PromptTokens: 2,
					TotalTokens:  2,
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error",
			inputText:      "Test input",
			mockResponse:   OpenAIResponse{},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:      "Empty embedding data",
			inputText: "Test input",
			mockResponse: OpenAIResponse{
				Object: "list",
				Data: []struct {
					Object    string    `json:"object"`
					Index     int       `json:"index"`
					Embedding []float32 `json:"embedding"`
				}{},
				Model: "text-embedding-3-small",
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				assert.Equal(t, "POST", r.Method)

				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

				var req OpenAIRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, tt.inputText, req.Input)
				assert.Equal(t, "test-model", req.Model)

				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			provider := NewOpenAIProvider(uuid.New(), &types.OpenAISettings{
				APIKey:      "test-api-key",
				ModelID:     "test-model",
				APIEndpoint: server.URL,
			}, logrus.New())

			ctx := context.Background()
			cleanup, err := observability.InitTracer(ctx, "smarty-pants-ai", logrus.New(), &config.Config{})
			defer cleanup()

			result, err := provider.GetEmbedding(ctx, tt.inputText)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, tt.inputText, result[0].Content)
				assert.Equal(t, tt.mockResponse.Data[0].Embedding, result[0].Embedding)
			}
		})
	}
}
