package processor

import (
	"context"

	"github.com/shaharia-lab/smarty-pants/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

// Embedding is a processor that adds embeddings to a document
type Embedding struct {
	embeddingProvider embedding.Provider
	logging           *logrus.Logger
}

// NewEmbedding creates a new embedding processor
func NewEmbedding(embeddingProvider embedding.Provider, logger *logrus.Logger) *Embedding {
	return &Embedding{embeddingProvider: embeddingProvider, logging: logger}
}

// Process adds embeddings to a document
func (e *Embedding) Process(ctx context.Context, d *types.Document) (*types.Document, error) {

	e.logging.WithField("document_uuid", d.UUID).Info("Starting embedding processor for document")

	em, err := e.embeddingProvider.GetEmbedding(ctx, d.Body)
	if err != nil {
		return &types.Document{}, err
	}

	d.Embedding.Embedding = em
	return d, nil
}
