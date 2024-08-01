// Code generated by mockery v2.43.2. DO NOT EDIT.

package storage

import (
	context "context"

	migration "github.com/shaharia-lab/smarty-pants/backend/internal/storage/migration"
	mock "github.com/stretchr/testify/mock"

	sql "database/sql"

	types "github.com/shaharia-lab/smarty-pants/backend/internal/types"

	uuid "github.com/google/uuid"
)

// StorageMock is an autogenerated mock type for the StorageMock type
type StorageMock struct {
	mock.Mock
}

// AddConversationTx provides a mock function with given fields: ctx, tx, interactionUUID, conversation
func (_m *StorageMock) AddConversationTx(ctx context.Context, tx *sql.Tx, interactionUUID string, conversation types.Conversation) (types.Conversation, error) {
	ret := _m.Called(ctx, tx, interactionUUID, conversation)

	if len(ret) == 0 {
		panic("no return value specified for AddConversationTx")
	}

	var r0 types.Conversation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *sql.Tx, string, types.Conversation) (types.Conversation, error)); ok {
		return rf(ctx, tx, interactionUUID, conversation)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *sql.Tx, string, types.Conversation) types.Conversation); ok {
		r0 = rf(ctx, tx, interactionUUID, conversation)
	} else {
		r0 = ret.Get(0).(types.Conversation)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *sql.Tx, string, types.Conversation) error); ok {
		r1 = rf(ctx, tx, interactionUUID, conversation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddDatasource provides a mock function with given fields: ctx, dsConfig
func (_m *StorageMock) AddDatasource(ctx context.Context, dsConfig types.DatasourceConfig) error {
	ret := _m.Called(ctx, dsConfig)

	if len(ret) == 0 {
		panic("no return value specified for AddDatasource")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.DatasourceConfig) error); ok {
		r0 = rf(ctx, dsConfig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateEmbeddingProvider provides a mock function with given fields: ctx, provider
func (_m *StorageMock) CreateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error {
	ret := _m.Called(ctx, provider)

	if len(ret) == 0 {
		panic("no return value specified for CreateEmbeddingProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.EmbeddingProviderConfig) error); ok {
		r0 = rf(ctx, provider)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateInteraction provides a mock function with given fields: ctx, interaction
func (_m *StorageMock) CreateInteraction(ctx context.Context, interaction types.Interaction) (types.Interaction, error) {
	ret := _m.Called(ctx, interaction)

	if len(ret) == 0 {
		panic("no return value specified for CreateInteraction")
	}

	var r0 types.Interaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Interaction) (types.Interaction, error)); ok {
		return rf(ctx, interaction)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.Interaction) types.Interaction); ok {
		r0 = rf(ctx, interaction)
	} else {
		r0 = ret.Get(0).(types.Interaction)
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.Interaction) error); ok {
		r1 = rf(ctx, interaction)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateLLMProvider provides a mock function with given fields: ctx, provider
func (_m *StorageMock) CreateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error {
	ret := _m.Called(ctx, provider)

	if len(ret) == 0 {
		panic("no return value specified for CreateLLMProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.LLMProviderConfig) error); ok {
		r0 = rf(ctx, provider)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDatasource provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) DeleteDatasource(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteDatasource")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteEmbeddingProvider provides a mock function with given fields: ctx, id
func (_m *StorageMock) DeleteEmbeddingProvider(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteEmbeddingProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteLLMProvider provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) DeleteLLMProvider(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteLLMProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureMigrationTableExists provides a mock function with given fields:
func (_m *StorageMock) EnsureMigrationTableExists() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EnsureMigrationTableExists")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, filter, options
func (_m *StorageMock) Get(ctx context.Context, filter types.DocumentFilter, options types.DocumentFilterOption) (types.PaginatedDocuments, error) {
	ret := _m.Called(ctx, filter, options)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 types.PaginatedDocuments
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.DocumentFilter, types.DocumentFilterOption) (types.PaginatedDocuments, error)); ok {
		return rf(ctx, filter, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.DocumentFilter, types.DocumentFilterOption) types.PaginatedDocuments); ok {
		r0 = rf(ctx, filter, options)
	} else {
		r0 = ret.Get(0).(types.PaginatedDocuments)
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.DocumentFilter, types.DocumentFilterOption) error); ok {
		r1 = rf(ctx, filter, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllDatasources provides a mock function with given fields: ctx, page, perPage
func (_m *StorageMock) GetAllDatasources(ctx context.Context, page int, perPage int) (*types.PaginatedDatasources, error) {
	ret := _m.Called(ctx, page, perPage)

	if len(ret) == 0 {
		panic("no return value specified for GetAllDatasources")
	}

	var r0 *types.PaginatedDatasources
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, int) (*types.PaginatedDatasources, error)); ok {
		return rf(ctx, page, perPage)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, int) *types.PaginatedDatasources); ok {
		r0 = rf(ctx, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.PaginatedDatasources)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, int) error); ok {
		r1 = rf(ctx, page, perPage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllEmbeddingProviders provides a mock function with given fields: ctx, filter, option
func (_m *StorageMock) GetAllEmbeddingProviders(ctx context.Context, filter types.EmbeddingProviderFilter, option types.EmbeddingProviderFilterOption) (*types.PaginatedEmbeddingProviders, error) {
	ret := _m.Called(ctx, filter, option)

	if len(ret) == 0 {
		panic("no return value specified for GetAllEmbeddingProviders")
	}

	var r0 *types.PaginatedEmbeddingProviders
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.EmbeddingProviderFilter, types.EmbeddingProviderFilterOption) (*types.PaginatedEmbeddingProviders, error)); ok {
		return rf(ctx, filter, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.EmbeddingProviderFilter, types.EmbeddingProviderFilterOption) *types.PaginatedEmbeddingProviders); ok {
		r0 = rf(ctx, filter, option)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.PaginatedEmbeddingProviders)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.EmbeddingProviderFilter, types.EmbeddingProviderFilterOption) error); ok {
		r1 = rf(ctx, filter, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllInteractions provides a mock function with given fields: ctx, page, perPage
func (_m *StorageMock) GetAllInteractions(ctx context.Context, page int, perPage int) (*types.PaginatedInteractions, error) {
	ret := _m.Called(ctx, page, perPage)

	if len(ret) == 0 {
		panic("no return value specified for GetAllInteractions")
	}

	var r0 *types.PaginatedInteractions
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, int) (*types.PaginatedInteractions, error)); ok {
		return rf(ctx, page, perPage)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, int) *types.PaginatedInteractions); ok {
		r0 = rf(ctx, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.PaginatedInteractions)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, int) error); ok {
		r1 = rf(ctx, page, perPage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllLLMProviders provides a mock function with given fields: ctx, filter, option
func (_m *StorageMock) GetAllLLMProviders(ctx context.Context, filter types.LLMProviderFilter, option types.LLMProviderFilterOption) (*types.PaginatedLLMProviders, error) {
	ret := _m.Called(ctx, filter, option)

	if len(ret) == 0 {
		panic("no return value specified for GetAllLLMProviders")
	}

	var r0 *types.PaginatedLLMProviders
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.LLMProviderFilter, types.LLMProviderFilterOption) (*types.PaginatedLLMProviders, error)); ok {
		return rf(ctx, filter, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.LLMProviderFilter, types.LLMProviderFilterOption) *types.PaginatedLLMProviders); ok {
		r0 = rf(ctx, filter, option)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.PaginatedLLMProviders)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.LLMProviderFilter, types.LLMProviderFilterOption) error); ok {
		r1 = rf(ctx, filter, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAnalyticsOverview provides a mock function with given fields: ctx
func (_m *StorageMock) GetAnalyticsOverview(ctx context.Context) (types.AnalyticsOverview, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAnalyticsOverview")
	}

	var r0 types.AnalyticsOverview
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (types.AnalyticsOverview, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) types.AnalyticsOverview); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(types.AnalyticsOverview)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetConversation provides a mock function with given fields: ctx, interactionUUID
func (_m *StorageMock) GetConversation(ctx context.Context, interactionUUID uuid.UUID) ([]types.Conversation, error) {
	ret := _m.Called(ctx, interactionUUID)

	if len(ret) == 0 {
		panic("no return value specified for GetConversation")
	}

	var r0 []types.Conversation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]types.Conversation, error)); ok {
		return rf(ctx, interactionUUID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []types.Conversation); ok {
		r0 = rf(ctx, interactionUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Conversation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, interactionUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCurrentVersion provides a mock function with given fields:
func (_m *StorageMock) GetCurrentVersion() (string, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCurrentVersion")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatasource provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) GetDatasource(ctx context.Context, _a1 uuid.UUID) (types.DatasourceConfig, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetDatasource")
	}

	var r0 types.DatasourceConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (types.DatasourceConfig, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) types.DatasourceConfig); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(types.DatasourceConfig)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetEmbeddingProvider provides a mock function with given fields: ctx, id
func (_m *StorageMock) GetEmbeddingProvider(ctx context.Context, id uuid.UUID) (*types.EmbeddingProviderConfig, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetEmbeddingProvider")
	}

	var r0 *types.EmbeddingProviderConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*types.EmbeddingProviderConfig, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *types.EmbeddingProviderConfig); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.EmbeddingProviderConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForProcessing provides a mock function with given fields: ctx, filter, batchLimit
func (_m *StorageMock) GetForProcessing(ctx context.Context, filter types.DocumentFilter, batchLimit int) ([]uuid.UUID, error) {
	ret := _m.Called(ctx, filter, batchLimit)

	if len(ret) == 0 {
		panic("no return value specified for GetForProcessing")
	}

	var r0 []uuid.UUID
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.DocumentFilter, int) ([]uuid.UUID, error)); ok {
		return rf(ctx, filter, batchLimit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.DocumentFilter, int) []uuid.UUID); ok {
		r0 = rf(ctx, filter, batchLimit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]uuid.UUID)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.DocumentFilter, int) error); ok {
		r1 = rf(ctx, filter, batchLimit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInteraction provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) GetInteraction(ctx context.Context, _a1 uuid.UUID) (types.Interaction, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetInteraction")
	}

	var r0 types.Interaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (types.Interaction, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) types.Interaction); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(types.Interaction)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLLMProvider provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) GetLLMProvider(ctx context.Context, _a1 uuid.UUID) (*types.LLMProviderConfig, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetLLMProvider")
	}

	var r0 *types.LLMProviderConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*types.LLMProviderConfig, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *types.LLMProviderConfig); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.LLMProviderConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSettings provides a mock function with given fields: ctx
