export interface AnalyticsOverview {
    embedding_providers: {
        total_providers: number;
        total_active_providers: number;
        active_provider: {
            name: string;
            type: string;
            model: string;
        };
    };
    llm_providers: {
        total_providers: number;
        total_active_providers: number;
        active_provider: {
            name: string;
            type: string;
            model: string;
        };
    };
    datasources: {
        configured_datasources: Array<{
            name: string;
            type: string;
            status: string;
            created_at: string;
        }> | null;
        total_datasources: number;
        total_datasources_by_type: { [key: string]: number };
        total_datasources_by_status: { [key: string]: number };
        total_documents_fetched_by_datasource_type: { [key: string]: number };
    };
}