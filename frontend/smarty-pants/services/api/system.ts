import { AxiosInstance, CancelToken } from 'axios';
import {GenerateResponseMsg, SystemInfo} from '@/types/api';

export class SystemAPI {
    constructor(private axiosInstance: AxiosInstance) {}

    async getSystemInfo(cancelToken?: CancelToken): Promise<SystemInfo> {
        const response = await this.axiosInstance.get<SystemInfo>('/system/info', { cancelToken });
        return response.data;
    }

    async ping(cancelToken?: CancelToken): Promise<GenerateResponseMsg> {
        const response = await this.axiosInstance.get<GenerateResponseMsg>('/system/ping', { cancelToken });
        return response.data;
    }

    async checkLiveness(cancelToken?: CancelToken): Promise<GenerateResponseMsg> {
        const response = await this.axiosInstance.get<GenerateResponseMsg>('/system/probes/liveness', { cancelToken });
        return response.data;
    }

    async checkReadiness(cancelToken?: CancelToken): Promise<GenerateResponseMsg> {
        const response = await this.axiosInstance.get<GenerateResponseMsg>('/system/probes/readiness', { cancelToken });
        return response.data;
    }
}