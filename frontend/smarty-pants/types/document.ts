export interface Document {
    uuid: string;
    title: string;
    body: string;
    embedding: Embedding;
    metadata: Metadata[];
    status: 'pending' | 'ready_to_search' | 'error_processing';
    created_at: string;
    updated_at: string;
    source: Source;
}

export interface Embedding {
    embedding: ContentPart[];
}

export interface ContentPart {
    content: string;
    embedding: number[];
}

export interface Metadata {
    key: string;
    value: string;
}

export interface Source {
    uuid: string;
    name: string;
    type: string;
}


export interface DatasourceConfig {
    uuid: string;
    name: string;
    status: string;
    source_type: string;
    settings: SlackSettings | GitHubSettings;
    state: SlackState | GitHubState;
}

export interface SlackSettings {
    token: string;
    channel_id: string;
}

export interface GitHubSettings {
    org: string;
}

export interface SlackState {
    type: 'slack';
    next_cursor: string;
}

export interface GitHubState {
    type: 'github';
    next_cursor: string;
}