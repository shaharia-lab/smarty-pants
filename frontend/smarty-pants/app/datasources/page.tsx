// File: src/pages/datasources/index.tsx

'use client';

import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import axios, { CancelTokenSource } from 'axios';
import Navbar from '../../components/Navbar';
import Header, { HeaderConfig } from '../../components/Header';
import { DatasourceConfig } from '@/types/datasource';
import { availableDatasources } from '@/utils/datasources';
import { Alert, AlertDescription } from '@/components/Alert';
import ListComponent from '../../components/ListComponent';
import AvailableProviders from '../../components/AvailableProviders';
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const DatasourcesPage: React.FC = () => {
    const [configuredDatasources, setConfiguredDatasources] = useState<DatasourceConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const datasourcesApi = useMemo(() => createApiService(AuthService).datasource, []);
    const cancelTokenSourceRef = useRef<CancelTokenSource | null>(null);

    const headerConfig: HeaderConfig = {
        title: "Datasources"
    };

    const fetchDatasources = useCallback(async () => {
        if (cancelTokenSourceRef.current) {
            cancelTokenSourceRef.current.cancel('Operation canceled due to new request.');
        }
        cancelTokenSourceRef.current = axios.CancelToken.source();

        setLoading(true);
        try {
            const data = await datasourcesApi.getDatasources(cancelTokenSourceRef.current.token);
            setConfiguredDatasources(data.datasources);
            setError(null);
        } catch (err) {
            if (!axios.isCancel(err)) {
                setError('Failed to load datasources. Please try again later.');
            }
        } finally {
            setLoading(false);
        }
    }, [datasourcesApi]);

    useEffect(() => {
        fetchDatasources();

        return () => {
            if (cancelTokenSourceRef.current) {
                cancelTokenSourceRef.current.cancel('Component unmounted');
            }
        };
    }, [fetchDatasources]);

    const handleAction = useCallback(async (action: 'delete' | 'activate' | 'deactivate', datasourceId: string) => {
        const source = axios.CancelToken.source();
        try {
            let message: string;
            switch (action) {
                case 'delete':
                    await datasourcesApi.deleteDatasource(datasourceId, source.token);
                    message = 'Datasource deleted successfully';
                    break;
                case 'activate':
                    const activateData = await datasourcesApi.activateDatasource(datasourceId, source.token);
                    message = activateData.message;
                    break;
                case 'deactivate':
                    const deactivateData = await datasourcesApi.deactivateDatasource(datasourceId, source.token);
                    message = deactivateData.message;
                    break;
            }
            setFlashMessage({ type: 'success', message });
            await fetchDatasources();
        } catch (err) {
            if (!axios.isCancel(err)) {
                setFlashMessage({
                    type: 'error',
                    message: err instanceof Error ? err.message : `Failed to ${action} datasource. Please try again.`
                });
            }
        }
    }, [datasourcesApi, fetchDatasources]);

    const handleDelete = useCallback((datasourceId: string) => {
        if (window.confirm('Are you sure you want to delete this datasource?')) {
            handleAction('delete', datasourceId);
        }
    }, [handleAction]);

    const configuredDatasourceItems = useMemo(() => configuredDatasources.map(datasource => ({
        id: datasource.uuid,
        name: datasource.name,
        sourceType: datasource.source_type,
        status: datasource.status,
        imageUrl: availableDatasources.find(d => d.id === datasource.source_type)?.imageUrl ?? '/default-datasource-icon.png',
        onDelete: handleDelete,
        onActivate: datasource.status === 'inactive' ? () => handleAction('activate', datasource.uuid) : undefined,
        onDeactivate: datasource.status === 'active' ? () => handleAction('deactivate', datasource.uuid) : undefined,
    })), [configuredDatasources, handleDelete, handleAction]);

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
                            title="Configured Datasources"
                            items={configuredDatasourceItems}
                            loading={loading}
                            error={error}
                            type="datasource"
                        />
                        <AvailableProviders
                            title="Available Datasources"
                            providers={availableDatasources.map(datasource => ({
                                id: datasource.id,
                                name: datasource.name,
                                description: datasource.description,
                                imageUrl: datasource.imageUrl,
                                configurationUrl: datasource.configurationUrl,
                            }))}
                        />
                    </div>
                </div>
            </main>
        </div>
    );
};

export default DatasourcesPage;