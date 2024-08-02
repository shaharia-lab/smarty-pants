package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

func invalidSettingsErr(sType interface{}, err error) error {
	return fmt.Errorf("invalid settings for %s: %v", sType, err)
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
	var settings types.EmbeddingProviderSettings

	switch providerType {
	case types.LLMProviderTypeOpenAI:
		settings = &types.OpenAILLMSettings{}

	case types.LLMProviderTypeNoOps:
		settings = &types.NoOpLLMProviderSettings{}

	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}

	if err := json.Unmarshal(settingsJSON, settings); err != nil {
		return nil, invalidSettingsErr(providerType, err)
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

// CompareVersions compares two version strings. It returns -1 if v1 < v2, 1 if v1 > v2, and 0 if v1 == v2
func CompareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	if len(parts1) < len(parts2) {
		return -1
	}
	if len(parts1) > len(parts2) {
		return 1
	}
	return 0
}
