export interface EmbeddingItem {
    content: string;
    embedding: number[];
}

export interface Metadata {
    key: string;
    value: string;
}

export interface Document {
    uuid: string;
    title: string;
    body: string;
    embedding: {
        embedding: EmbeddingItem[] | null;
    };
    metadata: Metadata[];
    status: string;
    created_at: string;
    updated_at: string;
}
