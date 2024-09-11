import { AxiosInstance, CancelToken } from 'axios';
import { LLMProviderConfig } from '@/types/llmProvider';
import {LLMProvidersApiResponse} from "@/types/api";

export class LLMProviderApi {
    constructor(private axiosInstance: AxiosInstance) {}

    getLLMProviders(cancelToken?: CancelToken): Promise<LLMProvidersApiResponse> {
        return this.axiosInstance.get<LLMProvidersApiResponse>('/api/v1/llm-provider', { cancelToken })
            .then(response => response.data);
    }

    getLLMProvider(providerId: string, cancelToken?: CancelToken): Promise<LLMProviderConfig> {
        return this.axiosInstance.get<LLMProviderConfig>(`/api/v1/llm-provider/${providerId}`, { cancelToken })
            .then(response => response.data);
    }

    updateLLMProvider(providerId: string, providerData: Partial<LLMProviderConfig>, cancelToken?: CancelToken): Promise<LLMProviderConfig> {
        return this.axiosInstance.put<LLMProviderConfig>(`/api/v1/llm-provider/${providerId}`, providerData, { cancelToken })
            .then(response => response.data);
    }

    createLLMProvider(providerData: Omit<LLMProviderConfig, 'uuid'>, cancelToken?: CancelToken): Promise<LLMProviderConfig> {
        return this.axiosInstance.post<LLMProviderConfig>('/api/v1/llm-provider', providerData, { cancelToken })
            .then(response => response.data);
    }

    deleteLLMProvider(providerId: string, cancelToken?: CancelToken): Promise<void> {
        return this.axiosInstance.delete(`/api/v1/llm-provider/${providerId}`, { cancelToken });
    }

    activateLLMProvider(providerId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/llm-provider/${providerId}/activate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }

    deactivateLLMProvider(providerId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/llm-provider/${providerId}/deactivate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }
}