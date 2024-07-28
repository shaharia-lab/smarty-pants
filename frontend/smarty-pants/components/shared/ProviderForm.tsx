import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ProviderConfig } from '@/types/provider';

interface ProviderFormProps {
    providerId?: string;
    providerType: 'embedding' | 'llm';
    initialData?: ProviderConfig;
    onSubmit: (data: Partial<ProviderConfig>) => Promise<void>;
}

const ProviderForm: React.FC<ProviderFormProps> = ({
                                                       providerId,
                                                       providerType,
                                                       initialData,
                                                       onSubmit,
                                                   }) => {
    const [name, setName] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [modelId, setModelId] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [isValidated, setIsValidated] = useState(false);
    const router = useRouter();

    const isEditMode = !!providerId;

    useEffect(() => {
        if (initialData) {
            setName(initialData.name);
            setApiKey(initialData.configuration.api_key);
            setModelId(initialData.configuration.model_id);
            setIsValidated(true);
        }
    }, [initialData]);

    const handleValidate = async () => {
        // Implement validation logic here
        setIsValidated(true);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        try {
            await onSubmit({
                name,
                provider: 'openai', // This should be dynamic based on the selected provider
                configuration: {
                    api_key: apiKey,
                    model_id: modelId,
                },
            });
            router.push(`/${providerType}`);
        } catch (err) {
            setError(`Failed to ${isEditMode ? 'update' : 'add'} provider. Please try again.`);
        }
    };

    return (
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
                <label className="block text-sm font-medium text-gray-700" htmlFor="modelId">
                    Model ID
                </label>
                <select
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    id="modelId"
                    value={modelId}
                    onChange={(e) => setModelId(e.target.value)}
                    required
                >
                    {providerType === 'embedding' ? (
                        <>
                            <option value="text-embedding-ada-002">ada v2</option>
                            <option value="text-embedding-3-small">Text Embedding 3 Small</option>
                            <option value="text-embedding-3-large">Text Embedding 3 Large</option>
                        </>
                    ) : (
                        <>
                            <option value="gpt-4">gpt-4</option>
                            <option value="gpt-4-turbo-preview">GPT-4 Turbo (Preview)</option>
                            <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                        </>
                    )}
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
    );
};

export default ProviderForm;