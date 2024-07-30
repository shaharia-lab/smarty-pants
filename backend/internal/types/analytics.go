package types

import "time"

type AnalyticsOverview struct {
	EmbeddingProviders EmbeddingProvidersOverview `json:"embedding_providers"`
	LLMProviders       LLMProvidersOverview       `json:"llm_providers"`
	Datasources        DatasourcesOverview        `json:"datasources"`
}

type EmbeddingProvidersOverview struct {
	TotalProviders       int          `json:"total_providers"`
	TotalActiveProviders int          `json:"total_active_providers"`
	ActiveProvider       ProviderInfo `json:"active_provider"`
}

type LLMProvidersOverview struct {
	TotalProviders       int          `json:"total_providers"`
	TotalActiveProviders int          `json:"total_active_providers"`
	ActiveProvider       ProviderInfo `json:"active_provider"`
}

type ProviderInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Model string `json:"model"`
}

type DatasourcesOverview struct {
	ConfiguredDatasources                 []DatasourceInfo `json:"configured_datasources"`
	TotalDatasources                      int              `json:"total_datasources"`
	TotalDatasourcesByType                map[string]int   `json:"total_datasources_by_type"`
	TotalDatasourcesByStatus              map[string]int   `json:"total_datasources_by_status"`
	TotalDocumentsFetchedByDatasourceType map[string]int   `json:"total_documents_fetched_by_datasource_type"`
}

type DatasourceInfo struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
