// File: src/services/DatasourcesApi.ts

import axios, { AxiosInstance, CancelToken } from 'axios';
import {DatasourceConfig, SlackDatasourcePayload} from '@/types/datasource';

interface DatasourcesApiResponse {
    datasources: DatasourceConfig[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}

export class DatasourcesApi {
    constructor(private axiosInstance: AxiosInstance) {}

    getDatasources(cancelToken?: CancelToken): Promise<DatasourcesApiResponse> {
        return this.axiosInstance.get<DatasourcesApiResponse>('/api/v1/datasource', { cancelToken })
            .then(response => response.data);
    }

    deleteDatasource(datasourceId: string, cancelToken?: CancelToken): Promise<void> {
        return this.axiosInstance.delete(`/api/v1/datasource/${datasourceId}`, { cancelToken });
    }

    activateDatasource(datasourceId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/datasource/${datasourceId}/activate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }

    deactivateDatasource(datasourceId: string, cancelToken?: CancelToken): Promise<{ message: string }> {
        return this.axiosInstance.put<{ message: string }>(
            `/api/v1/datasource/${datasourceId}/deactivate`,
            {},
            { cancelToken }
        ).then(response => response.data);
    }
    addSlackDatasource(payload: SlackDatasourcePayload, cancelToken?: CancelToken): Promise<DatasourceConfig> {
        return this.axiosInstance.post<DatasourceConfig>('/api/v1/datasource', payload, { cancelToken })
            .then(response => response.data);
    }
}