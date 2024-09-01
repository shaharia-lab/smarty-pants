'use client';

import React, { useEffect, useState, useCallback, useMemo } from 'react';
import Filter from './Filter';
import DocumentTable from './DocumentTable';
import Pagination from './Pagination';
import { Document } from '@/types/document';
import { createApiService } from "@/services/apiService";
import AuthService from "@/services/authService";
import axios, { CancelTokenSource } from "axios";

const DocumentClient: React.FC = () => {
    const apiService = useMemo(() => createApiService(AuthService), []);
    const [documents, setDocuments] = useState<Document[]>([]);
    const [currentPage, setCurrentPage] = useState<number>(1);
    const [totalPages, setTotalPages] = useState<number>(1);
    const [totalDocuments, setTotalDocuments] = useState<number>(0);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [status, setStatus] = useState<string>('');
    const [limit, setLimit] = useState<number>(10);

    const fetchDocuments = useCallback(async (page: number, cancelToken: CancelTokenSource) => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await apiService.documents.getDocuments(page, status, limit, cancelToken.token);
            setDocuments(response.documents);
            setTotalPages(response.total_pages);
            setTotalDocuments(response.total);
        } catch (error) {
            if (!axios.isCancel(error)) {
                setError('Error fetching documents. Please try again.');
                console.error('Error fetching documents:', error);
            }
        } finally {
            setIsLoading(false);
        }
    }, [apiService, status, limit]);

    useEffect(() => {
        const cancelToken = axios.CancelToken.source();
        fetchDocuments(currentPage, cancelToken);

        return () => {
            cancelToken.cancel('Operation canceled due to new request.');
        };
    }, [fetchDocuments, currentPage]);

    const handleFilterApply = useCallback((newStatus: string, newLimit: number) => {
        setStatus(newStatus);
        setLimit(newLimit);
        setCurrentPage(1);
    }, []);

    const handlePageChange = useCallback((newPage: number) => {
        setCurrentPage(newPage);
    }, []);

    return (
        <div className="p-6 bg-white">
            <Filter onFilterApply={handleFilterApply} />
            {isLoading ? (
                <p>Loading documents...</p>
            ) : error ? (
                <p className="text-red-500">{error}</p>
            ) : (
                <>
                    <DocumentTable documents={documents} />
                    <Pagination
                        currentPage={currentPage}
                        totalPages={totalPages}
                        onPageChange={handlePageChange}
                    />
                    <p className="mt-4 text-sm text-gray-500">
                        Total documents: {totalDocuments}
                    </p>
                </>
            )}
        </div>
    );
};

export default DocumentClient;