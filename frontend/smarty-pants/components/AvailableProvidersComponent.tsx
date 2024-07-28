// components/AvailableProvidersComponent.tsx
import React from 'react';
import Image from 'next/image';
import Link from 'next/link';

interface AvailableProviderProps {
    id: string;
    name: string;
    description: string;
    imageUrl: string;
    configurationUrl: string;
}

interface AvailableProvidersComponentProps {
    title: string;
    providers: AvailableProviderProps[];
}

const AvailableProvidersComponent: React.FC<AvailableProvidersComponentProps> = ({ title, providers }) => {
    return (
        <div className="w-1/2">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">{title}</h2>
            <div className="bg-white shadow overflow-hidden sm:rounded-lg h-[calc(100vh-240px)] overflow-y-auto">
                {providers.map((provider) => (
                    <div key={provider.id} className="border-b border-gray-200 last:border-b-0">
                        <div className="px-6 py-5 sm:px-6 flex items-center justify-between hover:bg-gray-50 transition-colors duration-200">
                            <div className="flex items-center">
                                <Image
                                    src={provider.imageUrl || '/default-provider-icon.png'}
                                    alt={`${provider.name} icon`}
                                    width={40}
                                    height={40}
                                    className="mr-4"
                                />
                                <div>
                                    <h3 className="text-lg font-medium text-gray-900">{provider.name}</h3>
                                    <p className="text-sm text-gray-500">{provider.description}</p>
                                </div>
                            </div>
                            <Link
                                href={provider.configurationUrl}
                                className="px-3 py-1 bg-blue-100 text-blue-700 rounded-md text-sm font-medium hover:bg-blue-200 transition-colors duration-200"
                            >
                                Configure
                            </Link>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default AvailableProvidersComponent;