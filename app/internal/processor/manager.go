package processor

import (
	"context"
	"sync"
	"time"

	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Manager is a processor that manages workers to process documents
type Manager struct {
	workers  []*Worker
	storage  storage.Storage
	registry *Registry
	config   Config
	logger   *logrus.Logger
	metrics  *ManagerMetrics
}

// ManagerMetrics contains metrics for the manager
type ManagerMetrics struct {
	activeWorkers    metric.Int64UpDownCounter
	processingCycles metric.Int64Counter
	cyclesDuration   metric.Int64Histogram
}

// NewManager creates a new manager with the given settings
func NewManager(storage storage.Storage, registry *Registry, config Config, logger *logrus.Logger, meter metric.Meter) *Manager {
	workers := make([]*Worker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		worker, err := NewWorker(i, storage, registry, logger, config, meter)
		if err != nil {
			logger.WithError(err).Error("Failed to create worker")
			return nil
		}
		workers[i] = worker
	}

	metrics, err := initManagerMetrics(meter)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize manager metrics")
	}

	return &Manager{
		workers:  workers,
		storage:  storage,
		registry: registry,
		config:   config,
		logger:   logger,
		metrics:  metrics,
	}
}

// Start starts the manager
func (m *Manager) Start(ctx context.Context) {
	ctx, span := observability.StartSpan(ctx, "Manager.Start")
	defer span.End()

	ticker := time.NewTicker(m.config.ProcessInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Manager stopping due to context cancellation")
			return
		case <-ticker.C:
			m.runProcessingCycle(ctx)
		}
	}
}

func (m *Manager) runProcessingCycle(ctx context.Context) {
	ctx, span := observability.StartSpan(ctx, "Manager.runProcessingCycle")
	defer span.End()

	startTime := time.Now()
	m.logger.Info("Starting processing cycle")

	docUUIDs, err := m.storage.GetForProcessing(ctx, types.DocumentFilter{Status: types.DocumentStatusPending}, m.config.BatchSize)
	if err != nil {
		m.logger.WithError(err).Error("Failed to fetch documents for processing")
		return
	}

	m.metrics.processingCycles.Add(ctx, 1)

	var wg sync.WaitGroup
	for _, worker := range m.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			m.metrics.activeWorkers.Add(ctx, 1)
			defer m.metrics.activeWorkers.Add(ctx, -1)
			for _, docUUID := range docUUIDs {
				if err := w.ProcessDocument(ctx, docUUID); err != nil {
					m.logger.WithError(err).WithField("document_id", docUUID).Error("Error processing document")
				}
			}
		}(worker)
	}

	wg.Wait()

	duration := time.Since(startTime)
	m.logger.WithField("duration", duration).Info("Processing cycle completed")
	m.metrics.cyclesDuration.Record(ctx, duration.Milliseconds())

	span.SetAttributes(attribute.Int64("cycle_duration_ms", duration.Milliseconds()))
}

func initManagerMetrics(meter metric.Meter) (*ManagerMetrics, error) {
	activeWorkers, err := meter.Int64UpDownCounter("processor_active_workers")
	if err != nil {
		return nil, err
	}

	processingCycles, err := meter.Int64Counter("processor_processing_cycles_total")
	if err != nil {
		return nil, err
	}

	cyclesDuration, err := meter.Int64Histogram("processor_cycles_duration_milliseconds")
	if err != nil {
		return nil, err
	}

	return &ManagerMetrics{
		activeWorkers:    activeWorkers,
		processingCycles: processingCycles,
		cyclesDuration:   cyclesDuration,
	}, nil
}
