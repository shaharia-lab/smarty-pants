package storage

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
)

type DatasourceConfig struct {
	UUID       string                 `json:"uuid"`
	Name       string                 `json:"name"`
	SourceType string                 `json:"source_type"`
	Settings   map[string]interface{} `json:"settings"`
	Status     string                 `json:"status"`
	State      types.State            `json:"state"`
}

type Storage interface {
	HealthCheck() error

	Store(ctx context.Context, document types.Document) error
	Update(ctx context.Context, document types.Document) error
	Get(ctx context.Context, filter types.DocumentFilter, options types.DocumentFilterOption) (types.PaginatedDocuments, error)
	GetForProcessing(ctx context.Context, filter types.DocumentFilter, batchLimit int) ([]uuid.UUID, error)

	AddDatasource(ctx context.Context, dsConfig types.DatasourceConfig) error
	GetDatasource(ctx context.Context, uuid uuid.UUID) (types.DatasourceConfig, error)
	GetAllDatasources(ctx context.Context, page, perPage int) (*types.PaginatedDatasources, error)
	UpdateDatasource(ctx context.Context, uuid uuid.UUID, settings types.DatasourceSettings, state types.DatasourceState) error
	SetActiveDatasource(ctx context.Context, uuid uuid.UUID) error
	SetDisableDatasource(ctx context.Context, uuid uuid.UUID) error
	DeleteDatasource(ctx context.Context, uuid uuid.UUID) error

	CreateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error
	UpdateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error
	DeleteEmbeddingProvider(ctx context.Context, id uuid.UUID) error
	GetEmbeddingProvider(ctx context.Context, id uuid.UUID) (*types.EmbeddingProviderConfig, error)
	GetAllEmbeddingProviders(ctx context.Context, filter types.EmbeddingProviderFilter, option types.EmbeddingProviderFilterOption) (*types.PaginatedEmbeddingProviders, error)
	SetActiveEmbeddingProvider(ctx context.Context, uuid uuid.UUID) error
	SetDisableEmbeddingProvider(ctx context.Context, uuid uuid.UUID) error

	CreateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error
	UpdateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error
	DeleteLLMProvider(ctx context.Context, uuid uuid.UUID) error
	GetLLMProvider(ctx context.Context, uuid uuid.UUID) (*types.LLMProviderConfig, error)
	GetAllLLMProviders(ctx context.Context, filter types.LLMProviderFilter, option types.LLMProviderFilterOption) (*types.PaginatedLLMProviders, error)
	SetActiveLLMProvider(ctx context.Context, id uuid.UUID) error
	SetDisableLLMProvider(ctx context.Context, id uuid.UUID) error

	RecordAIOpsUsage(ctx context.Context, usage types.AIUsage) error

	Search(ctx context.Context, config types.SearchConfig) (*types.SearchResults, error)

	UpdateSettings(ctx context.Context, settings types.Settings) error
	GetSettings(ctx context.Context) (types.Settings, error)

	CreateInteraction(ctx context.Context, interaction types.Interaction) (types.Interaction, error)
	GetInteraction(ctx context.Context, uuid uuid.UUID) (types.Interaction, error)
	GetAllInteractions(ctx context.Context, page, perPage int) (*types.PaginatedInteractions, error)
	AddConversationTx(ctx context.Context, tx *sql.Tx, interactionUUID string, conversation types.Conversation) (types.Conversation, error)
	GetConversation(ctx context.Context, interactionUUID uuid.UUID) ([]types.Conversation, error)

	GetAnalyticsOverview(ctx context.Context) (types.AnalyticsOverview, error)

	RunMigration() error
	HandleShutdown(ctx context.Context) error

	GetKeyPair() (privateKey, publicKey []byte, err error)
	UpdateKeyPair(privateKey, publicKey []byte) error
}
