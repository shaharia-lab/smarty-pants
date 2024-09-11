'use client';

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import axios, { CancelTokenSource } from 'axios';
import Navbar from '@/components/Navbar';
import Header from '@/components/Header';
import UserList from '@/components/UserList';
import { User } from '@/types/user';
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";

const UsersPage: React.FC = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const usersApi = useMemo(() => createApiService(AuthService).usersApi, []);
    const cancelTokenSourceRef = useRef<CancelTokenSource | null>(null);

    const fetchUsers = useCallback(async (page: number) => {
        if (cancelTokenSourceRef.current) {
            cancelTokenSourceRef.current.cancel('Operation canceled due to new request.');
        }
        cancelTokenSourceRef.current = axios.CancelToken.source();

        setIsLoading(true);
        setError(null);
        try {
            const data = await usersApi.getUsers(page, 10, cancelTokenSourceRef.current.token);
            console.log('Fetched users:', data.users); // Debug log
            setUsers(data.users);
            setTotalPages(data.total_pages);
        } catch (err) {
            if (!axios.isCancel(err)) {
                setError('Error fetching users. Please try again.');
                console.error('Error fetching users:', err);
            }
        } finally {
            setIsLoading(false);
        }
    }, [usersApi]);

    useEffect(() => {
        fetchUsers(currentPage);

        return () => {
            if (cancelTokenSourceRef.current) {
                cancelTokenSourceRef.current.cancel('Component unmounted');
            }
        };
    }, [currentPage, fetchUsers]);

    const handlePageChange = (newPage: number) => {
        setCurrentPage(newPage);
    };

    console.log('Rendering UsersPage, users:', users); // Debug log

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar />
            <Header config={{ title: "User Management" }} />
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    {isLoading ? (
                        <p>Loading users...</p>
                    ) : error ? (
                        <p className="text-red-500">{error}</p>
                    ) : (
                        <UserList
                            users={users}
                            currentPage={currentPage}
                            totalPages={totalPages}
                            onPageChange={handlePageChange}
                        />
                    )}
                </div>
            </main>
        </div>
    );
};

export default UsersPage;