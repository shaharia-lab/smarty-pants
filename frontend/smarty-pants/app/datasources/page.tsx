'use client';

import React, {useEffect, useState} from 'react';
import Navbar from '../../components/Navbar';
import Header, {HeaderConfig} from '../../components/Header';
import {DatasourceConfig} from '@/types/datasource';
import {availableDatasources} from '@/utils/datasources';
import {Alert, AlertDescription} from '@/components/Alert';
import ListComponent from '../../components/ListComponent';
import AvailableProviders from '../../components/AvailableProviders';

interface DatasourcesApiResponse {
    datasources: DatasourceConfig[];
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const DatasourcesPage: React.FC = () => {
    const [configuredDatasources, setConfiguredDatasources] = useState<DatasourceConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const headerConfig: HeaderConfig = {
        title: "Datasources"
    };

    useEffect(() => {
        fetchDatasources();
    }, []);

    const fetchDatasources = async () => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource`);
            if (!response.ok) {
                throw new Error('Failed to fetch datasources');
            }
            const data: DatasourcesApiResponse = await response.json();
            setConfiguredDatasources(data.datasources);
        } catch (err) {
            setError('Failed to load datasources. Please try again later.');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (datasourceId: string) => {
        if (window.confirm('Are you sure you want to delete this datasource?')) {
            try {
                const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource/${datasourceId}`, {
                    method: 'DELETE',
                });
                if (!response.ok) {
                    throw new Error('Failed to delete datasource');
                }
                setFlashMessage({type: 'success', message: 'Datasource deleted successfully'});
                fetchDatasources();
            } catch (err) {
                setFlashMessage({type: 'error', message: 'Failed to delete datasource. Please try again.'});
            }
        }
    };

    const handleActivate = async (datasourceId: string) => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource/${datasourceId}/activate`, {
                method: 'PUT',
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to activate datasource');
            }
            setFlashMessage({type: 'success', message: data.message});
            fetchDatasources();
        } catch (err) {
            setFlashMessage({type: 'error', message: err instanceof Error ? err.message : 'An error occurred'});
        }
    };

    const handleDeactivate = async (datasourceId: string) => {
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource/${datasourceId}/deactivate`, {
                method: 'PUT',
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to deactivate datasource');
            }
            setFlashMessage({type: 'success', message: data.message});
            fetchDatasources();
        } catch (err) {
            setFlashMessage({type: 'error', message: err instanceof Error ? err.message : 'An error occurred'});
        }
    };

    const configuredDatasourceItems = configuredDatasources.map(datasource => ({
        id: datasource.uuid,
        name: datasource.name,
        sourceType: datasource.source_type,
        status: datasource.status,
        imageUrl: availableDatasources.find(d => d.id === datasource.source_type)?.imageUrl ?? '/default-datasource-icon.png',
        onDelete: handleDelete,
        onActivate: datasource.status === 'inactive' ? handleActivate : undefined,
        onDeactivate: datasource.status === 'active' ? handleDeactivate : undefined,
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