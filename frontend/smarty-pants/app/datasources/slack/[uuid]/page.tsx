'use client';

import React, {useCallback, useEffect, useState} from 'react';
import {useParams, useRouter} from 'next/navigation';
import Image from 'next/image';
import Navbar from '../../../../components/Navbar';
import Header, {HeaderConfig} from '../../../../components/Header';
import {getDatasourceById} from '@/utils/datasources';
import {DatasourceConfig, isSlackDatasource} from '@/types/datasource';
import {Alert, AlertDescription} from '@/components/Alert';

interface FlashMessage {
    type: 'success' | 'error';
    message: string;
}

const SlackEditPage: React.FC = () => {
    useRouter();
    const params = useParams();
    const uuid = params.uuid as string;
    const slackDatasource = getDatasourceById('slack');

    const [datasource, setDatasource] = useState<DatasourceConfig | null>(null);
    const [formData, setFormData] = useState({
        name: '',
        workspace: '',
        token: '',
        channel_id: ''
    });
    const [isLoading, setIsLoading] = useState(true);
    const [flashMessage, setFlashMessage] = useState<FlashMessage | null>(null);

    const headerConfig: HeaderConfig = {
        title: `Edit ${slackDatasource?.name ?? 'Slack'} Datasource`
    };

    const fetchDatasourceDetails  = useCallback(async () => {
        console.log("Fetching datasource details for UUID:", uuid);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource/${uuid}`);
            console.log("API Response:", response);
            if (!response.ok) {
                throw new Error(`Failed to fetch datasource details: ${response.statusText}`);
            }
            const data: DatasourceConfig = await response.json();
            console.log("Received data:", data);
            setDatasource(data);
            if (isSlackDatasource(data)) {
                setFormData({
                    name: data.name,
                    workspace: data.settings.workspace || '',
                    token: data.settings.token || '',
                    channel_id: data.settings.channel_id || ''
                });
            } else {
                throw new Error('Invalid datasource type');
            }
        } catch (err) {
            console.error("Error fetching datasource details:", err);
            setFlashMessage({type: 'error', message: 'Failed to load datasource details. Please try again.'});
        } finally {
            setIsLoading(false);
        }
    }, [uuid]);

    useEffect(() => {
        fetchDatasourceDetails();
    }, [fetchDatasourceDetails]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const {name, value} = e.target;
        setFormData(prevData => ({...prevData, [name]: value}));
    };

    const handleUpdate = async () => {
        setIsLoading(true);
        setFlashMessage(null);
        try {
            if (!datasource) throw new Error('Datasource not loaded');

            const updatedDatasource: DatasourceConfig = {
                ...datasource,
                name: formData.name,
                settings: {
                    ...datasource.settings,
                    workspace: formData.workspace,
                    token: formData.token,
                    channel_id: formData.channel_id
                }
            };

            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/datasource/${uuid}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updatedDatasource),
            });

            if (!response.ok) {
                throw new Error('Failed to update datasource');
            }

            setFlashMessage({type: 'success', message: 'Datasource updated successfully'});
        } catch (err) {
            setFlashMessage({type: 'error', message: 'Failed to update datasource. Please try again.'});
        } finally {
            setIsLoading(false);
        }
    };

    if (isLoading) {
        return <div>Loading...</div>;
    }

    if (!slackDatasource || !datasource || !isSlackDatasource(datasource)) {
        return <div>Invalid datasource configuration.</div>;
    }

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

                    <div className="flex items-center mb-6">
                        <Image src={slackDatasource.imageUrl} alt={`${slackDatasource.name} Logo`} width={48}
                               height={48} className="mr-4"/>
                        <h1 className="text-3xl font-bold text-gray-900">Edit {slackDatasource.name} Datasource</h1>
                    </div>

                    <div className="flex flex-col md:flex-row gap-8">
                        {/* Left column: Form and State Information */}
                        <div className="w-full md:w-1/2">
                            <div className="bg-white shadow sm:rounded-lg mb-8">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Slack
                                        Configuration</h2>
                                    <form className="space-y-6">
                                        <div>
                                            <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                                                Datasource Name
                                            </label>
                                            <input
                                                type="text"
                                                name="name"
                                                id="name"
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                value={formData.name}
                                                onChange={handleInputChange}
                                            />
                                        </div>
                                        <div>
                                            <label htmlFor="workspace"
                                                   className="block text-sm font-medium text-gray-700">
                                                Slack Workspace Name
                                            </label>
                                            <input
                                                type="text"
                                                name="workspace"
                                                id="workspace"
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                value={formData.workspace}
                                                onChange={handleInputChange}
                                            />
                                        </div>
                                        <div>
                                            <label htmlFor="token" className="block text-sm font-medium text-gray-700">
                                                Bot User OAuth Token
                                            </label>
                                            <input
                                                type="password"
                                                name="token"
                                                id="token"
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                value={formData.token}
                                                onChange={handleInputChange}
                                            />
                                        </div>
                                        <div>
                                            <label htmlFor="channel_id"
                                                   className="block text-sm font-medium text-gray-700">
                                                Channel ID (Optional)
                                            </label>
                                            <input
                                                type="text"
                                                name="channel_id"
                                                id="channel_id"
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                value={formData.channel_id}
                                                onChange={handleInputChange}
                                            />
                                        </div>
                                        <div className="flex justify-end">
                                            <button
                                                type="button"
                                                onClick={handleUpdate}
                                                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                                            >
                                                Update Datasource
                                            </button>
                                        </div>
                                    </form>
                                </div>
                            </div>

                            {/* State Information */}
                            {datasource.state && (
                                <div className="bg-white shadow sm:rounded-lg mt-8">
                                    <div className="px-4 py-5 sm:p-6">
                                        <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Datasource
                                            State</h2>
                                        <div className="space-y-2">
                                            {Object.entries(datasource.state).map(([key, value]) => (
                                                <p key={key}><strong>{key}:</strong> {JSON.stringify(value)}</p>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Right column: Instructions and Important Information */}
                        <div className="w-full md:w-1/2">
                            <div className="bg-white shadow sm:rounded-lg mb-8">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Instructions</h2>
                                    <div className="prose prose-blue text-gray-500">
                                        <ol className="list-decimal list-inside space-y-2">
                                            <li>Review and update the Slack datasource details as needed.</li>
                                            <li>Ensure the Bot User OAuth Token is still valid.</li>
                                            <li>If changing the workspace, make sure the bot is invited to the necessary
                                                channels.
                                            </li>
                                            <li>Update the Channel ID if you want to change or set a specific channel
                                                for indexing.
                                            </li>
                                            <li>Click Update Datasource to save your changes.</li>
                                        </ol>
                                    </div>
                                </div>
                            </div>

                            <div className="bg-white shadow sm:rounded-lg">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Important
                                        Information</h2>
                                    <p className="mb-4 text-sm text-gray-600">
                                        Changing the workspace or token may affect the indexing of your Slack content.
                                        Ensure that the new configuration has the necessary permissions.
                                    </p>
                                    <p className="mb-4 text-sm text-gray-600">
                                        For more information on managing Slack Apps, please refer to the
                                        <a href="https://api.slack.com/start" target="_blank" rel="noopener noreferrer"
                                           className="text-blue-600 hover:text-blue-800"> Slack API documentation</a>.
                                    </p>
                                    <div
                                        className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 text-sm"
                                        role="alert">
                                        <p className="font-bold">Caution</p>
                                        <p>
                                            Updating the datasource configuration may trigger a re-indexing process.
                                            This could temporarily affect the availability of Slack data in your
                                            searches.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

export default SlackEditPage;