package embedding

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

// Provider is an interface for an embedding provider
type Provider interface {
	GetEmbedding(ctx context.Context, text string) ([]types.ContentPart, error)
	Process(ctx context.Context, d *types.Document) error
	GetID() uuid.UUID
	HealthCheck() error
}

// InitializeEmbeddingProvider initializes the embedding provider with the given storage and logger
func InitializeEmbeddingProvider(ctx context.Context, st storage.Storage, logging *logrus.Logger) (Provider, error) {
	activeProvider, err := st.GetAllEmbeddingProviders(ctx, types.EmbeddingProviderFilter{Status: "active"}, types.EmbeddingProviderFilterOption{Limit: 1, Page: 1})
	if err != nil {
		logging.WithError(err).Error("Failed to get active embedding provider")
		return nil, fmt.Errorf("failed to get active embedding provider: %w", err)
	}

	if activeProvider.Total == 0 {
		logging.Warning("No active embedding provider found")
		return nil, nil
	}

	var embeddingProvider Provider

	ep := activeProvider.EmbeddingProviders[0]

	switch ep.Provider {
	case types.EmbeddingProviderTypeOpenAI:
		if openAPIEP, ok := ep.Configuration.(*types.OpenAISettings); !ok {
			return nil, fmt.Errorf("invalid settings type for OpenAI embedding provider: %v", "OpenAI")
		} else {
			embeddingProvider = NewOpenAIProvider(ep.UUID, openAPIEP, logging)
		}
	case types.EmbeddingProviderTypeNoOp:
		if noOpSettings, ok := ep.Configuration.(*types.NoOpSettings); !ok {
			return nil, fmt.Errorf("invalid settings type for Noop embedding provider: %v", "Noop")
		} else {
			embeddingProvider = NewNoOpEmbeddingProvider(noOpSettings)
		}
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", ep.Provider)
	}

	logging.Info("Embedding provider initialized successfully")
	return embeddingProvider, nil
}