func (_m *StorageMock) GetSettings(ctx context.Context) (types.Settings, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetSettings")
	}

	var r0 types.Settings
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (types.Settings, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) types.Settings); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(types.Settings)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HealthCheck provides a mock function with given fields:
func (_m *StorageMock) HealthCheck() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HealthCheck")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Migrate provides a mock function with given fields: migrations
func (_m *StorageMock) Migrate(migrations []migration.Migration) error {
	ret := _m.Called(migrations)

	if len(ret) == 0 {
		panic("no return value specified for Migrate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]migration.Migration) error); ok {
		r0 = rf(migrations)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordAIOpsUsage provides a mock function with given fields: ctx, usage
func (_m *StorageMock) RecordAIOpsUsage(ctx context.Context, usage types.AIUsage) error {
	ret := _m.Called(ctx, usage)

	if len(ret) == 0 {
		panic("no return value specified for RecordAIOpsUsage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.AIUsage) error); ok {
		r0 = rf(ctx, usage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rollback provides a mock function with given fields: migrations
func (_m *StorageMock) Rollback(migrations []migration.Migration) error {
	ret := _m.Called(migrations)

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]migration.Migration) error); ok {
		r0 = rf(migrations)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Search provides a mock function with given fields: ctx, config
func (_m *StorageMock) Search(ctx context.Context, config types.SearchConfig) (*types.SearchResults, error) {
	ret := _m.Called(ctx, config)

	if len(ret) == 0 {
		panic("no return value specified for Search")
	}

	var r0 *types.SearchResults
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.SearchConfig) (*types.SearchResults, error)); ok {
		return rf(ctx, config)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.SearchConfig) *types.SearchResults); ok {
		r0 = rf(ctx, config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.SearchResults)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.SearchConfig) error); ok {
		r1 = rf(ctx, config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetActiveDatasource provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) SetActiveDatasource(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SetActiveDatasource")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetActiveEmbeddingProvider provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) SetActiveEmbeddingProvider(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SetActiveEmbeddingProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetActiveLLMProvider provides a mock function with given fields: ctx, id
func (_m *StorageMock) SetActiveLLMProvider(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for SetActiveLLMProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDisableDatasource provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) SetDisableDatasource(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SetDisableDatasource")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDisableEmbeddingProvider provides a mock function with given fields: ctx, _a1
func (_m *StorageMock) SetDisableEmbeddingProvider(ctx context.Context, _a1 uuid.UUID) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for SetDisableEmbeddingProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDisableLLMProvider provides a mock function with given fields: ctx, id
func (_m *StorageMock) SetDisableLLMProvider(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for SetDisableLLMProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields: ctx, document
func (_m *StorageMock) Store(ctx context.Context, document types.Document) error {
	ret := _m.Called(ctx, document)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Document) error); ok {
		r0 = rf(ctx, document)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, document
func (_m *StorageMock) Update(ctx context.Context, document types.Document) error {
	ret := _m.Called(ctx, document)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Document) error); ok {
		r0 = rf(ctx, document)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDatasource provides a mock function with given fields: ctx, _a1, settings, state
func (_m *StorageMock) UpdateDatasource(ctx context.Context, _a1 uuid.UUID, settings types.DatasourceSettings, state types.DatasourceState) error {
	ret := _m.Called(ctx, _a1, settings, state)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDatasource")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, types.DatasourceSettings, types.DatasourceState) error); ok {
		r0 = rf(ctx, _a1, settings, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateEmbeddingProvider provides a mock function with given fields: ctx, provider
func (_m *StorageMock) UpdateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error {
	ret := _m.Called(ctx, provider)

	if len(ret) == 0 {
		panic("no return value specified for UpdateEmbeddingProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.EmbeddingProviderConfig) error); ok {
		r0 = rf(ctx, provider)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateLLMProvider provides a mock function with given fields: ctx, provider
func (_m *StorageMock) UpdateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error {
	ret := _m.Called(ctx, provider)

	if len(ret) == 0 {
		panic("no return value specified for UpdateLLMProvider")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.LLMProviderConfig) error); ok {
		r0 = rf(ctx, provider)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSettings provides a mock function with given fields: ctx, settings
func (_m *StorageMock) UpdateSettings(ctx context.Context, settings types.Settings) error {
	ret := _m.Called(ctx, settings)

	if len(ret) == 0 {
		panic("no return value specified for UpdateSettings")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Settings) error); ok {
		r0 = rf(ctx, settings)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewStorage creates a new instance of StorageMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *StorageMock {
	mock := &StorageMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
