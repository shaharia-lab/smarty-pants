// Package search provides a search system for the application.
package search

import (
	"context"
	"fmt"

	"github.com/shaharia-lab/smarty-pants/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/internal/llm"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

// Request is a search request
type Request struct {
	Query string `json:"query"`
}

// System is a search system
type System struct {
	logger    *logrus.Logger
	dbStorage storage.Storage
}

// NewSearchSystem creates a new search system
func NewSearchSystem(logger *logrus.Logger, dbStorage storage.Storage) System {
	return System{
		logger:    logger,
		dbStorage: dbStorage,
	}
}

// HealthCheck checks the health of the system
func (s *System) HealthCheck() error {
	return nil
}

// SearchDocument searches for a document
func (s *System) SearchDocument(ctx context.Context, request Request) (*types.SearchResults, error) {
	embeddingProvider, err := embedding.InitializeEmbeddingProvider(ctx, s.dbStorage, s.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize embedding provider: %w", err)
	}

	queryEmbedding, err := embeddingProvider.GetEmbedding(ctx, request.Query)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get embedding for search query")
		return nil, fmt.Errorf("failed to get embedding for search query: %w", err)
	}

	searchConfig := types.SearchConfig{
		QueryText: request.Query,
		Embedding: queryEmbedding[0].Embedding,
		Limit:     10,
		Page:      1,
	}

	return s.dbStorage.Search(ctx, searchConfig)
}

// GenerateLLMContext generates LLM context
func (s *System) GenerateLLMContext(ctx context.Context, request Request) ([]llm.Documents, error) {
	var documents []llm.Documents

	searchResults, err := s.SearchDocument(ctx, request)
	if err != nil {
		return documents, err
	}

	for _, result := range searchResults.Documents {
		documents = append(documents, llm.Documents{
			Title:    "",
			Content:  result.ContentPart,
			Metadata: nil,
			URL:      "",
		})
	}

	return documents, nil
}
