// components/ListComponent.tsx
import React from 'react';
import Image from 'next/image';
import Link from 'next/link';

interface ListItemProps {
    sourceType: string;
    id: string;
    name: string;
    status?: string;
    imageUrl: string;
    onDelete?: (id: string) => void;
    onActivate?: (id: string) => void;
    onDeactivate?: (id: string) => void;
}

interface ListComponentProps {
    title: string;
    items: ListItemProps[];
    loading: boolean;
    error: string | null;
    type: 'llm' | 'embedding' | string; // Add this line
}

const ListComponent: React.FC<ListComponentProps> = ({title, items, loading, error, type}) => {
    const getEditLink = (id: string, sourceType: string) => {
        switch (type) {
            case 'llm':
                return `/llm-providers/${sourceType}/${id}`;
            case 'embedding':
                return `/embedding-providers/${sourceType}/${id}`;
            case 'datasource':
                return `/datasource/${sourceType}/${id}`;
            // Add more cases for other provider types as needed
            default:
                return `/${type}-providers/${id}`;
        }
    };

    return (
        <div className="w-1/2">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">{title}</h2>
            <div className="bg-white shadow overflow-hidden sm:rounded-lg h-[calc(100vh-240px)] overflow-y-auto">
                {loading ? (
                    <p className="p-4">Loading...</p>
                ) : error ? (
                    <p className="p-4 text-red-500">{error}</p>
                ) : items.length > 0 ? (
                    items.map((item) => (
                        <div key={item.id} className="border-b border-gray-200 last:border-b-0">
                            <div
                                className="px-6 py-5 sm:px-6 flex items-center justify-between hover:bg-gray-50 transition-colors duration-200">
                                <div className="flex items-center">
                                    <Image
                                        src={item.imageUrl || '/default-provider-icon.png'}
                                        alt={`${item.name} icon`}
                                        width={40}
                                        height={40}
                                        className="mr-4"
                                    />
                                    <div>
                                        <h3 className="text-lg font-medium text-gray-900">{item.name}</h3>
                                        {item.status && <p className="text-sm text-gray-500">Status: {item.status}</p>}
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Link
                                        href={getEditLink(item.id, item.sourceType)} // Update this line
                                        className="px-3 py-1 bg-indigo-100 text-indigo-700 rounded-md text-sm font-medium hover:bg-indigo-200 transition-colors duration-200"
                                    >
                                        Edit
                                    </Link>
                                    {item.status === 'inactive' && item.onActivate && (
                                        <button
                                            onClick={() => item.onActivate!(item.id)}
                                            className="px-3 py-1 bg-green-100 text-green-700 rounded-md text-sm font-medium hover:bg-green-200 transition-colors duration-200"
                                        >
                                            Activate
                                        </button>
                                    )}
                                    {item.status === 'active' && item.onDeactivate && (
                                        <button
                                            onClick={() => item.onDeactivate!(item.id)}
                                            className="px-3 py-1 bg-gray-100 text-gray-700 rounded-md text-sm font-medium hover:bg-gray-200 transition-colors duration-200"
                                        >
                                            Deactivate
                                        </button>
                                    )}
                                    {item.onDelete && (
                                        <button
                                            onClick={() => item.onDelete!(item.id)}
                                            className="px-3 py-1 bg-red-100 text-red-700 rounded-md text-sm font-medium hover:bg-red-200 transition-colors duration-200"
                                        >
                                            Delete
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="px-6 py-5 text-center text-gray-500">
                        No items to display.
                    </div>
                )}
            </div>
        </div>
    );
};

export default ListComponent;