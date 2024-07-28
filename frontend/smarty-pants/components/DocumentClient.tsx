'use client';

import React, { useState, useEffect } from 'react';
import Filter from './Filter';
import DocumentTable from './DocumentTable';
import Pagination from './Pagination';
import { Document } from '@/types/document';

const DocumentClient: React.FC = () => {
    const [documents, setDocuments] = useState<Document[]>([]);
    const [currentPage, setCurrentPage] = useState<number>(1);
    const [totalPages, setTotalPages] = useState<number>(1);
    const [totalDocuments, setTotalDocuments] = useState<number>(0);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const fetchDocuments = async (page: number = 1, status: string = '', limit: number = 10) => {
        setIsLoading(true);
        setError(null);
        try {
            const response = await fetch(`http://localhost:8080/api/documents?page=${page}&status=${status}&limit=${limit}`);
            if (!response.ok) {
                throw new Error('Failed to fetch documents');
            }
            const data = await response.json();
            setDocuments(data.documents);
            setTotalPages(data.total_pages);
            setTotalDocuments(data.total);
        } catch (error) {
            setError('Error fetching documents. Please try again.');
            console.error('Error fetching documents:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchDocuments();
    }, []);

    const handleFilterApply = (status: string, limit: number) => {
        setCurrentPage(1);
        fetchDocuments(1, status, limit);
    };

    const handlePageChange = (newPage: number) => {
        setCurrentPage(newPage);
        fetchDocuments(newPage);
    };

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