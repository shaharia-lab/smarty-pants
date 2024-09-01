'use client';

import { useParams } from 'next/navigation';
import React, { useEffect, useState, useMemo } from 'react';
import Navbar from '../../../components/Navbar';
import DocumentDetailsClient from '../../../components/DocumentDetailsClient';
import { Document } from '@/types';
import { createApiService } from "@/services/apiService";
import AuthService from "@/services/authService";
import axios, { CancelTokenSource } from "axios";

export default function DocumentDetailsPage() {
    const params = useParams() as { uuid: string };
    const uuid = params.uuid as string;
    const [document, setDocument] = useState<Document | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const apiService = useMemo(() => createApiService(AuthService), []);

    useEffect(() => {
        if (!uuid) {
            setError('No document UUID provided');
            setLoading(false);
            return;
        }

        const cancelTokenSource: CancelTokenSource = axios.CancelToken.source();

        async function fetchDocument() {
            try {
                const data = await apiService.documents.getDocumentByUuid(uuid, cancelTokenSource.token);
                setDocument(data);
            } catch (err) {
                if (!axios.isCancel(err)) {
                    setError('Error fetching document details');
                    console.error('Error fetching document:', err);
                }
            } finally {
                setLoading(false);
            }
        }

        fetchDocument();

        return () => {
            cancelTokenSource.cancel('Operation canceled due to component unmount or re-render.');
        };
    }, [uuid, apiService]);

    return (
        <div className="min-h-screen bg-gray-100">
            <Navbar/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    <h1 className="text-3xl font-bold text-gray-900 mb-6">Document Details</h1>
                    {loading && <div className="text-center mt-8">Loading...</div>}
                    {error && <div className="text-center mt-8 text-red-500">{error}</div>}
                    {!loading && !error && document && <DocumentDetailsClient document={document}/>}
                </div>
            </main>
        </div>
    );
}