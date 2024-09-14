import getConfig from 'next/config';

interface RuntimeConfig {
    API_BASE_URL: string;
}

export function getRuntimeConfig(): RuntimeConfig {
    const { serverRuntimeConfig, publicRuntimeConfig } = getConfig();

    if (typeof window === 'undefined') {
        // Server-side
        return {
            API_BASE_URL: serverRuntimeConfig.API_BASE_URL || 'http://localhost:8080',
        };
    }

    // Client-side
    return {
        API_BASE_URL: publicRuntimeConfig.API_BASE_URL || 'http://localhost:8080',
    };
}