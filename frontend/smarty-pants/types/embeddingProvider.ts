
export interface EmbeddingProviderConfig {
    uuid: string;
    name: string;
    provider: string;
    status: string;
    configuration: {
        api_key: string;
        model_id: string;
    };
}

export interface EmbeddingProvidersApiResponse {
    embedding_providers: EmbeddingProviderConfig[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}