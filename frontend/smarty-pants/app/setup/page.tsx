"use client";

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import SVGLogo from "@/components/SVGLogo";
import { SystemAPI } from "@/services/api/system";
import axios from 'axios';

const SetupPage: React.FC = () => {
    const [backendUrl, setBackendUrl] = useState('');
    const [error, setError] = useState('');
    const [isSetupComplete, setIsSetupComplete] = useState(false);
    const router = useRouter();

    useEffect(() => {
        if (isSetupComplete) {
            console.log('Setup complete, redirecting to login...');
            router.push('/login');
        }
    }, [isSetupComplete, router]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        try {
            console.log('Attempting to connect to:', backendUrl);
            const axiosInstance = axios.create({ baseURL: backendUrl });
            const systemApi = new SystemAPI(axiosInstance);
            const systemInfo = await systemApi.getSystemInfo();

            console.log('Connection successful, system info:', systemInfo);
            localStorage.setItem('backendUrl', backendUrl);
            localStorage.setItem('systemInfo', JSON.stringify(systemInfo));

            setIsSetupComplete(true);
        } catch (err) {
            console.error('Failed to connect to the backend:', err);
            setError('Failed to connect to the backend. Please check the URL and try again.');
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8">
                <div className="text-center">
                    <SVGLogo
                        width={80}
                        height={80}
                        leftBrainColor="black"
                        rightBrainColor="black"
                        centerSquareColor="white"
                        centerSquareBlinkColor="blue"
                        onHoverRotationDegree={15}
                    />
                    <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                        Set up SmartyPants
                    </h2>
                </div>
                <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
                    <div>
                        <label htmlFor="backendUrl" className="sr-only">Backend URL</label>
                        <input
                            id="backendUrl"
                            name="backendUrl"
                            type="url"
                            required
                            className="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                            placeholder="Backend URL (e.g., https://api.example.com)"
                            value={backendUrl}
                            onChange={(e) => setBackendUrl(e.target.value)}
                        />
                    </div>
                    {error && <p className="text-red-500 text-sm">{error}</p>}
                    <div>
                        <button
                            type="submit"
                            className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                        >
                            Set Up
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default SetupPage;