'use client';

import React, { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import Navbar from '../../../../components/Navbar';
import Header, { HeaderConfig } from '../../../../components/Header';
import { getDatasourceById } from '@/utils/datasources';
import { SlackDatasourcePayload } from '@/types/datasource';
import AuthService from "@/services/authService";
import {createApiService} from "@/services/apiService";

const SlackConfigPage: React.FC = () => {
    const router = useRouter();
    const slackDatasource = getDatasourceById('slack');
    const datasourcesApi = createApiService(AuthService).datasource;

    const headerConfig: HeaderConfig = {
        title: `Configure ${slackDatasource?.name ?? 'Slack'} Datasource`
    };

    const [formData, setFormData] = useState({
        name: '',
        workspace: '',
        token: '',
        channel_id: ''
    });
    const [isLoading, setIsLoading] = useState(false);
    const [isValidated, setIsValidated] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const buttonText = isLoading ? 'Validating...' : isValidated ? 'Validated' : 'Validate';

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setFormData(prevData => ({ ...prevData, [name]: value }));
    };

    const handleValidate = useCallback(async () => {
        setIsLoading(true);
        setError(null);
        try {
            // Implement your validation logic here
            await new Promise(resolve => setTimeout(resolve, 1000)); // Simulating API call
            setIsValidated(true);
        } catch (err) {
            setError('Validation failed. Please check your inputs and try again.');
        } finally {
            setIsLoading(false);
        }
    }, []);

    const handleSave = useCallback(async () => {
        setIsLoading(true);
        setError(null);
        try {
            const payload: SlackDatasourcePayload = {
                name: formData.name,
                source_type: 'slack',
                settings: {
                    workspace: formData.workspace,
                    token: formData.token,
                    channel_id: formData.channel_id,
                }
            };

            await datasourcesApi.addSlackDatasource(payload);
            router.push('/datasources');
        } catch (err) {
            setError('Failed to create datasource. Please try again.');
        } finally {
            setIsLoading(false);
        }
    }, [formData, datasourcesApi, router]);

    if (!slackDatasource) {
        return <div>Slack datasource configuration not found.</div>;
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    <div className="flex items-center mb-6">
                        <Image src={slackDatasource.imageUrl} alt={`${slackDatasource.name} Logo`} width={48}
                               height={48} className="mr-4"/>
                        <h1 className="text-3xl font-bold text-gray-900">Configure {slackDatasource.name} Datasource</h1>
                    </div>

                    <div className="flex flex-col md:flex-row gap-8">
                        <div className="w-full md:w-1/2">
                            <div className="bg-white shadow sm:rounded-lg">
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
                                                placeholder="e.g., My Slack Workspace"
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
                                                placeholder="your-workspace"
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
                                                placeholder="xoxb-your-token"
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
                                                placeholder="C0123456789"
                                            />
                                        </div>
                                        {error && (
                                            <div className="text-red-600 text-sm">{error}</div>
                                        )}
                                        <div className="flex justify-end space-x-4">
                                            <button
                                                type="button"
                                                onClick={handleValidate}
                                                disabled={isLoading || isValidated}
                                                className={`inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white ${
                                                    isLoading || isValidated ? 'bg-gray-400' : 'bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500'
                                                }`}
                                            >
                                                {buttonText}
                                            </button>
                                            {isValidated && (
                                                <button
                                                    type="button"
                                                    onClick={handleSave}
                                                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
                                                >
                                                    Save Datasource
                                                </button>
                                            )}
                                        </div>
                                    </form>
                                </div>
                            </div>
                        </div>

                        {/* Right column: Instructions and Important Information */}
                        <div className="w-full md:w-1/2">
                            <div className="bg-white shadow sm:rounded-lg mb-8">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Instructions</h2>
                                    <div className="prose prose-blue text-gray-500">
                                        <ol className="list-decimal list-inside space-y-2">
                                            <li>Create a Slack App in your workspace if you haven't already.</li>
                                            <li>Generate a Bot User OAuth Token for your app.</li>
                                            <li>Invite the bot to the channels you want to index.</li>
                                            <li>Enter your Slack workspace name and the Bot User OAuth Token.</li>
                                            <li>Optionally, specify a channel ID to limit indexing to a specific
                                                channel.
                                            </li>
                                            <li>Click "Validate" to test your configuration.</li>
                                            <li>If validation is successful, click "Save Datasource" to complete the
                                                setup.
                                            </li>
                                        </ol>
                                    </div>
                                </div>
                            </div>

                            <div className="bg-white shadow sm:rounded-lg">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Important
                                        Information</h2>
                                    <p className="mb-4 text-sm text-gray-600">
                                        Ensure that your Slack App has the necessary permissions to access the channels
                                        you want to index.
                                        You may need to add scopes such as `channels:history`, `channels:read`, and
                                        `users:read` to your app.
                                    </p>
                                    <p className="mb-4 text-sm text-gray-600">
                                        For more information on creating and configuring Slack Apps, please refer to the
                                        <a href="https://api.slack.com/start" target="_blank" rel="noopener noreferrer"
                                           className="text-blue-600 hover:text-blue-800"> Slack API documentation</a>.
                                    </p>
                                    <div
                                        className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 text-sm"
                                        role="alert">
                                        <p className="font-bold">Note on Privacy and Data Access</p>
                                        <p>
                                            By configuring this datasource, you're allowing the system to access and
                                            index content from your Slack workspace.
                                            Ensure you have the necessary permissions and comply with your
                                            organization's data policies.
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

export default SlackConfigPage;