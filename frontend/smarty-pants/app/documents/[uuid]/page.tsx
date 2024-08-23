'use client';

import {useParams} from 'next/navigation';
import {useEffect, useState} from 'react';
import Navbar from '../../../components/Navbar';
import DocumentDetailsClient from '../../../components/DocumentDetailsClient';
import {Document} from '@/types';

export default function DocumentDetailsPage() {
    const params = useParams() as { uuid: string };
    const uuid = params.uuid as string;
    const [document, setDocument] = useState<Document | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!uuid) {
            setError('No document UUID provided');
            setLoading(false);
            return;
        }

        async function fetchDocument() {
            try {
                const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/document/${uuid}`);
                if (!response.ok) {
                    throw new Error('Failed to fetch document');
                }
                const data: Document = await response.json();
                setDocument(data);
            } catch (err) {
                setError('Error fetching document details');
            } finally {
                setLoading(false);
            }
        }

        fetchDocument();
    }, [uuid]);

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