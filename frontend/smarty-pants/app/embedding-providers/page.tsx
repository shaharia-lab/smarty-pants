'use client';

import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import axios, { CancelTokenSource } from 'axios';
import Navbar from '../../components/Navbar';
import Header, { HeaderConfig } from '../../components/Header';
import { EmbeddingProviderConfig } from '@/types/embeddingProvider';
import { availableEmbeddingProviders } from '@/utils/embeddingProviders';
import { Alert, AlertDescription } from '@/components/Alert';
import ListComponent from '../../components/ListComponent';
import AvailableProviders from '../../components/AvailableProviders';
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const EmbeddingProvidersPage: React.FC = () => {
    const [configuredProviders, setConfiguredProviders] = useState<EmbeddingProviderConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const embeddingProviderApi = useMemo(() => createApiService(AuthService).embeddingProvider, []);
    const cancelTokenSourceRef = useRef<CancelTokenSource | null>(null);

    const headerConfig: HeaderConfig = {
        title: "Embedding Providers"
    };

    const fetchEmbeddingProviders = useCallback(async () => {
        if (cancelTokenSourceRef.current) {
            cancelTokenSourceRef.current.cancel('Operation canceled due to new request.');
        }
        cancelTokenSourceRef.current = axios.CancelToken.source();

        setLoading(true);
        try {
            const data = await embeddingProviderApi.getEmbeddingProviders(cancelTokenSourceRef.current.token);
            setConfiguredProviders(data.embedding_providers);
            setError(null);
        } catch (err) {
            if (!axios.isCancel(err)) {
                setError('Failed to load embedding providers. Please try again later.');
            }
        } finally {
            setLoading(false);
        }
    }, [embeddingProviderApi]);

    useEffect(() => {
        fetchEmbeddingProviders();

        return () => {
            if (cancelTokenSourceRef.current) {
                cancelTokenSourceRef.current.cancel('Component unmounted');
            }
        };
    }, [fetchEmbeddingProviders]);

    const handleAction = useCallback(async (action: 'delete' | 'activate' | 'deactivate', providerId: string) => {
        const source = axios.CancelToken.source();
        try {
            let message: string;
            switch (action) {
                case 'delete':
                    await embeddingProviderApi.deleteEmbeddingProvider(providerId, source.token);
                    message = 'Embedding provider deleted successfully';
                    break;
                case 'activate':
                    const activateData = await embeddingProviderApi.activateEmbeddingProvider(providerId, source.token);
                    message = activateData.message;
                    break;
                case 'deactivate':
                    const deactivateData = await embeddingProviderApi.deactivateEmbeddingProvider(providerId, source.token);
                    message = deactivateData.message;
                    break;
            }
            setFlashMessage({ type: 'success', message });
            await fetchEmbeddingProviders();
        } catch (err) {
            if (!axios.isCancel(err)) {
                setFlashMessage({
                    type: 'error',
                    message: err instanceof Error ? err.message : `Failed to ${action} embedding provider. Please try again.`
                });
            }
        }
    }, [embeddingProviderApi, fetchEmbeddingProviders]);

    const handleDelete = useCallback((providerId: string) => {
        if (window.confirm('Are you sure you want to delete this provider?')) {
            handleAction('delete', providerId);
        }
    }, [handleAction]);

    const configuredProviderItems = useMemo(() => configuredProviders.map(provider => ({
        id: provider.uuid,
        name: provider.name,
        status: provider.status,
        sourceType: provider.provider,
        imageUrl: availableEmbeddingProviders.find(p => p.id === provider.provider)?.imageUrl ?? '/default-provider-icon.png',
        onDelete: handleDelete,
        onActivate: provider.status === 'inactive' ? () => handleAction('activate', provider.uuid) : undefined,
        onDeactivate: provider.status === 'active' ? () => handleAction('deactivate', provider.uuid) : undefined,
    })), [configuredProviders, handleDelete, handleAction]);

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar />
            <Header config={headerConfig} />
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