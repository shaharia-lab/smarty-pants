// File: /components/AvailableItems.tsx

import React from 'react';
import Link from 'next/link';
import Image from 'next/image';

interface AvailableItem {
    id: string;
    name: string;
    imageUrl: string;
    description: string;
    configurationUrl: string;
}

interface AvailableItemsProps {
    items: AvailableItem[];
    title: string;
}

const AvailableItems: React.FC<AvailableItemsProps> = ({ items, title }) => {
    return (
        <section className="mb-12">
            <h2 className="text-2xl font-semibold text-gray-900 mb-6">{title}</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {items.map((item) => (
                    <Link href={item.configurationUrl} key={item.id} className="block">
                        <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300 overflow-hidden">
                            <div className="p-6">
                                <div className="flex items-center mb-4">
                                    <Image
                                        src={item.imageUrl}
                                        alt={`${item.name} icon`}
                                        width={48}
                                        height={48}
                                        className="rounded-full"
                                    />
                                    <h3 className="ml-4 text-xl font-semibold text-gray-900">{item.name}</h3>
                                </div>
                                <p className="text-gray-600 mb-4">{item.description}</p>
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
    );
};

export default AvailableItems;
