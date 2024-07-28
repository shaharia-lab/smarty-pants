// File: /components/OpenAILLMProviderForm.tsx

'use client';

import React, {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';
import Navbar from './Navbar';
import Header, {HeaderConfig} from './Header';

interface OpenAILLMProviderFormProps {
    providerId?: string;
}

const OpenAILLMProviderForm: React.FC<OpenAILLMProviderFormProps> = ({providerId}) => {
    const [name, setName] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [modelId, setModelId] = useState('text-llm-ada-002');
    const [error, setError] = useState<string | null>(null);
    const [isValidated, setIsValidated] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const router = useRouter();

    const isEditMode = !!providerId;

    const headerConfig: HeaderConfig = {
        title: isEditMode ? "Edit OpenAI LLM Provider" : "Add OpenAI LLM Provider",
    };

    useEffect(() => {
        if (isEditMode) {
            fetchProviderData();
        }
    }, [providerId]);

    const fetchProviderData = async () => {
        setIsLoading(true);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/llm-provider/${providerId}`);
            if (!response.ok) {
                throw new Error('Failed to fetch provider data');
            }
            const data = await response.json();
            setName(data.name);
            setApiKey(data.configuration.api_key);
            setModelId(data.configuration.model_id);
            setIsValidated(true);
        } catch (err) {
            setError('Failed to load provider data. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    const handleValidate = async () => {
        // Implement validation logic here
        setIsValidated(true);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        const url = isEditMode
            ? `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/llm-provider/${providerId}`
            : `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/llm-provider`;

        const method = isEditMode ? 'PUT' : 'POST';

        try {
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name,
                    provider: 'openai',
                    configuration: {
                        api_key: apiKey,
                        model_id: modelId,
                        encoding_format: 'float',
                        dimensions: 1536,
                    },
                }),
            });

            if (!response.ok) {
                throw new Error(`Failed to ${isEditMode ? 'update' : 'add'} llm provider`);
            }

            router.push('/llm-providers');
        } catch (err) {
            setError(`Failed to ${isEditMode ? 'update' : 'add'} llm provider. Please try again.`);
        }
    };

    if (isLoading) {
        return <div>Loading...</div>;
    }

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    <div className="flex flex-col md:flex-row gap-8">
                        {/* Left column: Form */}
                        <div className="w-full md:w-1/2">
                            <div className="bg-white shadow sm:rounded-lg">
                                <div className="px-4 py-5 sm:p-6">
                                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">OpenAI
                                        Configuration</h2>
                                    <form onSubmit={handleSubmit}>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700" htmlFor="name">
                                                Name
                                            </label>
                                            <input
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                id="name"
                                                type="text"
                                                value={name}
                                                onChange={(e) => setName(e.target.value)}
                                                required
                                            />
                                        </div>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700" htmlFor="apiKey">
                                                API Key
                                            </label>
                                            <input
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                id="apiKey"
                                                type="password"
                                                value={apiKey}
                                                onChange={(e) => setApiKey(e.target.value)}
                                                required
                                            />
                                        </div>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700"
                                                   htmlFor="modelId">
                                                Model ID
                                            </label>
                                            <select
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                                id="modelId"
                                                value={modelId}
                                                onChange={(e) => setModelId(e.target.value)}
                                                required
                                            >
                                                <option value="gpt-4">gpt-4</option>
                                                <option value="gpt-4-turbo-preview">GPT-4 Turbo (Preview)</option>
                                                <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                                            </select>
                                        </div>
                                        <div className="mb-4">
                                            <label className="block text-sm font-medium text-gray-700"
                                                   htmlFor="encodingFormat">
                                                Encoding Format
                                            </label>
                                            <input
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 bg-gray-100 sm:text-sm"
                                                id="encodingFormat"
                                                type="text"
                                                value="float"
                                                disabled
                                            />
                                        </div>
                                        <div className="mb-6">
                                            <label className="block text-sm font-medium text-gray-700"
                                                   htmlFor="dimensions">
                                                Dimensions
                                            </label>
                                            <select
                                                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 bg-gray-100 sm:text-sm"
                                                id="dimensions"
                                                value={1536}
                                                disabled
                                            >
                                                <option value={1536}>1536</option>
                                            </select>
                                        </div>
                                        {error && <p className="text-red-500 text-xs italic mb-4">{error}</p>}
                                        <div className="flex justify-end space-x-4">
                                            {!isEditMode && (
                                                <button
                                                    type="button"
                                                    onClick={handleValidate}
                                                    disabled={isValidated}
                                                    className={`inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white ${
                                                        isValidated ? 'bg-gray-400' : 'bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500'
                                                    }`}
                                                >
                                                    {isValidated ? 'Validated' : 'Validate'}
                                                </button>
                                            )}
                                            <button
                                                type="submit"
                                                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
                                            >
                                                {isEditMode ? 'Update Provider' : 'Save Provider'}
                                            </button>
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
                                            <li>Create an OpenAI account if you haven't already.</li>
                                            <li>Generate an API key from your OpenAI dashboard.</li>
                                            <li>Choose the appropriate llm model for your needs.</li>
                                            <li>Enter a name for this configuration, your API key, and select the model
                                                in the form.
                                            </li>
                                            {!isEditMode && <li>Click "Validate" to test your configuration.</li>}
                                            <li>Click "{isEditMode ? 'Update' : 'Save'} Provider" to complete the
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
                                    <h3 className="text-md font-semibold mb-2">LLM Pricing:</h3>
                                    <ul className="list-disc pl-5 mb-4 text-sm text-gray-600">
                                        <li>text-llm-3-small: $0.02 / 1M tokens</li>
                                        <li>text-llm-3-large: $0.13 / 1M tokens</li>
                                        <li>ada v2: $0.10 / 1M tokens</li>
                                    </ul>
                                    <p className="mb-4 text-sm text-gray-600">For up-to-date pricing, please visit
                                        the <a href="https://openai.com/api/pricing/" target="_blank"
                                               rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">OpenAI
                                            API Pricing page</a>.</p>
                                    <p className="mb-4 text-sm text-gray-600">Read more about OpenAI llms in the <a
                                        href="https://platform.openai.com/docs/guides/llms/what-are-llms"
                                        target="_blank" rel="noopener noreferrer"
                                        className="text-blue-600 hover:text-blue-800">OpenAI LLMs Guide</a>.</p>
                                    <div
                                        className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 text-sm"
                                        role="alert">
                                        <p className="font-bold">Warning: Dimensions</p>
                                        <p>We shouldn't change the llm dimensions because if we change the dimensions,
                                            in the database backend the vector indexing may corrupt and require running
                                            llm on all documents all over again.</p>
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

export default OpenAILLMProviderForm;