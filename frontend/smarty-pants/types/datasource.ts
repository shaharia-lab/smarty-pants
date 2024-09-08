export interface DatasourceConfig {
    uuid: string;
    name: string;
    source_type: string;
    status: 'active' | 'inactive';
    settings: {
        [key: string]: any;
    };
    state?: {
        [key: string]: any;
    };
}

// You can also create type guards or type predicates for specific datasource types if needed
export function isSlackDatasource(datasource: DatasourceConfig): boolean {
    return datasource.source_type === 'slack' &&
        'workspace' in datasource.settings &&
        'token' in datasource.settings;
}

export interface AvailableDatasource {
    id: string;
    name: string;
    description: string;
    imageUrl: string;
    configurationUrl: string;
}

export interface AvailableDatasource {
    id: string;
    name: string;
    imageUrl: string;
    description: string;
    configurationUrl: string;
}

// Types for API responses
export interface DatasourceSettings {
    [key: string]: string | number | boolean;
}

export interface DatasourceState {
    type: string;
    next_cursor: string;
}

// Types for API payloads
export interface DatasourcePayload {
    name: string;
    source_type: string;
    settings: DatasourceSettings;
}

// Specific payload type for Slack
export interface SlackDatasourcePayload extends DatasourcePayload {
    source_type: 'slack';
    settings: {
        workspace: string;
        token: string;
        channel_id?: string;
    }
}