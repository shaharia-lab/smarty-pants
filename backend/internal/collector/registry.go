package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

// Registry Datasource is an interface for a data source that can be used by the collector
type Registry struct {
	datasources map[uuid.UUID]Datasource
	storage     storage.Storage
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// NewRegistry creates a new registry with the given storage and logger
func NewRegistry(storage storage.Storage, logger *logrus.Logger) *Registry {
	return &Registry{
		datasources: make(map[uuid.UUID]Datasource),
		storage:     storage,
		logger:      logger,
	}
}

// Register adds the given datasource to the registry
func (r *Registry) Register(ds Datasource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.datasources[ds.GetID()]; exists {
		return fmt.Errorf("datasource with ID %s already exists", ds.GetID())
	}

	r.datasources[ds.GetID()] = ds
	return nil
}

// Unregister removes the datasource with the given ID from the registry
func (r *Registry) Unregister(id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.datasources, id)
}

// Get returns the datasource with the given ID, if it exists
func (r *Registry) Get(id uuid.UUID) (Datasource, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, exists := r.datasources[id]
	return ds, exists
}

// GetAll returns all datasources in the registry
func (r *Registry) GetAll() []Datasource {
	r.mu.RLock()
	defer r.mu.RUnlock()

	datasources := make([]Datasource, 0, len(r.datasources))
	for _, ds := range r.datasources {
		datasources = append(datasources, ds)
	}
	return datasources
}

// RefreshDatasources fetches all active datasource configs from the storage and creates new datasource instances
func (r *Registry) RefreshDatasources(ctx context.Context) error {
	r.logger.Info("Refreshing datasources")

	configs, err := r.storage.GetAllDatasources(ctx, 1, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch datasource configs: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.datasources = make(map[uuid.UUID]Datasource)

	for _, config := range configs.Datasources {
		if config.Status != types.DatasourceStatusActive {
			continue
		}

		datasource, err := CreateDatasourceFromConfig(config, r.logger)
		if err != nil {
			r.logger.WithError(err).Errorf("Failed to create datasource from config: %s", config.Name)
			continue
		}
		r.datasources[config.UUID] = datasource
	}

	return nil
}
