import { AxiosInstance, AxiosError, CancelToken } from 'axios';

export interface IAuthService {
    getAuthenticatedAxiosInstance(): AxiosInstance;
}

export interface AnalyticsOverview {
    embedding_providers: {
        total_providers: number;
        total_active_providers: number;
        active_provider: {
            name: string;
            type: string;
            model: string;
        };
    };
    llm_providers: {
        total_providers: number;
        total_active_providers: number;
        active_provider: {
            name: string;
            type: string;
            model: string;
        };
    };
    datasources: {
        configured_datasources: Array<{
            name: string;
            type: string;
            status: string;
            created_at: string;
        }> | null;
        total_datasources: number;
        total_datasources_by_type: { [key: string]: number };
        total_datasources_by_status: { [key: string]: number };
        total_documents_fetched_by_datasource_type: { [key: string]: number };
    };
}

export class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

export class ApiService {
    private axiosInstance: AxiosInstance;

    constructor(private authService: IAuthService) {
        this.axiosInstance = this.authService.getAuthenticatedAxiosInstance();
    }

    private async request<T>(method: string, url: string, data?: any, cancelToken?: CancelToken): Promise<T> {
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

    public async getAnalyticsOverview(cancelToken?: CancelToken): Promise<AnalyticsOverview> {
        return this.request<AnalyticsOverview>('GET', '/api/v1/analytics/overview', undefined, cancelToken);
    }

    // Add more API methods here
}

export const createApiService = (authService: IAuthService) => new ApiService(authService);