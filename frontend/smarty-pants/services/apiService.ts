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
import { SystemApi } from "@/services/api/SystemApi";
import {getRuntimeConfig} from "@/config";

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
    public systemApi: SystemApi;

    constructor(private authService: IAuthService) {
        const { API_BASE_URL } = getRuntimeConfig();
        this.authenticatedAxiosInstance = this.authService.getAuthenticatedAxiosInstance();
        this.unauthenticatedAxiosInstance = axios.create({
            baseURL: API_BASE_URL,
        });

        this.analytics = new AnalyticsApi(this.authenticatedAxiosInstance);
        this.documents = new DocumentApi(this.authenticatedAxiosInstance);
        this.chatHisories = new ChatHistoriesApi(this.authenticatedAxiosInstance);
        this.datasource = new DatasourcesApi(this.authenticatedAxiosInstance);
        this.embeddingProvider = new EmbeddingProviderApi(this.authenticatedAxiosInstance)
        this.llmProvider = new LLMProviderApi(this.authenticatedAxiosInstance);
        this.usersApi = new UsersApi(this.authenticatedAxiosInstance);
        this.settingsApi = new SettingsApi(this.authenticatedAxiosInstance);
        this.systemApi = new SystemApi(this.unauthenticatedAxiosInstance);
    }
}

export const createApiService = (authService: IAuthService) => new ApiService(authService);