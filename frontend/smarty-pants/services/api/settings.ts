import { AxiosInstance, CancelToken } from 'axios';

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

export class SettingsApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getSettings(cancelToken?: CancelToken): Promise<Settings> {
        const response = await this.axiosInstance.get<Settings>('/api/v1/settings', { cancelToken });
        return response.data;
    }

    async updateSettings(settings: Settings, cancelToken?: CancelToken): Promise<Settings> {
        const response = await this.axiosInstance.put<Settings>('/api/v1/settings', settings, { cancelToken });
        return response.data;
    }
}