import { AxiosInstance, CancelToken } from 'axios';
import { PaginatedUsers, User, UserRole, UserStatus } from '@/types/user';

export class UsersApi {
    constructor(private axiosInstance: AxiosInstance) {}

    async getUsers(page: number, perPage: number, cancelToken?: CancelToken): Promise<PaginatedUsers> {
        const response = await this.axiosInstance.get<PaginatedUsers>('/api/v1/users', {
            params: { page, per_page: perPage },
            cancelToken
        });
        return response.data;
    }

    async updateUserStatus(uuid: string, status: UserStatus, cancelToken?: CancelToken): Promise<User> {
        const action = status === 'active' ? 'activate' : 'deactivate';
        const response = await this.axiosInstance.put<User>(`/api/v1/users/${uuid}/${action}`, {}, { cancelToken });
        return response.data;
    }

    async updateUserRoles(uuid: string, roles: UserRole[], cancelToken?: CancelToken): Promise<User> {
        const response = await this.axiosInstance.put<User>(`/api/v1/users/${uuid}`, { roles }, { cancelToken });
        return response.data;
    }
}