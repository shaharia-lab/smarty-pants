package llm

import "github.com/shaharia-lab/smarty-pants-ai/internal/types"

// NoOpsLLM is a language model provider that does nothing
type NoOpsLLM struct {
	settings *types.NoOpLLMProviderSettings
}

// NewNoOpsLLM creates a new NoOpsLLM with the given settings
func NewNoOpsLLM(settings *types.NoOpLLMProviderSettings) *NoOpsLLM {
	return &NoOpsLLM{
		settings: settings,
	}
}

// HealthCheck checks the health of the language model provider
func (n *NoOpsLLM) HealthCheck() error {
	return nil
}

// GetResponse returns a response from the language model provider
func (n *NoOpsLLM) GetResponse(_ Prompt) (string, error) {
	return n.settings.ResponseToReturn, nil
}
