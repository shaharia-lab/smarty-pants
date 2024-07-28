// components/shared/ProviderDetails.tsx
import React from 'react';

interface ProviderDetailsProps {
    providerType: 'embedding' | 'llm';
    isEditMode: boolean;
}

const ProviderDetails: React.FC<ProviderDetailsProps> = ({providerType, isEditMode}) => {
    return (
        <div>
            <div className="bg-white shadow sm:rounded-lg mb-8">
                <div className="px-4 py-5 sm:p-6">
                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Instructions</h2>
                    <div className="prose prose-blue text-gray-500">
                        <ol className="list-decimal list-inside space-y-2">
                            <li>Create an OpenAI account if you haven't already.</li>
                            <li>Generate an API key from your OpenAI dashboard.</li>
                            <li>Choose the appropriate {providerType} model for your needs.</li>
                            <li>Enter a name for this configuration, your API key, and select the model in the form.
                            </li>
                            {!isEditMode && <li>Click "Validate" to test your configuration.</li>}
                            <li>Click "{isEditMode ? 'Update' : 'Save'} Provider" to complete the setup.</li>
                        </ol>
                    </div>
                </div>
            </div>

            <div className="bg-white shadow sm:rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                    <h2 className="text-lg leading-6 font-medium text-gray-900 mb-4">Important Information</h2>
                    <h3 className="text-md font-semibold mb-2">{providerType === 'embedding' ? 'Embedding' : 'LLM'} Pricing:</h3>
                    <ul className="list-disc pl-5 mb-4 text-sm text-gray-600">
                        {providerType === 'embedding' ? (
                            <>
                                <li>text-embedding-3-small: $0.02 / 1M tokens</li>
                                <li>text-embedding-3-large: $0.13 / 1M tokens</li>
                                <li>ada v2: $0.10 / 1M tokens</li>
                            </>
                        ) : (
                            <>
                                <li>GPT-4: $0.03 / 1K tokens (prompt), $0.06 / 1K tokens (completion)</li>
                                <li>GPT-4 Turbo: $0.01 / 1K tokens (prompt), $0.03 / 1K tokens (completion)</li>
                                <li>GPT-3.5 Turbo: $0.0015 / 1K tokens (prompt), $0.002 / 1K tokens (completion)</li>
                            </>
                        )}
                    </ul>
                    <p className="mb-4 text-sm text-gray-600">For up-to-date pricing, please visit the <a
                        href="https://openai.com/api/pricing/" target="_blank" rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-800">OpenAI API Pricing page</a>.</p>
                    <p className="mb-4 text-sm text-gray-600">Read more about
                        OpenAI {providerType === 'embedding' ? 'embeddings' : 'language models'} in the <a
                            href={`https://platform.openai.com/docs/guides/${providerType === 'embedding' ? 'embeddings' : 'language-models'}`}
                            target="_blank" rel="noopener noreferrer"
                            className="text-blue-600 hover:text-blue-800">OpenAI {providerType === 'embedding' ? 'Embeddings' : 'Language Models'} Guide</a>.
                    </p>
                    {providerType === 'embedding' && (
                        <div className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 text-sm"
                             role="alert">
                            <p className="font-bold">Warning: Dimensions</p>
                            <p>We shouldn't change the embedding dimensions because if we change the dimensions, in the
                                database backend the vector indexing may corrupt and require running embedding on all
                                documents all over again.</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ProviderDetails;