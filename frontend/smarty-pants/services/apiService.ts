import { AxiosInstance, AxiosError, CancelToken } from 'axios';
import { IAuthService } from './authService';
import { AnalyticsApi } from './api/analytics';
import {DocumentApi} from "@/services/api/document";
import {ChatHistoriesApi} from "@/services/api/interactions";
import {DatasourcesApi} from "@/services/api/datasource";
import {EmbeddingProviderApi} from "@/services/api/embedding_provider";
import {LLMProviderApi} from "@/services/api/llm_provider";
import {UsersApi} from "@/services/api/users";

export class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

export class ApiService {
    private axiosInstance: AxiosInstance;
    public analytics: AnalyticsApi;
    public documents: DocumentApi;
    public chatHisories: ChatHistoriesApi;
    public datasource: DatasourcesApi;
    public embeddingProvider: EmbeddingProviderApi
    public llmProvider: LLMProviderApi;
    public usersApi: UsersApi;

    constructor(private authService: IAuthService) {
        this.axiosInstance = this.authService.getAuthenticatedAxiosInstance();
        this.analytics = new AnalyticsApi(this.axiosInstance);
        this.documents = new DocumentApi(this.axiosInstance);
        this.chatHisories = new ChatHistoriesApi(this.axiosInstance);
        this.datasource = new DatasourcesApi(this.axiosInstance);
        this.embeddingProvider = new EmbeddingProviderApi(this.axiosInstance)
        this.llmProvider = new LLMProviderApi(this.axiosInstance);
        this.usersApi = new UsersApi(this.axiosInstance);
    }

    // Make this method protected so it can be used by subclasses if needed
    protected async request<T>(method: string, url: string, data?: any, cancelToken?: CancelToken): Promise<T> {
        try {
            const response = await this.axiosInstance.request<T>({
                method,
                url,
                data,
                cancelToken,
            });
            return response.data;
        } catch (error) {
            if (error instanceof AxiosError) {
                throw new ApiError(error.response?.status || 500, error.message);
            }
            throw error;
        }
    }
}

export const createApiService = (authService: IAuthService) => new ApiService(authService);