// File: app/users/page.tsx
'use client';

import React, { useState, useEffect } from 'react';
import Navbar from '@/components/Navbar';
import Header from '@/components/Header';
import UserList from '../../components/UserList';
import { User, PaginatedUsers } from '@/types/user';

const UsersPage: React.FC = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetchUsers(currentPage);
    }, [currentPage]);

    const fetchUsers = async (page: number) => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/users?page=${page}&per_page=10`);
            if (!response.ok) {
                throw new Error('Failed to fetch users');
            }
            const data: PaginatedUsers = await response.json();
            setUsers(data.users);
            setTotalPages(data.total_pages);
        } catch (err) {
            setError('Error fetching users. Please try again.');
            console.error('Error fetching users:', err);
        } finally {
            setIsLoading(false);
        }
    };

    const handlePageChange = (newPage: number) => {
        setCurrentPage(newPage);
    };

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