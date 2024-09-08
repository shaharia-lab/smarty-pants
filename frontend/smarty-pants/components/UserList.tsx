import React, { useState, useEffect, useCallback } from 'react';
import { User, UserRole, UserStatus } from '@/types/user';
import Pagination from '@/components/Pagination';
import {createApiService} from '@/services/apiService';
import AuthService from "@/services/authService";

interface UserListProps {
    users: User[];
    currentPage: number;
    totalPages: number;
    onPageChange: (page: number) => void;
}

const UserList: React.FC<UserListProps> = ({ users: initialUsers, currentPage, totalPages, onPageChange }) => {
    const [expandedUser, setExpandedUser] = useState<string | null>(null);
    const [filter, setFilter] = useState({ name: '', email: '', status: '', role: '' });
    const [localUsers, setLocalUsers] = useState<User[]>(initialUsers);

    const usersApi = createApiService(AuthService).usersApi;

    useEffect(() => {
        setLocalUsers(initialUsers);
    }, [initialUsers]);

    const toggleUserExpansion = (uuid: string) => {
        setExpandedUser(expandedUser === uuid ? null : uuid);
    };

    const handleFilterChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        setFilter({ ...filter, [e.target.name]: e.target.value });
    };

    const filteredUsers = localUsers.filter(user => {
        return (
            user.name.toLowerCase().includes(filter.name.toLowerCase()) &&
            user.email.toLowerCase().includes(filter.email.toLowerCase()) &&
            (filter.status === '' || user.status === filter.status) &&
            (filter.role === '' || user.roles.includes(filter.role as UserRole))
        );
    });

    const updateUserStatus = useCallback(async (uuid: string, newStatus: UserStatus) => {
        try {
            const updatedUser = await usersApi.updateUserStatus(uuid, newStatus);
            setLocalUsers(prevUsers =>
                prevUsers.map(user =>
                    user.uuid === uuid ? updatedUser : user
                )
            );
        } catch (error) {
            console.error('Error updating user status:', error);
        }
    }, [usersApi]);

    const updateUserRoles = useCallback(async (uuid: string, roles: UserRole[]) => {
        try {
            const updatedUser = await usersApi.updateUserRoles(uuid, roles);
            setLocalUsers(prevUsers =>
                prevUsers.map(user =>
                    user.uuid === uuid ? updatedUser : user
                )
            );
        } catch (error) {
            console.error('Error updating user roles:', error);
        }
    }, [usersApi]);

    return (
        <div>
            {/* Filter inputs */}
            <div className="mb-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <input
                    type="text"
                    name="name"
                    placeholder="Filter by name"
                    value={filter.name}
                    onChange={handleFilterChange}
                    className="border p-2 rounded"
                />
                <input
                    type="text"
                    name="email"
                    placeholder="Filter by email"
                    value={filter.email}
                    onChange={handleFilterChange}
                    className="border p-2 rounded"
                />
                <select
                    data-testid="status-filter"
                    name="status"
                    value={filter.status}
                    onChange={handleFilterChange}
                    className="border p-2 rounded"
                >
                    <option value="">All Statuses</option>
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                </select>
                <select
                    data-testid="role-filter"
                    name="role"
                    value={filter.role}
                    onChange={handleFilterChange}
                    className="border p-2 rounded"
                >
                    <option value="">All Roles</option>
                    <option value="user">User</option>
                    <option value="developer">Developer</option>
                    <option value="admin">Admin</option>
                </select>
            </div>

            {/* User list */}
            <div className="bg-white shadow overflow-hidden sm:rounded-md">
                <ul className="divide-y divide-gray-200">
                    {filteredUsers.map((user) => (
                        <li key={user.uuid} className="hover:bg-gray-50">
                            <div className="px-4 py-4 sm:px-6">
                                <div className="flex items-center justify-between">
                                    <p className="text-sm font-medium text-indigo-600 truncate">{user.name}</p>
                                    <div className="ml-2 flex-shrink-0 flex">
                                        <p className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                                            user.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                                        }`}>
                                            {user.status}
                                        </p>
                                    </div>
                                </div>
                                <div className="mt-2 sm:flex sm:justify-between">
                                    <div className="sm:flex">
                                        <p className="flex items-center text-sm text-gray-500">
                                            {user.email}
                                        </p>
                                    </div>
                                    <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                                        <p>{user.roles.join(', ')}</p>
                                    </div>
                                </div>
                                <button
                                    onClick={() => toggleUserExpansion(user.uuid)}
                                    className="mt-2 text-sm text-indigo-600 hover:text-indigo-900 focus:outline-none focus:underline"
                                >
                                    {expandedUser === user.uuid ? 'Hide Details' : 'Show Details'}
                                </button>
                            </div>
                            {expandedUser === user.uuid && (
                                <div className="px-4 py-5 sm:px-6 bg-gray-50 border-t border-gray-200">
                                    <div className="space-y-6">
                                        <div>
                                            <h4 className="text-lg leading-6 font-medium text-gray-900">User Details</h4>
                                            <dl className="mt-2 grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
                                                <div className="sm:col-span-1">
                                                    <dt className="text-sm font-medium text-gray-500">Created At</dt>
                                                    <dd className="mt-1 text-sm text-gray-900">{new Date(user.created_at).toLocaleString()}</dd>
                                                </div>
                                                <div className="sm:col-span-1">
                                                    <dt className="text-sm font-medium text-gray-500">Updated At</dt>
                                                    <dd className="mt-1 text-sm text-gray-900">{new Date(user.updated_at).toLocaleString()}</dd>
                                                </div>
                                            </dl>
                                        </div>
                                        <div>
                                            <h4 className="text-lg leading-6 font-medium text-gray-900">Manage Status</h4>
                                            <button
                                                onClick={() => updateUserStatus(user.uuid, user.status === 'active' ? 'inactive' : 'active')}
                                                className={`mt-2 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white ${
                                                    user.status === 'active'
                                                        ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500'
                                                        : 'bg-green-600 hover:bg-green-700 focus:ring-green-500'
                                                } focus:outline-none focus:ring-2 focus:ring-offset-2`}
                                            >
                                                {user.status === 'active' ? 'Deactivate' : 'Activate'}
                                            </button>
                                        </div>
                                        <div>
                                            <h4 className="text-lg leading-6 font-medium text-gray-900">Manage Roles</h4>
                                            <div className="mt-2 space-y-2 sm:space-y-0 sm:space-x-4">
                                                {(['user', 'developer', 'admin'] as UserRole[]).map((role) => (
                                                    <label key={role} className="inline-flex items-center">
                                                        <input
                                                            type="checkbox"
                                                            checked={user.roles.includes(role)}
                                                            onChange={(e) => {
                                                                const newRoles = e.target.checked
                                                                    ? [...user.roles, role]
                                                                    : user.roles.filter((r) => r !== role);
                                                                updateUserRoles(user.uuid, newRoles);
                                                            }}
                                                            className="form-checkbox h-5 w-5 text-indigo-600 transition duration-150 ease-in-out"
                                                        />
                                                        <span className="ml-2 text-sm text-gray-700">{role}</span>
                                                    </label>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </li>
                    ))}
                </ul>
            </div>
            <div className="mt-4">
                <Pagination
                    currentPage={currentPage}
                    totalPages={totalPages}
                    onPageChange={onPageChange}
                />
            </div>
        </div>
    );
};

export default UserList;