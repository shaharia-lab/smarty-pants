package processor

import (
	"context"
	"fmt"
	"sync"

	"github.com/shaharia-lab/smarty-pants-ai/internal/embedding"
	"github.com/shaharia-lab/smarty-pants-ai/internal/storage"
	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
	"github.com/sirupsen/logrus"
)

// ProcessorUnit is a processor unit that processes documents
type ProcessorUnit interface {
	Process(ctx context.Context, d *types.Document) error
}

// Registry is a registry of processors
type Registry struct {
	processors map[string]ProcessorUnit
	storage    storage.Storage
	logger     *logrus.Logger
	mu         sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry(storage storage.Storage, logger *logrus.Logger) *Registry {
	return &Registry{
		processors: make(map[string]ProcessorUnit),
		storage:    storage,
		logger:     logger,
	}
}

// Register registers a processor
func (r *Registry) Register(name string, processor ProcessorUnit) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.processors[name] = processor
	return nil
}

// Get returns a processor
func (r *Registry) Get(name string) (ProcessorUnit, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.processors[name]
	return p, exists
}

// GetAll returns all processors
func (r *Registry) GetAll() []ProcessorUnit {
	r.mu.RLock()
	defer r.mu.RUnlock()

	processors := make([]ProcessorUnit, 0, len(r.processors))
	for _, p := range r.processors {
		processors = append(processors, p)
	}
	return processors
}

// RefreshProcessors refreshes the processors in the registry
func (r *Registry) RefreshProcessors(ctx context.Context) error {
	emProvider, err := embedding.InitializeEmbeddingProvider(ctx, r.storage, r.logger)
	if err != nil {
		r.logger.WithError(err).Error("Failed to initialize embedding provider")
		return fmt.Errorf("failed to initialize embedding provider: %w", err)
	}

	if emProvider == nil {
		r.logger.Warn("No active embedding provider found")
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.processors = make(map[string]ProcessorUnit)

	r.logger.WithField("provider", emProvider.GetID()).Info("Active embedding provider found. Updating processor registry")
	r.processors["embedding_generation"] = emProvider

	return nil
}
