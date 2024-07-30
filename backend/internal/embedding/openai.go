package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	defaultOpenAIEmbeddingURL = "https://api.openai.com/v1/embeddings"
)

// OpenAIProvider is an embedding provider that uses OpenAI to generate embeddings
type OpenAIProvider struct {
	apiKey                string
	model                 string
	client                *http.Client
	url                   string
	logging               *logrus.Logger
	embeddingProviderUUID uuid.UUID
}

// OpenAIRequest is a request to the OpenAI API
type OpenAIRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// OpenAIResponse is a response from the OpenAI API
type OpenAIResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int32 `json:"prompt_tokens"`
		TotalTokens  int32 `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIProvider creates a new OpenAIProvider with the given settings
func NewOpenAIProvider(providerUUID uuid.UUID, settings *types.OpenAISettings, logging *logrus.Logger) Provider {
	apiEndpoint := settings.APIEndpoint
	if apiEndpoint == "" {
		apiEndpoint = defaultOpenAIEmbeddingURL
	}

	return &OpenAIProvider{
		apiKey:                settings.APIKey,
		model:                 settings.ModelID,
		client:                &http.Client{},
		url:                   apiEndpoint,
		logging:               logging,
		embeddingProviderUUID: providerUUID,
	}
}

// HealthCheck checks the health of the embedding provider
func (p *OpenAIProvider) HealthCheck() error {
	return nil
}

// GetID returns the ID of the embedding provider
func (p *OpenAIProvider) GetID() uuid.UUID {
	return p.embeddingProviderUUID
}

// GetEmbedding returns the embedding for the given text
func (p *OpenAIProvider) GetEmbedding(ctx context.Context, text string) ([]types.ContentPart, error) {
	p.logging.WithField("embedding_provider", "openai").Info("Generating embedding")

	ctx, span := observability.StartSpan(ctx, "llm.GetEmbedding")
	span.SetAttributes(
		attribute.String("embedding_provider", string(types.EmbeddingProviderTypeOpenAI)),
		attribute.String("model", p.model),
	)

	defer span.End()

	req := OpenAIRequest{
		Input: text,
		Model: p.model,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.url, bytes.NewBuffer(reqBody))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			p.logging.WithError(closeErr).Error("Failed to close response body")
			span.RecordError(closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.SetAttributes(attribute.Int("status_code", resp.StatusCode))
		span.RecordError(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		span.SetStatus(codes.Error, "unexpected status code")

		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Data) == 0 {
		errMsg := "no embedding data in response"
		span.RecordError(fmt.Errorf(errMsg))
		span.SetStatus(codes.Error, errMsg)

		return nil, fmt.Errorf(errMsg)
	}

	span.SetAttributes(
		attribute.Int("prompt_token", int(openAIResp.Usage.PromptTokens)),
		attribute.Int("total_tokens", int(openAIResp.Usage.TotalTokens)),
	)

	p.logging.WithField("total_tokens", openAIResp.Usage.TotalTokens).Info("Embedding generated successfully")

	return []types.ContentPart{
		types.NewContentPart(text, openAIResp.Data[0].Embedding, p.embeddingProviderUUID, openAIResp.Usage.PromptTokens),
	}, nil
}

// Process processes the document by generating an embedding for the document body
func (p *OpenAIProvider) Process(ctx context.Context, d *types.Document) error {
	p.logging.WithField("document_uuid", d.UUID).Info("Starting embedding processor for document")

	em, err := p.GetEmbedding(ctx, d.Body)
	if err != nil {
		return err
	}

	d.Embedding.Embedding = em
	return nil
}
