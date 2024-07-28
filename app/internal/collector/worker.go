package collector

import (
	"context"
	"time"

	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Worker Config contains configuration options for the worker
type Worker struct {
	id      int
	storage storage.Storage
	logger  *logrus.Logger
	config  Config
	metrics *WorkerMetrics
}

// WorkerMetrics contains metrics for the worker
type WorkerMetrics struct {
	documentsCollected metric.Int64Counter
	collectionErrors   metric.Int64Counter
}

// NewWorker creates a new worker with the given ID, storage, logger, configuration, and meter
func NewWorker(id int, storage storage.Storage, logger *logrus.Logger, config Config, meter metric.Meter) *Worker {
	metrics, err := initWorkerMetrics(meter)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize worker metrics")
	}

	return &Worker{
		id:      id,
		storage: storage,
		logger:  logger,
		config:  config,
		metrics: metrics,
	}
}

func initWorkerMetrics(meter metric.Meter) (*WorkerMetrics, error) {
	documentsCollected, err := meter.Int64Counter("collector_documents_collected_total")
	if err != nil {
		return nil, err
	}

	collectionErrors, err := meter.Int64Counter("collector_collection_errors_total")
	if err != nil {
		return nil, err
	}

	return &WorkerMetrics{
		documentsCollected: documentsCollected,
		collectionErrors:   collectionErrors,
	}, nil
}

// CollectFromDatasource collects data from the given datasource
func (w *Worker) CollectFromDatasource(ctx context.Context, ds Datasource) error {
	ctx, span := observability.StartSpan(ctx, "Worker.CollectFromDatasource")
	defer span.End()

	logger := w.logger.WithFields(logrus.Fields{
		"worker_id":     w.id,
		"datasource_id": ds.GetID(),
	})

	logger.Info("Starting collection for datasource")
	startTime := time.Now()

	dds, err := w.storage.GetDatasource(ctx, ds.GetID())
	if err != nil {
		logger.WithError(err).Error("Failed to get state for datasource")
		w.metrics.collectionErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("datasource_id", ds.GetID().String())))
		return err
	}

	var documents []types.Document
	var newState types.DatasourceState
	var collectionErr error

	for attempt := 0; attempt < w.config.RetryAttempts; attempt++ {
		documents, newState, err = ds.GetData(ctx, dds.State)
		if err != nil {
			logger.WithError(err).Warnf("Failed to collect from datasource, attempt %d", attempt+1)
			time.Sleep(w.config.RetryDelay)
			collectionErr = err
			continue
		}

		collectionErr = nil
		break
	}

	if collectionErr != nil {
		w.metrics.collectionErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("datasource_id", ds.GetID().String())))
		logger.WithError(collectionErr).Error("Failed to collect data after all retries")
		return collectionErr
	}

	err = w.storage.UpdateDatasource(ctx, ds.GetID(), dds.Settings, newState)
	if err != nil {
		logger.WithError(err).Error("Failed to store state for datasource")
		w.metrics.collectionErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("datasource_id", ds.GetID().String())))
		return err
	}

	for _, doc := range documents {
		w.addAdditionalInfoForDoc(&doc, ds, dds)

		err = w.storage.Store(ctx, doc)
		if err != nil {
			logger.WithError(err).WithField("document_id", doc.UUID).Error("Failed to store document")
			w.metrics.collectionErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("datasource_id", ds.GetID().String())))
		} else {
			w.metrics.documentsCollected.Add(ctx, 1, metric.WithAttributes(attribute.String("datasource_id", ds.GetID().String())))
		}
	}

	duration := time.Since(startTime)
	logger.WithFields(logrus.Fields{
		"duration":            duration,
		"documents_collected": len(documents),
	}).Info("Completed collection for datasource")

	span.SetAttributes(
		attribute.Int64("collection_duration_ms", duration.Milliseconds()),
		attribute.Int("documents_collected", len(documents)),
	)

	return nil
}

func (w *Worker) addAdditionalInfoForDoc(doc *types.Document, ds Datasource, dds types.DatasourceConfig) {
	doc.Source = types.Source{
		UUID:       ds.GetID(),
		Name:       dds.Name,
		SourceType: dds.SourceType,
	}
	doc.FetchedAt = time.Now().UTC()
}
