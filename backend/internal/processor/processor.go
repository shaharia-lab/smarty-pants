package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/storage"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/metric"
)

// Processor is a processor that processes documents
type Processor struct {
	config   Config
	storage  storage.Storage
	registry *Registry
	manager  *Manager
	logger   *logrus.Logger
	metrics  *ProcessorMetrics
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

// ProcessorMetrics contains metrics for the processor
type ProcessorMetrics struct {
	documentsProcessed metric.Int64Counter
	processingDuration metric.Int64Histogram
	processingErrors   metric.Int64Counter
	documentsInQueue   metric.Int64UpDownCounter
}

// NewProcessor creates a new processor with the given settings
func NewProcessor(config Config, storage storage.Storage, logger *logrus.Logger, meter metric.Meter) (*Processor, error) {
	registry := NewRegistry(storage, logger)
	manager := NewManager(storage, registry, config, logger, meter)

	metrics, err := initProcessorMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize processor metrics: %w", err)
	}

	return &Processor{
		config:   config,
		storage:  storage,
		registry: registry,
		manager:  manager,
		logger:   logger,
		metrics:  metrics,
		stopCh:   make(chan struct{}),
	}, nil
}

// Start starts the processor
func (p *Processor) Start(ctx context.Context) error {
	ctx, span := observability.StartSpan(ctx, "Processor.Start")
	defer span.End()

	p.logger.Info("Starting processor")

	if err := p.registry.RefreshProcessors(ctx); err != nil {
		return fmt.Errorf("failed to refresh processors: %w", err)
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.manager.Start(ctx)
	}()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.config.ProcessorRefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopCh:
				return
			case <-ticker.C:
				if err := p.registry.RefreshProcessors(ctx); err != nil {
					p.logger.WithError(err).Error("Failed to refresh processors")
				}
			}
		}
	}()

	return nil
}

// Stop stops the processor
func (p *Processor) Stop() {
	ctx, span := observability.StartSpan(context.Background(), "Processor.Stop")
	defer span.End()

	p.logger.Info("Stopping processor")
	close(p.stopCh)

	timeoutCtx, cancel := context.WithTimeout(ctx, p.config.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("Processor stopped gracefully")
	case <-timeoutCtx.Done():
		p.logger.Warn("Processor stop timed out, forcing shutdown")
	}
}

func initProcessorMetrics(meter metric.Meter) (*ProcessorMetrics, error) {
	documentsProcessed, err := meter.Int64Counter("processor_documents_processed_total")
	if err != nil {
		return nil, err
	}

	processingDuration, err := meter.Int64Histogram("processor_processing_duration_milliseconds")
	if err != nil {
		return nil, err
	}

	processingErrors, err := meter.Int64Counter("processor_processing_errors_total")
	if err != nil {
		return nil, err
	}

	documentsInQueue, err := meter.Int64UpDownCounter("processor_documents_in_queue")
	if err != nil {
		return nil, err
	}

	return &ProcessorMetrics{
		documentsProcessed: documentsProcessed,
		processingDuration: processingDuration,
		processingErrors:   processingErrors,
		documentsInQueue:   documentsInQueue,
	}, nil
}
