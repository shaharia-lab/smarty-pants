import axios, { AxiosInstance, CancelToken } from 'axios';
import {Interaction, Message, PaginatedInteractionsResponse} from '@/types/api';

export class ChatHistoriesApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getChatHistories(cancelToken?: CancelToken): Promise<PaginatedInteractionsResponse> {
        const response = await this.axiosInstance.get<PaginatedInteractionsResponse>('/api/v1/interactions', { cancelToken });
        return response.data;
    }

    async startNewSession(message: Message, cancelToken?: CancelToken): Promise<Interaction> {
        const response = await this.axiosInstance.post<Interaction>(
            '/api/v1/interactions',
            message,
            cancelToken ? { cancelToken } : {}
        );
        return response.data;
    }

    async getInteraction(id: string, cancelToken?: CancelToken): Promise<Interaction> {
        const response = await this.axiosInstance.get<Interaction>(`/api/v1/interactions/${id}`, { cancelToken });
        return response.data;
    }

    async sendMessage(interactionId: string, message: Message, cancelToken?: CancelToken): Promise<Message> {
        const response = await this.axiosInstance.post<Message>(
            `/api/v1/interactions/${interactionId}/message`,
            message,
            cancelToken ? { cancelToken } : {}
        );
        return response.data;
    }
}