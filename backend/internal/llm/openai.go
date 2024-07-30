package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	openAIAPIURL       = "https://api.openai.com/v1/chat/completions"
	errDecodingRespMsg = "Error decoding response"
)

// OpenAILLM is a language model provider that uses OpenAI to generate responses
type OpenAILLM struct {
	apiKey     string
	httpClient HTTPClient
	modelID    string
	logging    *logrus.Logger
}

// HTTPClient is an interface for an HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewOpenAILLM creates a new OpenAILLM with the given settings
func NewOpenAILLM(settings *types.OpenAILLMSettings, client HTTPClient, logging *logrus.Logger) *OpenAILLM {
	if client == nil {
		client = &http.Client{}
	}
	return &OpenAILLM{
		apiKey:     settings.APIKey,
		httpClient: client,
		modelID:    settings.ModelID,
		logging:    logging,
	}
}

// HealthCheck checks the health of the language model provider
func (o *OpenAILLM) HealthCheck() error {
	return nil
}

// GetResponse returns a response from the language model provider
func (o *OpenAILLM) GetResponse(prompt Prompt) (string, error) {
	_, span := observability.StartSpan(context.Background(), "llm.GetResponse")
	defer span.End()

	requestBody, err := o.createRequestBody(prompt)
	if err != nil {
		o.logging.WithError(err).Error("Error creating request body")
		return "", fmt.Errorf("error creating request body: %w", err)
	}

	span.SetAttributes(
		attribute.String("llm_endpoint", openAIAPIURL),
		attribute.String("llm_model_id", o.modelID),
		attribute.String("llm_provider", string(types.LLMProviderTypeOpenAI)),
	)

	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		o.logging.WithError(err).Error("Error creating request")
		return "", fmt.Errorf("error creating request: %w", err)
	}

	o.setHeaders(req)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.logging.WithError(err).Error("Error sending request")
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			o.logging.WithError(err).Error("Error closing response body")
		}
	}(resp.Body)

	span.SetAttributes(attribute.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		o.logging.WithField("status_code", resp.StatusCode).Error("Unexpected status code")
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return o.parseResponse(span, resp.Body)
}

func (o *OpenAILLM) createRequestBody(prompt Prompt) ([]byte, error) {
	requestData := struct {
		Model    string        `json:"model"`
		Messages []interface{} `json:"messages"`
	}{
		Model: o.modelID,
		Messages: []interface{}{
			map[string]string{"role": "system", "content": "You are a helpful assistant."},
			map[string]string{"role": "user", "content": prompt.Text},
		},
	}

	return json.Marshal(requestData)
}

func (o *OpenAILLM) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))
}

func (o *OpenAILLM) parseResponse(span trace.Span, body io.Reader) (string, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}
	}

	if err := json.NewDecoder(body).Decode(&response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		o.logging.WithField("raw_response_body", body).Debug(errDecodingRespMsg)
		o.logging.WithError(err).Error(errDecodingRespMsg)
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Choices) == 0 {
		span.RecordError(errors.New("no choices in response"))
		span.SetStatus(codes.Error, "No choices in response")

		o.logging.WithField("raw_response_body", body).Debug(errDecodingRespMsg)
		o.logging.Error("No choices in response")
		return "", errors.New("no choices in response")
	}

	span.SetAttributes(
		attribute.Int("prompt_tokens", response.Usage.PromptTokens),
		attribute.Int("completion_tokens", response.Usage.CompletionTokens),
		attribute.Int("total_tokens", response.Usage.TotalTokens),
	)

	return response.Choices[0].Message.Content, nil
}
