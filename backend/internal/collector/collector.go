// Package collector provides the main collector struct and methods for managing the collection of data from datasources.
package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
)

// Config contains configuration options for the collector
type Collector struct {
	config   Config
	registry *Registry
	manager  *Manager
	storage  storage.Storage
	logger   *logrus.Logger
	wg       sync.WaitGroup
	stopCh   chan struct{}
	metrics  *CollectorMetrics
}

// Config contains configuration options for the collector
type CollectorMetrics struct {
	collectionCycles   metric.Int64Counter
	collectionDuration metric.Int64Histogram
	datasourceDuration metric.Int64Histogram
	errorsTotal        metric.Int64Counter
}

// NewCollector creates a new collector with the given configuration, storage, logger, and meter
func NewCollector(config Config, storage storage.Storage, logger *logrus.Logger, meter metric.Meter) (*Collector, error) {
	registry := NewRegistry(storage, logger)
	workers := make([]*Worker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		workers[i] = NewWorker(i, storage, logger, config, meter)
	}
	manager := NewManager(registry, workers, config, logger, meter)

	metrics, err := initMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return &Collector{
		config:   config,
		registry: registry,
		manager:  manager,
		storage:  storage,
		logger:   logger,
		stopCh:   make(chan struct{}),
		metrics:  metrics,
	}, nil
}

func initMetrics(meter metric.Meter) (*CollectorMetrics, error) {
	collectionCycles, err := meter.Int64Counter("collector_collection_cycles_total")
	if err != nil {
		return nil, err
	}

	collectionDuration, err := meter.Int64Histogram("collector_collection_duration_seconds")
	if err != nil {
		return nil, err
	}

	datasourceDuration, err := meter.Int64Histogram("collector_datasource_duration_seconds")
	if err != nil {
		return nil, err
	}

	errorsTotal, err := meter.Int64Counter("collector_errors_total")
	if err != nil {
		return nil, err
	}

	return &CollectorMetrics{
		collectionCycles:   collectionCycles,
		collectionDuration: collectionDuration,
		datasourceDuration: datasourceDuration,
		errorsTotal:        errorsTotal,
	}, nil
}

// Start starts the collector and its workers
func (c *Collector) Start(ctx context.Context) error {
	ctx, span := observability.StartSpan(ctx, "Collector.Start")
	defer span.End()

	c.logger.Info("Starting collector")

	if err := c.registry.RefreshDatasources(ctx); err != nil {
		return fmt.Errorf("failed to refresh datasources: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.manager.Start(ctx)
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.config.DatasourceRefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopCh:
				return
			case <-ticker.C:
				if err := c.registry.RefreshDatasources(ctx); err != nil {
					c.logger.WithError(err).Error("Failed to refresh datasources")
				}
			}
		}
	}()

	return nil
}

func (c *Collector) Stop() {
	ctx, span := observability.StartSpan(context.Background(), "Collector.Stop")
	defer span.End()

	c.logger.Info("Stopping collector")
	close(c.stopCh)

	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("Collector stopped gracefully")
	case <-timeoutCtx.Done():
		c.logger.Warn("Collector stop timed out, forcing shutdown")
	}
}

func (c *Collector) RegisterDatasource(ds types.DataSource) error {
	_, span := observability.StartSpan(context.Background(), "Collector.RegisterDatasource")
	defer span.End()

	c.logger.WithField("datasource_id", ds.GetID()).Info("Registering datasource")
	return c.registry.Register(ds)
}

func (c *Collector) UnregisterDatasource(id string) {
	_, span := observability.StartSpan(context.Background(), "Collector.UnregisterDatasource")
	defer span.End()

	c.logger.WithField("datasource_id", id).Info("Unregistering datasource")
	c.registry.Unregister(uuid.MustParse(id))
}
