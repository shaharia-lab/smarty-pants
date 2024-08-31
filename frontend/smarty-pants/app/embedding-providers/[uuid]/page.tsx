'use client';

import React, {useEffect, useState} from 'react';
import {useParams, useRouter} from 'next/navigation';
import Navbar from '../../../components/Navbar';
import Header, {HeaderConfig} from '../../../components/Header';
import {EmbeddingProviderConfig} from '@/types/embeddingProvider';

const EmbeddingProviderDetailsPage: React.FC = () => {
    const [provider, setProvider] = useState<EmbeddingProviderConfig | null>(null);
    const [apiKey, setApiKey] = useState('');
    const [modelId, setModelId] = useState('');
    const [isEditing, setIsEditing] = useState(false);
    const [error, setError] = useState<string | null>(null);
    useRouter();
    const params = useParams() as { uuid: string };
    const uuid = params.uuid as string;

    const headerConfig: HeaderConfig = {
        title: provider ? `${provider.name} Details` : "Embedding Provider Details"
    };

    useEffect(() => {
        const fetchProviderDetails = async () => {
            try {
                const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider/${uuid}`);
                if (!response.ok) {
                    throw new Error('Failed to fetch embedding provider details');
                }
                const data: EmbeddingProviderConfig = await response.json();
                setProvider(data);
                setApiKey(data.configuration.api_key);
                setModelId(data.configuration.model_id);
            } catch (err) {
                setError('Failed to load embedding provider details. Please try again later.');
            }
        };

        fetchProviderDetails();
    }, [uuid]);

    const handleUpdate = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider/${uuid}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: provider?.name,
                    provider: provider?.provider,
                    configuration: {
                        api_key: apiKey,
                        model_id: modelId,
                    },
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to update embedding provider');
            }

            setIsEditing(false);
            // Refresh provider details
            const updatedProvider = await response.json();
            setProvider(updatedProvider);
        } catch (err) {
            setError('Failed to update embedding provider. Please try again.');
        }
    };

    if (!provider) {
        return <div>Loading...</div>;
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    <div className="bg-white shadow overflow-hidden sm:rounded-lg">
                        <div className="px-4 py-5 sm:px-6">
                            <h3 className="text-lg leading-6 font-medium text-gray-900">Embedding Provider Details</h3>
                            <p className="mt-1 max-w-2xl text-sm text-gray-500">Details and settings for this embedding
                                provider.</p>
                        </div>
                        <div className="border-t border-gray-200 px-4 py-5 sm:p-0">
                            <dl className="sm:divide-y sm:divide-gray-200">
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500">Name</dt>
                                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{provider.name}</dd>
                                </div>
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500">Provider</dt>
                                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{provider.provider}</dd>
                                </div>
                                <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                    <dt className="text-sm font-medium text-gray-500">Status</dt>
                                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{provider.status}</dd>
                                </div>
                                <form onSubmit={handleUpdate}>
                                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                        <dt className="text-sm font-medium text-gray-500">API Key</dt>
                                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                                            {isEditing ? (
                                                <input
                                                    type="password"
                                                    value={apiKey}
                                                    onChange={(e) => setApiKey(e.target.value)}
                                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                />
                                            ) : (
                                                '********'
                                            )}
                                        </dd>
                                    </div>
                                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                        <dt className="text-sm font-medium text-gray-500">Model ID</dt>
                                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                                            {isEditing ? (
                                                <input
                                                    type="text"
                                                    value={modelId}
                                                    onChange={(e) => setModelId(e.target.value)}
                                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                />
                                            ) : (
                                                modelId
                                            )}
                                        </dd>
                                    </div>
                                    {error && (
                                        <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                                            <dt className="text-sm font-medium text-red-500">Error</dt>
                                            <dd className="mt-1 text-sm text-red-500 sm:mt-0 sm:col-span-2">{error}</dd>
                                        </div>
                                    )}
                                    <div className="py-4 sm:py-5 sm:px-6">
                                        {isEditing ? (
                                            <div className="flex justify-end space-x-3">
                                                <button
                                                    type="button"
                                                    onClick={() => setIsEditing(false)}
                                                    className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                                >
                                                    Cancel
                                                </button>
                                                <button
                                                    type="submit"
                                                    className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                                >
                                                    Save Changes
                                                </button>
                                            </div>
                                        ) : (
                                            <button
                                                type="button"
                                                onClick={() => setIsEditing(true)}
                                                className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                            >
                                                Edit Settings
                                            </button>
                                        )}
                                    </div>
                                </form>
                            </dl>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

export default EmbeddingProviderDetailsPage;