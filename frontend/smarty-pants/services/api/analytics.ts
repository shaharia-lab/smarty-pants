import { AxiosInstance, CancelToken } from 'axios';
import { AnalyticsOverview } from '@/types/api';

export class AnalyticsApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getAnalyticsOverview(cancelToken?: CancelToken): Promise<AnalyticsOverview> {
        const response = await this.axiosInstance.get<AnalyticsOverview>('/api/v1/analytics/overview', { cancelToken });
        return response.data;
    }
}