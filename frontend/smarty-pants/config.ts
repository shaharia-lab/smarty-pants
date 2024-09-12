import getConfig from 'next/config';

interface RuntimeConfig {
    API_BASE_URL: string;
}

export function getRuntimeConfig(): RuntimeConfig {
    if (typeof window === 'undefined') {
        // Server-side
        return {
            API_BASE_URL: process.env.API_BASE_URL || 'http://localhost:8080',
        };
    }

    // Client-side
    if (process.env.NEXT_PUBLIC_API_BASE_URL) {
        return {
            API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
        };
    }

    // Fallback to runtime config if available
    const runtimeConfig = getConfig()?.publicRuntimeConfig;
    return {
        API_BASE_URL: runtimeConfig?.API_BASE_URL || 'http://localhost:8080',
    };
}