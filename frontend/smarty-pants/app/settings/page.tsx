'use client';

import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import axios, { CancelTokenSource } from 'axios';
import Navbar from '../../components/Navbar';
import Header, { HeaderConfig } from '../../components/Header';
import { Settings } from '@/services/api/settings';
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";
import { Alert, AlertDescription } from '@/components/Alert';

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const SettingsPage: React.FC = () => {
    const [settings, setSettings] = useState<Settings>({
        general: { application_name: '' },
        debugging: { log_level: 'info', log_format: 'json', log_output: 'stdout' },
        search: { per_page: 10 },
    });
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const settingsApi = useMemo(() => createApiService(AuthService).settingsApi, []);
    const cancelTokenSourceRef = useRef<CancelTokenSource | null>(null);

    const headerConfig: HeaderConfig = {
        title: "Settings"
    };

    const fetchSettings = useCallback(async () => {
        if (cancelTokenSourceRef.current) {
            cancelTokenSourceRef.current.cancel('Operation canceled due to new request.');
        }
        cancelTokenSourceRef.current = axios.CancelToken.source();

        setIsLoading(true);
        try {
            const data = await settingsApi.getSettings(cancelTokenSourceRef.current.token);
            setSettings(data);
            setError(null);
        } catch (err) {
            if (!axios.isCancel(err)) {
                setError('Failed to load settings. Please try again later.');
            }
        } finally {
            setIsLoading(false);
        }
    }, [settingsApi]);

    useEffect(() => {
        fetchSettings();

        return () => {
            if (cancelTokenSourceRef.current) {
                cancelTokenSourceRef.current.cancel('Component unmounted');
            }
        };
    }, [fetchSettings]);

    const handleInputChange = (section: keyof Settings, key: string, value: string | number) => {
        setSettings(prevSettings => ({
            ...prevSettings,
            [section]: {
                ...prevSettings[section],
                [key]: value,
            },
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setFlashMessage(null);
        try {
            await settingsApi.updateSettings(settings);
            setFlashMessage({ type: 'success', message: 'Settings updated successfully' });
        } catch (err) {
            setFlashMessage({ type: 'error', message: 'Failed to update settings. Please try again.' });
        }
    };

    if (isLoading) {
        return <div>Loading settings...</div>;
    }

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

                    <form onSubmit={handleSubmit} className="space-y-8 divide-y divide-gray-200">
                        {/* General Settings */}
                        <div>
                            <div className="flex items-center">
                                <svg className="h-6 w-6 text-gray-600 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                                <h3 className="text-lg leading-6 font-medium text-gray-900">General Settings</h3>
                            </div>
                            <p className="mt-1 max-w-2xl text-sm text-gray-500">
                                Configure general application settings.
                            </p>
                            <div className="mt-6 sm:mt-5 space-y-6 sm:space-y-5">
                                <div className="sm:grid sm:grid-cols-3 sm:gap-4 sm:items-start sm:border-t sm:border-gray-200 sm:pt-5">
                                    <label htmlFor="application_name" className="block text-gray-700 text-sm font-bold mb-2 sm:mt-px sm:pt-2">
                                        Application Name
                                    </label>
                                    <div className="mt-1 sm:mt-0 sm:col-span-2">
                                        <input
                                            type="text"
                                            name="application_name"
                                            id="application_name"
                                            value={settings.general.application_name}
                                            onChange={(e) => handleInputChange('general', 'application_name', e.target.value)}
                                            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                            required
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Debugging Settings */}
                        <div className="pt-8 space-y-6 sm:pt-10 sm:space-y-5">
                            <div className="flex items-center">
                                <svg className="h-6 w-6 text-gray-600 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                                </svg>
                                <h3 className="text-lg leading-6 font-medium text-gray-900">Debugging Settings</h3>
                            </div>
                            <p className="mt-1 max-w-2xl text-sm text-gray-500">
                                Configure debugging and logging settings.
                            </p>
                            <div className="space-y-6 sm:space-y-5">
                                <div className="sm:grid sm:grid-cols-3 sm:gap-4 sm:items-start sm:border-t sm:border-gray-200 sm:pt-5">
                                    <label htmlFor="log_level" className="block text-gray-700 text-sm font-bold mb-2 sm:mt-px sm:pt-2">
                                        Log Level
                                    </label>
                                    <div className="mt-1 sm:mt-0 sm:col-span-2">
                                        <select
                                            id="log_level"
                                            name="log_level"
                                            value={settings.debugging.log_level}
                                            onChange={(e) => handleInputChange('debugging', 'log_level', e.target.value)}
                                            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                            required
                                        >
                                            <option value="debug">Debug</option>
                                            <option value="info">Info</option>
                                            <option value="warn">Warn</option>
                                            <option value="error">Error</option>
                                        </select>
                                    </div>
                                </div>
                                <div className="sm:grid sm:grid-cols-3 sm:gap-4 sm:items-start sm:border-t sm:border-gray-200 sm:pt-5">
                                    <label htmlFor="log_format" className="block text-gray-700 text-sm font-bold mb-2 sm:mt-px sm:pt-2">
                                        Log Format
                                    </label>
                                    <div className="mt-1 sm:mt-0 sm:col-span-2">
                                        <select
                                            id="log_format"
                                            name="log_format"
                                            value={settings.debugging.log_format}
                                            onChange={(e) => handleInputChange('debugging', 'log_format', e.target.value)}
                                            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                            required
                                        >
                                            <option value="json">JSON</option>
                                            <option value="text">Text</option>
                                        </select>
                                    </div>
                                </div>
                                <div className="sm:grid sm:grid-cols-3 sm:gap-4 sm:items-start sm:border-t sm:border-gray-200 sm:pt-5">
                                    <label htmlFor="log_output" className="block text-gray-700 text-sm font-bold mb-2 sm:mt-px sm:pt-2">
                                        Log Output
                                    </label>
                                    <div className="mt-1 sm:mt-0 sm:col-span-2">
                                        <select
                                            id="log_output"
                                            name="log_output"
                                            value={settings.debugging.log_output}
                                            onChange={(e) => handleInputChange('debugging', 'log_output', e.target.value)}
                                            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                            required
                                        >
                                            <option value="stdout">stdout</option>
                                            <option value="stderr">stderr</option>
                                            <option value="file">file</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Search Settings */}
                        <div className="pt-8 space-y-6 sm:pt-10 sm:space-y-5">
                            <div className="flex items-center">
                                <svg className="h-6 w-6 text-gray-600 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                                </svg>
                                <h3 className="text-lg leading-6 font-medium text-gray-900">Search Settings</h3>
                            </div>
                            <p className="mt-1 max-w-2xl text-sm text-gray-500">
                                Configure search-related settings.
                            </p>
                            <div className="space-y-6 sm:space-y-5">
                                <div className="sm:grid sm:grid-cols-3 sm:gap-4 sm:items-start sm:border-t sm:border-gray-200 sm:pt-5">
                                    <label htmlFor="per_page" className="block text-gray-700 text-sm font-bold mb-2 sm:mt-px sm:pt-2">
                                        Results Per Page
                                    </label>
                                    <div className="mt-1 sm:mt-0 sm:col-span-2">
                                        <input
                                            type="number"
                                            name="per_page"
                                            id="per_page"
                                            value={settings.search.per_page}
                                            onChange={(e) => handleInputChange('search', 'per_page', parseInt(e.target.value))}
                                            min="1"
                                            max="100"
                                            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                            required
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="pt-5">
                            <div className="flex justify-end">
                                <button
                                    type="button"
                                    onClick={fetchSettings}
                                    className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                >
                                    Reset
                                </button>
                                <button
                                    type="submit"
                                    className="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                >
                                    Save Settings
                                </button>
                            </div>
                        </div>
                    </form>

                    {error && (
                        <div className="mt-4 text-red-600">
                            {error}
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
};

export default SettingsPage;