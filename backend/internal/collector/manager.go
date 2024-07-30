package collector

import (
	"context"
	"sync"
	"time"

	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Manager Config contains configuration options for the manager
type Manager struct {
	registry *Registry
	workers  []*Worker
	config   Config
	logger   *logrus.Logger
	metrics  *ManagerMetrics
}

// ManagerMetrics contains metrics for the manager
type ManagerMetrics struct {
	activeWorkers metric.Int64UpDownCounter
}

// NewManager creates a new manager with the given registry, workers, configuration, logger, and meter
func NewManager(registry *Registry, workers []*Worker, config Config, logger *logrus.Logger, meter metric.Meter) *Manager {
	metrics, err := initManagerMetrics(meter)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize manager metrics")
	}

	return &Manager{
		registry: registry,
		workers:  workers,
		config:   config,
		logger:   logger,
		metrics:  metrics,
	}
}

func initManagerMetrics(meter metric.Meter) (*ManagerMetrics, error) {
	activeWorkers, err := meter.Int64UpDownCounter("collector_active_workers")
	if err != nil {
		return nil, err
	}

	return &ManagerMetrics{
		activeWorkers: activeWorkers,
	}, nil
}

// Start starts the manager, which runs collection cycles at regular intervals
func (m *Manager) Start(ctx context.Context) {
	ctx, span := observability.StartSpan(ctx, "Manager.Start")
	defer span.End()

	ticker := time.NewTicker(m.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Manager stopping due to context cancellation")
			return
		case <-ticker.C:
			m.runCollection(ctx)
		}
	}
}

func (m *Manager) runCollection(ctx context.Context) {
	ctx, span := observability.StartSpan(ctx, "Manager.runCollection")
	defer span.End()

	startTime := time.Now()
	m.logger.Info("Starting collection cycle")

	datasources := m.registry.GetAll()
	jobs := make(chan Datasource, len(datasources))

	var wg sync.WaitGroup
	for _, worker := range m.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			m.metrics.activeWorkers.Add(ctx, 1)
			defer m.metrics.activeWorkers.Add(ctx, -1)
			for ds := range jobs {
				if err := w.CollectFromDatasource(ctx, ds); err != nil {
					m.logger.WithError(err).WithField("datasource_id", ds.GetID()).Error("Error collecting from datasource")
				}
			}
		}(worker)
	}

	for _, ds := range datasources {
		jobs <- ds
	}
	close(jobs)

	wg.Wait()

	duration := time.Since(startTime)
	m.logger.WithField("duration", duration).Info("Collection cycle completed")
	span.SetAttributes(attribute.Int64("collection_duration_ms", duration.Milliseconds()))
}
