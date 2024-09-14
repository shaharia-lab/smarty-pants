/** @type {import('next').NextConfig} */
const nextConfig = {
    images: {
        domains: ['upload.wikimedia.org', 'cdn-icons-png.flaticon.com', 'static-00.iconduck.com', 'huggingface.co'],
    },
    experimental: {
        forceSwcTransforms: false,
    },
    serverRuntimeConfig: {
        API_BASE_URL: process.env.API_BASE_URL,
    },
    publicRuntimeConfig: {
        API_BASE_URL: process.env.API_BASE_URL,
    },
    env: {
        NEXT_PUBLIC_API_BASE_URL: process.env.API_BASE_URL,
    },
};


export default nextConfig;
