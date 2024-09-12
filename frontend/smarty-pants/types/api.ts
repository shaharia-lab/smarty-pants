import {LLMProviderConfig} from "@/types/llmProvider";

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

export interface InteractionSummary {
    uuid: string;
    title: string;
}

export interface InteractionsResponse {
    interactions: InteractionSummary[];
    limit: number;
    per_page: number;
}

export interface Message {
    role: 'system' | 'user';
    text: string;
}

export interface Interaction {
    uuid: string;
    query: string;
    conversations: Message[];
}

export interface LLMProvidersApiResponse {
    llm_providers: LLMProviderConfig[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}

export interface Settings {
    general: {
        application_name: string;
    };
    debugging: {
        log_level: 'debug' | 'info' | 'warn' | 'error';
        log_format: 'json' | 'text';
        log_output: 'stdout' | 'stderr' | 'file';
    };
    search: {
        per_page: number;
    };
}

export interface SystemInfo {
    version: string;
    app: SystemApp;
    settings: SystemSettings;
}

export interface SystemApp {
    name: string;
}

export interface SystemSettings {
    auth_enabled: boolean;
    oauth_providers: string[];
}

export interface GenerateResponseMsg {
    message: string;
}