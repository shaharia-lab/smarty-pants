'use client';

import React, {useEffect, useState} from 'react';
import Navbar from '../../components/Navbar';
import Header, {HeaderConfig} from '../../components/Header';
import {EmbeddingProviderConfig} from '@/types/embeddingProvider';
import {availableEmbeddingProviders} from '@/utils/embeddingProviders';
import {Alert, AlertDescription} from '@/components/Alert';
import ListComponent from '../../components/ListComponent';
import AvailableProviders from '../../components/AvailableProviders';

interface EmbeddingProvidersApiResponse {
    embedding_providers: EmbeddingProviderConfig[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const EmbeddingProvidersPage: React.FC = () => {
    const [configuredProviders, setConfiguredProviders] = useState<EmbeddingProviderConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const headerConfig: HeaderConfig = {
        title: "Embedding Providers"
    };

    useEffect(() => {
        fetchEmbeddingProviders();
    }, []);

    const fetchEmbeddingProviders = async () => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-providers`);
            if (!response.ok) {
                throw new Error('Failed to fetch embedding providers');
            }
            const data: EmbeddingProvidersApiResponse = await response.json();
            setConfiguredProviders(data.embedding_providers);
        } catch (err) {
            setError('Failed to load embedding providers. Please try again later.');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (providerId: string) => {
        if (window.confirm('Are you sure you want to delete this provider?')) {
            try {
                const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider/${providerId}`, {
                    method: 'DELETE',
                });
                if (!response.ok) {
                    throw new Error('Failed to delete embedding provider');
                }
                setFlashMessage({type: 'success', message: 'Embedding provider deleted successfully'});
                fetchEmbeddingProviders();
            } catch (err) {
                setFlashMessage({type: 'error', message: 'Failed to delete embedding provider. Please try again.'});
            }
        }
    };

    const handleActivate = async (providerId: string) => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider/${providerId}/activate`, {
                method: 'PUT',
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to activate embedding provider');
            }
            setFlashMessage({type: 'success', message: data.message});
            fetchEmbeddingProviders();
        } catch (err) {
            setFlashMessage({type: 'error', message: err instanceof Error ? err.message : 'An error occurred'});
        }
    };

    const handleDeactivate = async (providerId: string) => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider/${providerId}/deactivate`, {
                method: 'PUT',
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to deactivate embedding provider');
            }
            setFlashMessage({type: 'success', message: data.message});
            fetchEmbeddingProviders();
        } catch (err) {
            setFlashMessage({type: 'error', message: err instanceof Error ? err.message : 'An error occurred'});
        }
    };

    const configuredProviderItems = configuredProviders.map(provider => ({
        id: provider.uuid,
        name: provider.name,
        status: provider.status,
        sourceType: provider.provider,
        imageUrl: availableEmbeddingProviders.find(p => p.id === provider.provider)?.imageUrl ?? '/default-provider-icon.png',
        onDelete: handleDelete,
        onActivate: provider.status === 'inactive' ? handleActivate : undefined,
        onDeactivate: provider.status === 'active' ? handleDeactivate : undefined,
    }));

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    {flashMessage && (
                        <Alert variant={flashMessage.type === 'success' ? 'default' : 'destructive'}>
                            <AlertDescription>{flashMessage.message}</AlertDescription>
                        </Alert>
                    )}

                    <div className="flex gap-6">
                        <ListComponent
                            title="Configured Providers"
                            type="embedding"
                            items={configuredProviderItems}
                            loading={loading}
                            error={error}
                        />
                        <AvailableProviders
                            title="Available Providers"
                            providers={availableEmbeddingProviders.map(provider => ({
                                ...provider,
                                configurationUrl: provider.configurationUrl
                            }))}
                        />
                    </div>
                </div>
            </main>
        </div>
    );
};

export default EmbeddingProvidersPage;