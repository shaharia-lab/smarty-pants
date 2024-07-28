/** @type {import('next').NextConfig} */
const nextConfig = {
    images: {
        domains: ['upload.wikimedia.org', 'cdn-icons-png.flaticon.com', 'static-00.iconduck.com', 'huggingface.co'],
    },
    experimental: {
        forceSwcTransforms: false,
    }
};



export default nextConfig;
