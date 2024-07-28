package processor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, 5, config.WorkerCount)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 10*time.Second, config.ProcessInterval)

}

func TestEmbeddingProcess(t *testing.T) {
	mockProvider := new(embedding.EmbeddingProviderMock)
	logging := logrus.New()

	em := NewEmbedding(mockProvider, logging)

	ctx := context.Background()
	doc := &types.Document{
		UUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Body: "test body",
	}

	mockProvider.On("GetEmbedding", ctx, doc.Body).Return([]types.ContentPart{
		{Content: "test1", Embedding: []float32{0.1, 0.2, 0.3}},
	}, nil)

	processedDoc, err := em.Process(ctx, doc)

	assert.NoError(t, err)
	assert.Equal(t, []types.ContentPart{
		{Content: "test1", Embedding: []float32{0.1, 0.2, 0.3}},
	}, processedDoc.Embedding.Embedding)
	mockProvider.AssertExpectations(t)
}

func TestManagerStart(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	registry := NewRegistry(mockStorage, logrus.New())
	config := DefaultConfig()
	config.ProcessInterval = 100 * time.Millisecond
	meter := noop.NewMeterProvider().Meter("")

	manager := NewManager(mockStorage, registry, config, logger.NoOpsLogger(), meter)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockStorage.On("GetForProcessing", mock.Anything, mock.Anything, mock.Anything).Return([]uuid.UUID{
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
	}, nil)

	mockStorage.On("Get", mock.Anything, types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"}, types.DocumentFilterOption{Limit: 1, Page: 1}).Return(types.PaginatedDocuments{}, nil)

	go manager.Start(ctx)

	time.Sleep(200 * time.Millisecond)
	mockStorage.AssertCalled(t, "GetForProcessing", mock.Anything, mock.Anything, mock.Anything)
	mockStorage.AssertCalled(t, "Get", mock.Anything, types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"}, types.DocumentFilterOption{Limit: 1, Page: 1})
}

func TestProcessorStart(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	mockStorage.On("GetAllEmbeddingProviders", mock.Anything, mock.Anything, mock.Anything).Return(&types.PaginatedEmbeddingProviders{}, nil)

	config := DefaultConfig()
	meter := noop.NewMeterProvider().Meter("")

	proc, err := NewProcessor(config, mockStorage, logger.NoOpsLogger(), meter)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockStorage.On("Get", mock.Anything, mock.Anything).Return(&types.PaginatedDocuments{}, nil)

	err = proc.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	proc.Stop()
}

func TestRegistryRegisterAndGet(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	registry := NewRegistry(mockStorage, logger.NoOpsLogger())

	mockProcessor := new(embedding.EmbeddingProviderMock)
	err := registry.Register("test", mockProcessor)
	assert.NoError(t, err)

	retrievedProcessor, exists := registry.Get("test")
	assert.True(t, exists)
	assert.Equal(t, mockProcessor, retrievedProcessor)
}

func TestWorkerProcessDocument(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	registry := NewRegistry(mockStorage, logrus.New())
	config := DefaultConfig()
	meter := noop.NewMeterProvider().Meter("")

	worker, err := NewWorker(1, mockStorage, registry, logger.NoOpsLogger(), config, meter)
	assert.NoError(t, err)

	ctx := context.Background()
	docUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	mockStorage.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(types.PaginatedDocuments{
		Documents: []types.Document{{UUID: docUUID, Body: "test body"}},
	}, nil)
	mockStorage.On("Update", mock.Anything, mock.Anything).Return(nil)

	err = worker.ProcessDocument(ctx, docUUID)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}
