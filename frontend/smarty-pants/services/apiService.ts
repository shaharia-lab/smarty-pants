import { AxiosInstance, AxiosError, CancelToken, default as axios } from 'axios';
import { IAuthService } from './authService';
import { AnalyticsApi } from './api/analytics';
import { DocumentApi } from "@/services/api/document";
import { ChatHistoriesApi } from "@/services/api/interactions";
import { DatasourcesApi } from "@/services/api/datasource";
import { EmbeddingProviderApi } from "@/services/api/embedding_provider";
import { LLMProviderApi } from "@/services/api/llm_provider";
import { UsersApi } from "@/services/api/users";
import { SettingsApi } from "@/services/api/settings";
import { SystemAPI } from "@/services/api/system";

export class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

export class ApiService {
    private authenticatedAxiosInstance: AxiosInstance;
    private unauthenticatedAxiosInstance: AxiosInstance;
    public analytics: AnalyticsApi;
    public documents: DocumentApi;
    public chatHisories: ChatHistoriesApi;
    public datasource: DatasourcesApi;
    public embeddingProvider: EmbeddingProviderApi
    public llmProvider: LLMProviderApi;
    public usersApi: UsersApi;
    public settingsApi: SettingsApi
    public systemApi: SystemAPI;

    constructor(private authService: IAuthService) {
        let backendUrl = process.env.NEXT_PUBLIC_API_BASE_URL || ''; // Fallback to empty string or default URL

        if (typeof window !== 'undefined') {
            backendUrl = localStorage.getItem('backendUrl') || backendUrl;
        }

        this.authenticatedAxiosInstance = this.authService.getAuthenticatedAxiosInstance();
        this.authenticatedAxiosInstance.defaults.baseURL = backendUrl;

        this.unauthenticatedAxiosInstance = axios.create({
            baseURL: backendUrl,
        });

        this.analytics = new AnalyticsApi(this.authenticatedAxiosInstance);
        this.documents = new DocumentApi(this.authenticatedAxiosInstance);
        this.chatHisories = new ChatHistoriesApi(this.authenticatedAxiosInstance);
        this.datasource = new DatasourcesApi(this.authenticatedAxiosInstance);
        this.embeddingProvider = new EmbeddingProviderApi(this.authenticatedAxiosInstance)
        this.llmProvider = new LLMProviderApi(this.authenticatedAxiosInstance);
        this.usersApi = new UsersApi(this.authenticatedAxiosInstance);
        this.settingsApi = new SettingsApi(this.authenticatedAxiosInstance);
        this.systemApi = new SystemAPI(this.unauthenticatedAxiosInstance);
    }
}

export const createApiService = (authService: IAuthService) => new ApiService(authService);