// File: /components/SetupGuide.tsx

'use client';

import React, {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';
import Link from 'next/link';

const SetupGuide: React.FC = () => {
    const [hasDatasource, setHasDatasource] = useState<boolean | null>(null);
    const [hasEmbeddingProvider, setHasEmbeddingProvider] = useState<boolean | null>(null);
    const router = useRouter();

    useEffect(() => {
        const checkConfiguration = async () => {
            try {
                const datasourcesResponse = await fetch('http://localhost:8080/api/v1/datasource?limit=1');
                const datasourcesData = await datasourcesResponse.json();
                setHasDatasource(datasourcesData.total > 0);

                const embeddingProvidersResponse = await fetch('http://localhost:8080/api/v1/embedding-provider?limit=1');
                const embeddingProvidersData = await embeddingProvidersResponse.json();
                setHasEmbeddingProvider(embeddingProvidersData.total > 0);
            } catch (error) {
                console.error('Error checking configuration:', error);
            }
        };

        checkConfiguration();
    }, []);

    if (hasDatasource === null || hasEmbeddingProvider === null) {
        return <div>Loading...</div>;
    }

    if (!hasDatasource) {
        return (
            <div className="text-center">
                <h2 className="text-2xl font-semibold mb-4">Welcome to SmartyPants AI!</h2>
                <p className="mb-4">To get started, you need to configure a datasource.</p>
                <Link href="/datasources"
                      className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                    Configure Datasource
                </Link>
            </div>
        );
    }

    if (!hasEmbeddingProvider) {
        return (
            <div className="text-center">
                <h2 className="text-2xl font-semibold mb-4">Almost there!</h2>
                <p className="mb-4">Now, let's configure an embedding provider.</p>
                <Link href="/embedding-providers"
                      className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                    Configure Embedding Provider
                </Link>
            </div>
        );
    }

    router.push('/dashboard');
    return null;
};

export default SetupGuide;