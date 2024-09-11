'use client'

import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import axios, { CancelTokenSource } from 'axios';
import Navbar from '../../components/Navbar';
import Header, { HeaderConfig } from '../../components/Header';
import { LLMProviderConfig } from '@/types/llmProvider';
import { availableLLMProviders } from '@/utils/llmProviders';
import { Alert, AlertDescription } from '@/components/Alert';
import ListComponent from "@/components/ListComponent";
import AvailableProviders from "@/components/AvailableProviders";
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const LLMProvidersPage: React.FC = () => {
    const [configuredProviders, setConfiguredProviders] = useState<LLMProviderConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const llmProviderApi = useMemo(() => createApiService(AuthService).llmProvider, []);
    const cancelTokenSourceRef = useRef<CancelTokenSource | null>(null);

    const headerConfig: HeaderConfig = {
        title: "LLM Providers"
    };

    const fetchLLMProviders = useCallback(async () => {
        if (cancelTokenSourceRef.current) {
            cancelTokenSourceRef.current.cancel('Operation canceled due to new request.');
        }
        cancelTokenSourceRef.current = axios.CancelToken.source();

        setLoading(true);
        try {
            const data = await llmProviderApi.getLLMProviders(cancelTokenSourceRef.current.token);
            setConfiguredProviders(data.llm_providers);
            setError(null);
        } catch (err) {
            if (!axios.isCancel(err)) {
                setError('Failed to load LLM providers. Please try again later.');
            }
        } finally {
            setLoading(false);
        }
    }, [llmProviderApi]);

    useEffect(() => {
        fetchLLMProviders();

        return () => {
            if (cancelTokenSourceRef.current) {
                cancelTokenSourceRef.current.cancel('Component unmounted');
            }
        };
    }, [fetchLLMProviders]);

    const handleAction = useCallback(async (action: 'delete' | 'activate' | 'deactivate', providerId: string) => {
        const source = axios.CancelToken.source();
        try {
            let message: string;
            switch (action) {
                case 'delete':
                    await llmProviderApi.deleteLLMProvider(providerId, source.token);
                    message = 'LLM provider deleted successfully';
                    break;
                case 'activate':
                    const activateData = await llmProviderApi.activateLLMProvider(providerId, source.token);
                    message = activateData.message;
                    break;
                case 'deactivate':
                    const deactivateData = await llmProviderApi.deactivateLLMProvider(providerId, source.token);
                    message = deactivateData.message;
                    break;
            }
            setFlashMessage({ type: 'success', message });
            await fetchLLMProviders();
        } catch (err) {
            if (!axios.isCancel(err)) {
                setFlashMessage({
                    type: 'error',
                    message: action === 'delete'
                        ? 'Failed to delete LLM provider. Please try again.'
                        : err instanceof Error ? err.message : `Failed to ${action} LLM provider. Please try again.`
                });
            }
        }
    }, [llmProviderApi, fetchLLMProviders]);

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
        imageUrl: availableLLMProviders.find(p => p.id === provider.provider)?.imageUrl || '/default-provider-icon.png',
        onDelete: handleDelete,
        onActivate: provider.status === 'inactive' ? () => handleAction('activate', provider.uuid) : undefined,
        onDeactivate: provider.status === 'active' ? () => handleAction('deactivate', provider.uuid) : undefined,
    })), [configuredProviders, handleDelete, handleAction]);

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
                            type="llm"
                            items={configuredProviderItems}
                            loading={loading}
                            error={error}
                        />
                        <AvailableProviders
                            title="Available Providers"
                            providers={availableLLMProviders.map(provider => ({
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

export default LLMProvidersPage;