// File: app/auth/google/callback/page.tsx

"use client";

import React, { useEffect, useRef, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import authService from '@/services/authService';

const CallbackPageContent: React.FC = () => {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const callbackExecuted = useRef(false);

    useEffect(() => {
        const handleCallback = async () => {
            if (callbackExecuted.current) return;
            callbackExecuted.current = true;

            const code = (searchParams as URLSearchParams).get('code');
            const state = (searchParams as URLSearchParams).get('state');

            if (typeof code !== 'string' || typeof state !== 'string') {
                setError('Invalid code or state');
                setIsLoading(false);
                return;
            }

            const storedState = sessionStorage.getItem('auth_state');
            if (state !== storedState) {
                setError('State mismatch');
                setIsLoading(false);
                return;
            }

            try {
                await authService.handleCallback('google', code, state);
                sessionStorage.removeItem('auth_state'); // Clean up the stored state
                router.push('/'); // Redirect to home page or dashboard
            } catch (error) {
                console.error('Error handling callback:', error);
                setError('Authentication failed. Please try again.');
                setIsLoading(false);
            }
        };

        handleCallback();
    }, [router, searchParams]);

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
                <div className="bg-white p-8 rounded-lg shadow-md">
                    <h1 className="text-2xl font-bold mb-4 text-red-600">Authentication Error</h1>
                    <p className="text-gray-700 mb-4">{error}</p>
                    <button
                        onClick={() => router.push('/login')}
                        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
                    >
                        Return to Login
                    </button>
                </div>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen bg-gray-100">
                <div className="bg-white p-8 rounded-lg shadow-md">
                    <h1 className="text-2xl font-bold mb-4">Processing login...</h1>
                    <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
                </div>
            </div>
        );
    }

    return null;
};

const CallbackPage: React.FC = () => {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen bg-gray-100">Loading...</div>}>
            <CallbackPageContent />
        </Suspense>
    );
};

export default CallbackPage;
