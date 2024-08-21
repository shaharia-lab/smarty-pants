// File: types/user.ts
export type UserRole = 'user' | 'developer' | 'admin';

export type UserStatus = 'active' | 'inactive';

export interface User {
    uuid: string;
    name: string;
    email: string;
    status: UserStatus;
    roles: UserRole[];
    created_at: string;
    updated_at: string;
}

export interface PaginatedUsers {
    users: User[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}