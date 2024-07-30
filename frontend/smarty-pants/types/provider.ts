// types/provider.ts

export interface ProviderConfig {
    uuid: string;
    name: string;
    provider: string;
    status: 'active' | 'inactive';
    configuration: {
        api_key: string;
        model_id: string;
        [key: string]: any;
    };
}

export interface AvailableProvider {
    id: string;
    name: string;
    imageUrl: string;
    description: string;
}