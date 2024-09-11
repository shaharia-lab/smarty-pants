import axios, { AxiosInstance, CancelToken } from 'axios';
import {InteractionsResponse, Interaction} from '@/types/api';

export class ChatHistoriesApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getChatHistories(cancelToken?: CancelToken): Promise<InteractionsResponse> {
        const response = await this.axiosInstance.get<InteractionsResponse>('/api/v1/interactions', { cancelToken });
        return response.data;
    }

    async startNewSession(cancelToken?: CancelToken): Promise<Interaction> {
        const response = await this.axiosInstance.post<Interaction>(
            '/api/v1/interactions',
            { query: 'Start new session' },
            cancelToken ? { cancelToken } : {}
        );
        return response.data;
    }

    async getInteraction(id: string, cancelToken?: CancelToken): Promise<Interaction> {
        const response = await this.axiosInstance.get<Interaction>(`/api/v1/interactions/${id}`, { cancelToken });
        return response.data;
    }
}