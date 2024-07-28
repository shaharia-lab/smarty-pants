// File: /components/EmbeddingProviders.tsx

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Navbar from './Navbar';
import Header from './Header';
import { EmbeddingProviderConfig } from '@/types/embeddingProvider';

const EmbeddingProviders: React.FC = () => {
    const [embeddingProviders, setEmbeddingProviders] = useState<EmbeddingProviderConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();

    useEffect(() => {
        const fetchEmbeddingProviders = async () => {
            try {
                const response = await fetch('http://localhost:8080/api/embedding-providers');
                if (!response.ok) {
                    throw new Error('Failed to fetch embedding providers');
                }
                const data = await response.json();
                setEmbeddingProviders(data.embedding_providers);
            } catch (err) {
                setError('Failed to load embedding providers. Please try again later.');
            } finally {
                setLoading(false);
            }
        };

        fetchEmbeddingProviders();
    }, []);

    const handleAddProvider = () => {
        router.push('/embedding-providers/add');
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar />
            <Header config={{ title: "Embedding Providers" }} />
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    {loading ? (
                        <p>Loading embedding providers...</p>
                    ) : error ? (
                        <p className="text-red-500">{error}</p>
                    ) : (
                        <>
                            <div className="mb-4">
                                <button
                                    onClick={handleAddProvider}
                                    className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
                                >
                                    Add Embedding Provider
                                </button>
                            </div>
                            {embeddingProviders.length > 0 ? (
                                <ul className="divide-y divide-gray-200">
                                    {embeddingProviders.map((provider) => (
                                        <li key={provider.uuid} className="py-4">
                                            <div className="flex items-center space-x-4">
                                                <div className="flex-1 min-w-0">
                                                    <p className="text-sm font-medium text-gray-900 truncate">
                                                        {provider.name}
                                                    </p>
                                                    <p className="text-sm text-gray-500 truncate">
                                                        {provider.provider}
                                                    </p>
                                                </div>
                                                <div className="inline-flex items-center text-base font-semibold text-gray-900">
                                                    {provider.status}
                                                </div>
                                            </div>
                                        </li>
                                    ))}
                                </ul>
                            ) : (
                                <p>No embedding providers configured yet.</p>
                            )}
                        </>
                    )}
                </div>
            </main>
        </div>
    );
};

export default EmbeddingProviders;