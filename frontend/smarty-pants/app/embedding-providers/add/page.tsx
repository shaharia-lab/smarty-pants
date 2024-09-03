'use client';

import React, {useState} from 'react';
import {useRouter} from 'next/navigation';
import Navbar from '../../../components/Navbar';
import Header, {HeaderConfig} from '../../../components/Header';

const AddEmbeddingProviderPage: React.FC = () => {
    const [name, setName] = useState('');
    const [provider, setProvider] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [modelId, setModelId] = useState('');
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();

    const headerConfig: HeaderConfig = {
        title: "Add Embedding Provider"
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/embedding-provider`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name,
                    provider,
                    configuration: {
                        api_key: apiKey,
                        model_id: modelId,
                    },
                }),
            });

            if (!response.ok) {
                throw new Error('Failed to add embedding provider');
            }

            router.push('/embedding-providers');
        } catch (err) {
            setError('Failed to add embedding provider. Please try again.');
        }
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                <div className="px-4 py-6 sm:px-0">
                    <form onSubmit={handleSubmit} className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4">
                        <div className="mb-4">
                            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="name">
                                Name
                            </label>
                            <input
                                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                id="name"
                                type="text"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                            />
                        </div>
                        <div className="mb-4">
                            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="provider">
                                Provider
                            </label>
                            <select
                                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                id="provider"
                                value={provider}
                                onChange={(e) => setProvider(e.target.value)}
                                required
                            >
                                <option value="">Select a provider</option>
                                <option value="openai">OpenAI</option>
                                {/* Add more provider options as needed */}
                            </select>
                        </div>
                        <div className="mb-4">
                            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="apiKey">
                                API Key
                            </label>
                            <input
                                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                id="apiKey"
                                type="password"
                                value={apiKey}
                                onChange={(e) => setApiKey(e.target.value)}
                                required
                            />
                        </div>
                        <div className="mb-6">
                            <label className="block text-gray-700 text-sm font-bold mb-2" htmlFor="modelId">
                                Model ID
                            </label>
                            <input
                                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                                id="modelId"
                                type="text"
                                value={modelId}
                                onChange={(e) => setModelId(e.target.value)}
                                required
                            />
                        </div>
                        {error && <p className="text-red-500 text-xs italic mb-4">{error}</p>}
                        <div className="flex items-center justify-between">
                            <button
                                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
                                type="submit"
                            >
                                Add Provider
                            </button>
                        </div>
                    </form>
                </div>
            </main>
        </div>
    );
};

export default AddEmbeddingProviderPage;