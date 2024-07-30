// File: /types/embeddingProvider.ts

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