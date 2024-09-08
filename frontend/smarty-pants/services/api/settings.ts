import { AxiosInstance, CancelToken } from 'axios';

import { Settings } from '@/types/api';

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