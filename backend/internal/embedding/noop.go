// Package embedding provides the embedding provider interface and its implementations.
package embedding

import (
	"context"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
)

// Provider is an interface for an embedding provider
type NoOpEmbeddingProvider struct {
	settings *types.NoOpSettings
}

// NewNoOpEmbeddingProvider creates a new NoOpEmbeddingProvider with the given settings
func NewNoOpEmbeddingProvider(settings *types.NoOpSettings) Provider {
	return &NoOpEmbeddingProvider{
		settings: settings,
	}
}

// GetEmbedding returns the embedding for the given text
func (n *NoOpEmbeddingProvider) GetEmbedding(_ context.Context, _ string) ([]types.ContentPart, error) {
	return n.settings.ContentParts, nil
}

// Process processes the given document
func (n *NoOpEmbeddingProvider) Process(_ context.Context, d *types.Document) error {
	d.Embedding.Embedding = n.settings.ContentParts
	return nil
}

// GetID returns the ID of the embedding provider
func (n *NoOpEmbeddingProvider) GetID() uuid.UUID {
	return uuid.UUID{}
}

// HealthCheck checks the health of the embedding provider
func (n *NoOpEmbeddingProvider) HealthCheck() error {
	return nil
}
