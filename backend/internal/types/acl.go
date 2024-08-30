package types

const (
	APiAccessOpsDatasourceAdd        = "datasource:add"
	APiAccessOpsDatasourceGet        = "datasource:get"
	APiAccessOpsDatasourceUpdate     = "datasource:update"
	APiAccessOpsDatasourceActivate   = "datasource:activate"
	APiAccessOpsDatasourceDeactivate = "datasource:deactivate"
	APiAccessOpsDatasourceDelete     = "datasource:delete"

	APiAccessOpsDocumentsGet = "documents:get"
	APiAccessOpsDocumentGet  = "document:get"

	APIAccessOpsEmbeddingProvidersGet       = "embedding_providers:get"
	APIAccessOpsEmbeddingProviderGet        = "embedding_provider:get"
	APIAccessOpsEmbeddingProviderDelete     = "embedding_provider:delete"
	APIAccessOpsEmbeddingProviderUpdate     = "embedding_provider:update"
	APIAccessOpsEmbeddingProviderActivate   = "embedding_provider:activate"
	APIAccessOpsEmbeddingProviderDeactivate = "embedding_provider:deactivate"

	APIAccessOpsLLMProvidersGet       = "llm_providers:get"
	APIAccessOpsLLMProviderGet        = "llm_provider:get"
	APIAccessOpsLLMProviderDelete     = "llm_provider:delete"
	APIAccessOpsLLMProviderUpdate     = "llm_provider:update"
	APIAccessOpsLLMProviderActivate   = "llm_provider:activate"
	APIAccessOpsLLMProviderDeactivate = "llm_provider:deactivate"

	APIAccessOpsInteractionCreate = "interaction:create"
	APIAccessOpsInteractionsGet   = "interactions:get"
	APIAccessOpsInteractionGet    = "interaction:get"
	APIAccessOpsMessageSend       = "message:send"
)
