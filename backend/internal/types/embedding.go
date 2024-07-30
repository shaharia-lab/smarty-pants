package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrEmbeddingProviderNotFound = errors.New("embedding provider not found")

type EmbeddingProviderType string

const (
	EmbeddingProviderTypeOpenAI EmbeddingProviderType = "openai"
	EmbeddingProviderTypeNoOp   EmbeddingProviderType = "noop"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type EmbeddingProviderConfig struct {
	UUID          uuid.UUID                 `json:"uuid"`
	Name          string                    `json:"name"`
	Provider      EmbeddingProviderType     `json:"provider"`
	Configuration EmbeddingProviderSettings `json:"configuration"`
	Status        string                    `json:"status"`
}

type EmbeddingProviderSettings interface {
	Validate() error
	RawJSON() json.RawMessage
	ToMap() map[string]interface{}
}

type OpenAISettings struct {
	APIKey      string `json:"api_key"`
	ModelID     string `json:"model_id"`
	APIEndpoint string `json:"api_endpoint,omitempty"`
}

func (s *OpenAISettings) RawJSON() json.RawMessage {
	data, _ := json.Marshal(s)
	return data
}

func (s *OpenAISettings) ToMap() map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"api_key":  s.APIKey,
		"model_id": s.ModelID,
	}
}

func (s *OpenAISettings) Validate() error {
	if s.APIKey == "" {
		return &ValidationError{Message: "API key is required for OpenAI settings"}
	}
	if s.ModelID == "" {
		return &ValidationError{Message: "Model ID is required for OpenAI settings"}
	}
	return nil
}

type NoOpSettings struct {
	ContentParts []ContentPart
}

func (s *NoOpSettings) RawJSON() json.RawMessage {
	data, _ := json.Marshal(s)
	return data
}

func (s *NoOpSettings) ToMap() map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"content_parts": s.ContentParts,
	}
}

func (s *NoOpSettings) Validate() error {
	return nil
}

func (c *EmbeddingProviderConfig) MarshalJSON() ([]byte, error) {
	type Alias EmbeddingProviderConfig
	return json.Marshal(&struct {
		UUID          string                 `json:"uuid"`
		Configuration map[string]interface{} `json:"configuration"`
		*Alias
	}{
		UUID:          c.UUID.String(),
		Configuration: c.Configuration.ToMap(),
		Alias:         (*Alias)(c),
	})
}

func (c *EmbeddingProviderConfig) UnmarshalJSON(data []byte) error {
	type Alias EmbeddingProviderConfig
	aux := &struct {
		*Alias
		Configuration json.RawMessage `json:"configuration"`
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	c.Configuration, err = ParseEmbeddingProviderSettings(c.Provider, aux.Configuration)
	return err
}

type PaginatedEmbeddingProviders struct {
	EmbeddingProviders []EmbeddingProviderConfig `json:"embedding_providers"`
	Total              int                       `json:"total"`
	Page               int                       `json:"page"`
	PerPage            int                       `json:"per_page"`
	TotalPages         int                       `json:"total_pages"`
}

func ParseEmbeddingProviderSettings(providerType EmbeddingProviderType, settingsJSON json.RawMessage) (EmbeddingProviderSettings, error) {
	var settings EmbeddingProviderSettings

	switch providerType {
	case EmbeddingProviderTypeOpenAI:
		settings = &OpenAISettings{}

	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}

	if err := json.Unmarshal(settingsJSON, settings); err != nil {
		return nil, fmt.Errorf("invalid settings for %s: %v", providerType, err)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

type EmbeddingProviderFilter struct {
	Status string `json:"status"`
}

type EmbeddingProviderFilterOption struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}
