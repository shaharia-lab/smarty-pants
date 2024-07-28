package llm

import (
	"context"
	"fmt"

	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/storage"
	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
)

// LLM is an interface for a language model provider
type LLM interface {
	HealthCheck() error
	GetResponse(prompt Prompt) (string, error)
}

// InitializeLLMProvider initializes the LLM provider
func InitializeLLMProvider(ctx context.Context, st storage.Storage, logging *logrus.Logger) (LLM, error) {
	filter := types.LLMProviderFilter{
		Status: "active",
	}

	option := types.LLMProviderFilterOption{
		Limit: 1,
		Page:  1,
	}

	_, span := observability.StartSpan(ctx, "llm.InitializeLLMProvider.GetActiveProvider")
	defer span.End()

	providers, err := st.GetAllLLMProviders(ctx, filter, option)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to retrieve LLM providers: %w", err)
	}

	if len(providers.LLMProviders) == 0 {
		return nil, fmt.Errorf("no active LLM provider found")
	}

	activeProvider := providers.LLMProviders[0]

	switch activeProvider.Provider {
	case types.LLMProviderTypeOpenAI:
		openAISettings, ok := activeProvider.Configuration.(*types.OpenAILLMSettings)
		if !ok {
			return nil, fmt.Errorf("invalid configuration for OpenAI provider")
		}

		logging.Infof("Initializing OpenAI LLM provider with model: %s", openAISettings.ModelID)
		return NewOpenAILLM(openAISettings, nil, logging), nil

	case types.LLMProviderTypeNoOps:
		NoOpsLLMSettings, ok := activeProvider.Configuration.(*types.NoOpLLMProviderSettings)
		if !ok {
			return nil, fmt.Errorf("invalid configuration for NoOps provider")
		}

		return NewNoOpsLLM(NoOpsLLMSettings), nil

	default:
		logging.WithField("provider", activeProvider.Provider).Error("Unsupported LLM provider type")
		return &Fake{}, nil
	}
}
