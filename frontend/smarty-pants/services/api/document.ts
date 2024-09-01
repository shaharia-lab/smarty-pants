import { AxiosInstance, CancelToken } from 'axios';
import { Document } from '@/types/document';

export interface DocumentResponse {
    documents: Document[];
    total_pages: number;
    total: number;
}

export class DocumentApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getDocuments(
        page: number = 1,
        status: string = '',
        limit: number = 10,
        cancelToken?: CancelToken
    ): Promise<DocumentResponse> {
        const response = await this.axiosInstance.get<DocumentResponse>(
            '/api/v1/document',
            {
                params: { page, status, limit },
                cancelToken
            }
        );
        return response.data;
    }
}