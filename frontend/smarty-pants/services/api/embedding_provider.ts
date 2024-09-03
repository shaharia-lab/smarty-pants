import { AxiosInstance, CancelToken } from 'axios';
import { EmbeddingProviderConfig, EmbeddingProvidersApiResponse } from '@/types/embeddingProvider';

export class EmbeddingProviderApi {
    constructor(private axiosInstance: AxiosInstance) {}

    getEmbeddingProviders(cancelToken?: CancelToken): Promise<EmbeddingProvidersApiResponse> {
        return this.axiosInstance.get<EmbeddingProvidersApiResponse>('/api/v1/embedding-provider', { cancelToken })
            .then(response => response.data);
    }

    deleteEmbeddingProvider(providerId: string, cancelToken?: CancelToken): Promise<void> {
        return this.axiosInstance.delete(`/api/v1/embedding-provider/${providerId}`, { cancelToken });
    }

    activateEmbeddingProvider(providerId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/embedding-provider/${providerId}/activate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }

    deactivateEmbeddingProvider(providerId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/embedding-provider/${providerId}/deactivate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }
}