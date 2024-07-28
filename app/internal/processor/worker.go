package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Worker is a processing worker that processes documents
type Worker struct {
	id       int
	storage  storage.Storage
	registry *Registry
	logger   *logrus.Logger
	config   Config
	metrics  *WorkerMetrics
}

// WorkerMetrics contains metrics for the worker
type WorkerMetrics struct {
	documentsProcessed   metric.Int64Counter
	processingDuration   metric.Int64Histogram
	processingErrors     metric.Int64Counter
	processorErrors      metric.Int64Counter
	documentsPartialFail metric.Int64Counter
}

// NewWorker creates a new worker with the given settings
func NewWorker(id int, storage storage.Storage, registry *Registry, logger *logrus.Logger, config Config, meter metric.Meter) (*Worker, error) {
	metrics, err := initWorkerMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize worker metrics: %w", err)
	}

	return &Worker{
		id:       id,
		storage:  storage,
		registry: registry,
		logger:   logger,
		config:   config,
		metrics:  metrics,
	}, nil
}

// ProcessDocument processes a document
func (w *Worker) ProcessDocument(ctx context.Context, docUUID uuid.UUID) error {
	ctx, span := observability.StartSpan(ctx, "Worker.ProcessDocument")
	defer span.End()

	logger := w.logger.WithFields(logrus.Fields{
		"worker_id":   w.id,
		"document_id": docUUID,
	})

	startTime := time.Now()
	logger.Info("Starting document processing")

	doc, err := w.fetchDocument(ctx, docUUID)
	if err != nil {
		return err
	}

	processors := w.registry.GetAll()

	logger.WithField("total_processors", len(processors)).Info("Total number of processors in the registry")

	processingErrors := w.runProcessors(ctx, doc, processors)

	logger.WithField("processing_errors", processingErrors).Info("Document processing completed. Updating status")
	w.updateDocumentStatus(ctx, doc, processingErrors, processors)

	if err := w.storage.Update(ctx, *doc); err != nil {
		logger.WithError(err).Error("Failed to update document status")
		w.metrics.processingErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("error", "update_document")))
		return fmt.Errorf("failed to update document status: %w", err)
	}

	w.recordMetrics(ctx, startTime, doc.Status)

	span.SetAttributes(
		attribute.Int64("processing_duration_ms", time.Since(startTime).Milliseconds()),
		attribute.String("document_status", string(doc.Status)),
	)

	return nil
}

func (w *Worker) fetchDocument(ctx context.Context, docUUID uuid.UUID) (*types.Document, error) {
	logger := w.logger.WithField("document_id", docUUID)
	logger.Info("Fetching document from storage")

	paginatedDocs, err := w.storage.Get(ctx, types.DocumentFilter{UUID: docUUID.String()}, types.DocumentFilterOption{Limit: 1, Page: 1})
	if err != nil {
		logger.WithError(err).Error("Failed to fetch document")
		w.metrics.processingErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("error", "fetch_document")))
		return nil, fmt.Errorf("failed to fetch document: %w", err)
	}

	if len(paginatedDocs.Documents) == 0 {
		logger.Error("No document found")
		w.metrics.processingErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("error", "document_not_found")))
		return nil, fmt.Errorf("document not found")
	}

	return &paginatedDocs.Documents[0], nil
}

func (w *Worker) runProcessors(ctx context.Context, doc *types.Document, processors []ProcessorUnit) []error {
	var processingErrors []error
	for _, processor := range processors {
		processorName := fmt.Sprintf("%T", processor)

		logger := w.logger.WithFields(logrus.Fields{
			"processor":   processorName,
			"document_id": doc.UUID,
		})

		logger.Info("Running processor")

		if err := processor.Process(ctx, doc); err != nil {
			processingErrors = append(processingErrors, fmt.Errorf("%s: %w", processorName, err))
			logger.WithError(err).Error("Processor failed")
			w.metrics.processorErrors.Add(ctx, 1, metric.WithAttributes(
				attribute.String("processor", processorName),
				attribute.String("error", err.Error()),
			))
		} else {
			logger.Info("Processor completed successfully")
		}
	}
	return processingErrors
}

func (w *Worker) updateDocumentStatus(ctx context.Context, doc *types.Document, processingErrors []error, processors []ProcessorUnit) {
	logger := w.logger.WithField("document_id", doc.UUID)

	if len(processors) == 0 {
		doc.Status = types.DocumentStatusPending
		logger.Error("No processors found")
		w.metrics.processingErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("error", "no_processors")))
		return
	}

	if len(processingErrors) == 0 {
		doc.Status = types.DocumentStatusReadyToSearch
		logger.Info("Document processing completed successfully")
		return
	}

	doc.Status = types.DocumentStatusErrorProcessing
	logger.WithField("errors", processingErrors).Error("Document processing failed completely")
}

func (w *Worker) recordMetrics(ctx context.Context, startTime time.Time, status types.DocumentStatus) {
	duration := time.Since(startTime)
	w.logger.WithFields(logrus.Fields{
		"duration": duration,
		"status":   status,
	}).Info("Completed document processing")

	w.metrics.documentsProcessed.Add(ctx, 1)
	w.metrics.processingDuration.Record(ctx, duration.Milliseconds(), metric.WithAttributes(attribute.String("status", string(status))))
}

func initWorkerMetrics(meter metric.Meter) (*WorkerMetrics, error) {
	documentsProcessed, err := meter.Int64Counter("processor_documents_processed_total")
	if err != nil {
		return nil, fmt.Errorf("failed to create documents processed counter: %w", err)
	}

	processingDuration, err := meter.Int64Histogram("processor_processing_duration_milliseconds")
	if err != nil {
		return nil, fmt.Errorf("failed to create processing duration histogram: %w", err)
	}

	processingErrors, err := meter.Int64Counter("processor_processing_errors_total")
	if err != nil {
		return nil, fmt.Errorf("failed to create processing errors counter: %w", err)
	}

	processorErrors, err := meter.Int64Counter("processor_individual_errors_total")
	if err != nil {
		return nil, fmt.Errorf("failed to create processor errors counter: %w", err)
	}

	documentsPartialFail, err := meter.Int64Counter("processor_documents_partial_fail_total")
	if err != nil {
		return nil, fmt.Errorf("failed to create documents partial fail counter: %w", err)
	}

	return &WorkerMetrics{
		documentsProcessed:   documentsProcessed,
		processingDuration:   processingDuration,
		processingErrors:     processingErrors,
		processorErrors:      processorErrors,
		documentsPartialFail: documentsPartialFail,
	}, nil
}
