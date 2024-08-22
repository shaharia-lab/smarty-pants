package util

import (
	"encoding/json"
	"fmt"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
)

func ParseSettings(sourceType types.DatasourceType, settingsJSON json.RawMessage) (types.DatasourceSettings, error) {
	var settings types.DatasourceSettings

	switch sourceType {
	case types.DatasourceTypeSlack:
		settings = &types.SlackSettings{}
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}

	if err := json.Unmarshal(settingsJSON, settings); err != nil {
		return nil, invalidSettingsErr(sourceType, err)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

func ParseEmbeddingProviderSettings(providerType types.EmbeddingProviderType, settingsJSON json.RawMessage) (types.EmbeddingProviderSettings, error) {
	var settings types.EmbeddingProviderSettings

	switch providerType {
	case types.EmbeddingProviderTypeOpenAI:
		settings = &types.OpenAISettings{}

	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}

	if err := json.Unmarshal(settingsJSON, settings); err != nil {
		return nil, invalidSettingsErr(providerType, err)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

func ParseLLMProviderSettings(providerType types.LLMProviderType, settingsJSON json.RawMessage) (types.LLMProviderSettings, error) {
	switch providerType {
	case types.LLMProviderTypeOpenAI:
		settings := &types.OpenAILLMSettings{}
		if err := json.Unmarshal(settingsJSON, settings); err != nil {
			return nil, invalidSettingsErr(providerType, err)
		}
		if err := settings.Validate(); err != nil {
			return nil, err
		}
		return settings, nil

	case types.LLMProviderTypeNoOps:
		settings := &types.NoOpLLMProviderSettings{}
		if err := json.Unmarshal(settingsJSON, settings); err != nil {
			return nil, invalidSettingsErr(providerType, err)
		}
		// NoOp settings don't need validation
		return settings, nil

	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}
}

func invalidSettingsErr(sType interface{}, err error) error {
	return fmt.Errorf("invalid settings for %v: %v", sType, err)
}
