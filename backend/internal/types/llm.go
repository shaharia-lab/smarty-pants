package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrLLMProviderNotFound = errors.New("LLM provider not found")

type LLMProviderType string

type InteractionRole string

type LLMProviderStatus string

const (
	LLMProviderTypeOpenAI LLMProviderType = "openai"
	LLMProviderTypeNoOps  LLMProviderType = "noop"

	InteractionRoleUser InteractionRole = "user"

	LLMProviderStatusActive   LLMProviderStatus = "active"
	LLMProviderStatusInactive LLMProviderStatus = "inactive"
)

type LLMProviderConfig struct {
	UUID          uuid.UUID           `json:"uuid"`
	Name          string              `json:"name"`
	Provider      LLMProviderType     `json:"provider"`
	Configuration LLMProviderSettings `json:"configuration"`
	Status        string              `json:"status"`
}

type NoOpLLMProviderSettings struct {
	ResponseToReturn string `json:"response_to_return"`
	HealthCheckError *error `json:"health_check_error,omitempty"`
}

func (s *NoOpLLMProviderSettings) Validate() error {
	return nil
}

func (s *NoOpLLMProviderSettings) RawJSON() json.RawMessage {
	data, _ := json.Marshal(s)
	return data
}

func (s *NoOpLLMProviderSettings) ToMap() map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"response_to_return": s.ResponseToReturn,
		"health_check_error": s.HealthCheckError,
	}
}

type LLMProviderSettings interface {
	Validate() error
	RawJSON() json.RawMessage
	ToMap() map[string]interface{}
}

type OpenAILLMSettings struct {
	APIKey  string `json:"api_key"`
	ModelID string `json:"model_id"`
}

func (s *OpenAILLMSettings) RawJSON() json.RawMessage {
	data, _ := json.Marshal(s)
	return data
}

func (s *OpenAILLMSettings) ToMap() map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"api_key":  s.APIKey,
		"model_id": s.ModelID,
	}
}

func (s *OpenAILLMSettings) Validate() error {
	if s.APIKey == "" {
		return &ValidationError{Message: "API key is required for OpenAI settings"}
	}
	if s.ModelID == "" {
		return &ValidationError{Message: "Model ID is required for OpenAI settings"}
	}
	return nil
}

func (c *LLMProviderConfig) MarshalJSON() ([]byte, error) {
	type Alias LLMProviderConfig
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

func (c *LLMProviderConfig) UnmarshalJSON(data []byte) error {
	type Alias LLMProviderConfig
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
	c.Configuration, err = ParseLLMProviderSettings(c.Provider, aux.Configuration)
	return err
}

type PaginatedLLMProviders struct {
	LLMProviders []LLMProviderConfig `json:"llm_providers"`
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	PerPage      int                 `json:"per_page"`
	TotalPages   int                 `json:"total_pages"`
}

func ParseLLMProviderSettings(providerType LLMProviderType, settingsJSON json.RawMessage) (LLMProviderSettings, error) {
	var settings LLMProviderSettings

	switch providerType {
	case LLMProviderTypeOpenAI:
		settings = &OpenAILLMSettings{}
	case LLMProviderTypeNoOps:
		settings = &NoOpLLMProviderSettings{}
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}

	if err := json.Unmarshal(settingsJSON, settings); err != nil {
		return nil, fmt.Errorf("invalid settings for %s: %v", providerType, err)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

type LLMProviderFilter struct {
	Status string `json:"status"`
}

type LLMProviderFilterOption struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

type Conversation struct {
	UUID      uuid.UUID       `json:"uuid"`
	Role      InteractionRole `json:"role"`
	Text      string          `json:"text"`
	CreatedAt time.Time       `json:"created_at"`
}
