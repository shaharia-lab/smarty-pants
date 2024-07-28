import React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { ProviderConfig, AvailableProvider } from '@/types/provider';

interface ProviderListProps {
    availableProviders: AvailableProvider[];
    configuredProviders: ProviderConfig[];
    onDelete: (providerId: string) => void;
    providerType: 'embedding' | 'llm';
}

const ProviderList: React.FC<ProviderListProps> = ({
                                                       availableProviders,
                                                       configuredProviders,
                                                       onDelete,
                                                       providerType,
                                                   }) => {
    return (
        <div>
            <section className="mb-12">
                <h2 className="text-2xl font-semibold text-gray-900 mb-6">Available {providerType === 'embedding' ? 'Embedding' : 'LLM'} Providers</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                    {availableProviders.map((provider) => (
                        <Link href={`/${providerType}/${provider.id}/add`} key={provider.id}>
                            <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300 overflow-hidden">
                                <div className="p-6">
                                    <div className="flex items-center mb-4">
                                        <Image
                                            src={provider.imageUrl}
                                            alt={`${provider.name} icon`}
                                            width={48}
                                            height={48}
                                            className="rounded-full"
                                        />
                                        <h3 className="ml-4 text-xl font-semibold text-gray-900">{provider.name}</h3>
                                    </div>
                                    <p className="text-gray-600 mb-4">{provider.description}</p>
                                    <div className="flex justify-end">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
                      Configure
                    </span>
                                    </div>
                                </div>
                            </div>
                        </Link>
                    ))}
                </div>
            </section>

            <section className="mt-12">
                <h2 className="text-2xl font-semibold text-gray-900 mb-6">Configured {providerType === 'embedding' ? 'Embedding' : 'LLM'} Providers</h2>
                <div className="bg-white shadow overflow-hidden sm:rounded-lg">
                    {configuredProviders.map((provider) => (
                        <div key={provider.uuid} className="border-b border-gray-200 last:border-b-0">
                            <div className="px-6 py-5 flex items-center justify-between hover:bg-gray-50 transition-colors duration-200">
                                <div className="flex items-center">
                                    <Image
                                        src={availableProviders.find(p => p.id === provider.provider)?.imageUrl || '/default-provider-icon.png'}
                                        alt={`${provider.provider} icon`}
                                        width={40}
                                        height={40}
                                        className="mr-4"
                                    />
                                    <div>
                                        <h3 className="text-lg font-medium text-gray-900">{provider.name}</h3>
                                        <p className="text-sm text-gray-500">Provider: {provider.provider}</p>
                                        <p className="text-sm text-gray-500">Status: {provider.status}</p>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-4">
                                    <Link
                                        href={`/${providerType}/${provider.provider}/${provider.uuid}`}
                                        className="text-indigo-600 hover:text-indigo-900 font-medium"
                                    >
                                        Edit
                                    </Link>
                                    <button
                                        onClick={() => onDelete(provider.uuid)}
                                        className="text-red-600 hover:text-red-900 font-medium"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </section>
        </div>
    );
};

export default ProviderList;