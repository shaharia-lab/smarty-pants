package util

import (
	"encoding/json"
	"testing"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestParseSettings(t *testing.T) {
	tests := []struct {
		name           string
		sourceType     types.DatasourceType
		settingsJSON   string
		expectedResult types.DatasourceSettings
		expectedError  string
	}{
		{
			name:         "Valid Slack Settings",
			sourceType:   types.DatasourceTypeSlack,
			settingsJSON: `{"token": "test-token", "channel_id": "test-channel", "workspace": "test-workspace"}`,
			expectedResult: &types.SlackSettings{
				Token:     "test-token",
				ChannelID: "test-channel",
				Workspace: "test-workspace",
			},
			expectedError: "",
		},
		{
			name:           "Invalid Slack Settings - Missing Fields",
			sourceType:     types.DatasourceTypeSlack,
			settingsJSON:   `{"token": "test-token"}`,
			expectedResult: nil,
			expectedError:  "slack channel_id is required",
		},
		{
			name:           "Unsupported Source Type",
			sourceType:     "unsupported",
			settingsJSON:   `{}`,
			expectedResult: nil,
			expectedError:  "unsupported source type: unsupported",
		},
		{
			name:           "Invalid JSON",
			sourceType:     types.DatasourceTypeSlack,
			settingsJSON:   `{invalid-json}`,
			expectedResult: nil,
			expectedError:  "invalid settings for slack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSettings(tt.sourceType, json.RawMessage(tt.settingsJSON))

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestParseEmbeddingProviderSettings(t *testing.T) {
	tests := []struct {
		name           string
		providerType   types.EmbeddingProviderType
		settingsJSON   string
		expectedResult types.EmbeddingProviderSettings
		expectedError  string
	}{
		{
			name:         "Valid OpenAI Settings",
			providerType: types.EmbeddingProviderTypeOpenAI,
			settingsJSON: `{"api_key": "test-key", "model_id": "test-model"}`,
			expectedResult: &types.OpenAISettings{
				APIKey:  "test-key",
				ModelID: "test-model",
			},
			expectedError: "",
		},
		{
			name:           "Invalid OpenAI Settings - Missing Fields",
			providerType:   types.EmbeddingProviderTypeOpenAI,
			settingsJSON:   `{"api_key": "test-key"}`,
			expectedResult: nil,
			expectedError:  "Model ID is required for OpenAI settings",
		},
		{
			name:           "Unsupported Provider Type",
			providerType:   "unsupported",
			settingsJSON:   `{}`,
			expectedResult: nil,
			expectedError:  "unsupported embedding provider type: unsupported",
		},
		{
			name:           "Invalid JSON",
			providerType:   types.EmbeddingProviderTypeOpenAI,
			settingsJSON:   `{invalid-json}`,
			expectedResult: nil,
			expectedError:  "invalid settings for openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEmbeddingProviderSettings(tt.providerType, json.RawMessage(tt.settingsJSON))

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestParseLLMProviderSettings(t *testing.T) {
	tests := []struct {
		name           string
		providerType   types.LLMProviderType
		settingsJSON   string
		expectedResult types.LLMProviderSettings
		expectedError  string
	}{
		{
			name:         "Valid OpenAI LLM Settings",
			providerType: types.LLMProviderTypeOpenAI,
			settingsJSON: `{"api_key": "test-key", "model_id": "test-model"}`,
			expectedResult: &types.OpenAILLMSettings{
				APIKey:  "test-key",
				ModelID: "test-model",
			},
			expectedError: "",
		},
		{
			name:         "Valid NoOp LLM Settings",
			providerType: types.LLMProviderTypeNoOps,
			settingsJSON: `{"response_to_return": "test-response"}`,
			expectedResult: &types.NoOpLLMProviderSettings{
				ResponseToReturn: "test-response",
			},
			expectedError: "",
		},
		{
			name:           "Invalid OpenAI LLM Settings - Missing Fields",
			providerType:   types.LLMProviderTypeOpenAI,
			settingsJSON:   `{"api_key": "test-key"}`,
			expectedResult: nil,
			expectedError:  "Model ID is required for OpenAI settings",
		},
		{
			name:           "Unsupported Provider Type",
			providerType:   "unsupported",
			settingsJSON:   `{}`,
			expectedResult: nil,
			expectedError:  "unsupported LLM provider type: unsupported",
		},
		{
			name:           "Invalid JSON",
			providerType:   types.LLMProviderTypeOpenAI,
			settingsJSON:   `{invalid-json}`,
			expectedResult: nil,
			expectedError:  "invalid settings for openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLLMProviderSettings(tt.providerType, json.RawMessage(tt.settingsJSON))

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestInvalidSettingsErr(t *testing.T) {
	tests := []struct {
		name          string
		sType         interface{}
		err           error
		expectedError string
	}{
		{
			name:          "String sType",
			sType:         "test-type",
			err:           assert.AnError,
			expectedError: "invalid settings for test-type: assert.AnError general error for testing",
		},
		{
			name:          "Integer sType",
			sType:         123,
			expectedError: "invalid settings for 123: assert.AnError general error for testing",
		},
		{
			name:          "Struct sType",
			sType:         struct{ Name string }{"test"},
			err:           assert.AnError,
			expectedError: "invalid settings for {test}: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := invalidSettingsErr(tt.sType, assert.AnError)
			assert.EqualError(t, result, tt.expectedError)
		})
	}
}
