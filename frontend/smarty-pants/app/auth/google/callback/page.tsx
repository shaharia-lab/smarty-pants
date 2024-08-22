"use client";

import React, { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import authService from '@/services/authService';

const CallbackPage: React.FC = () => {
    const router = useRouter();
    const searchParams = useSearchParams();

    useEffect(() => {
        const handleCallback = async () => {
            const code = searchParams.get('code');
            const state = searchParams.get('state');

            if (typeof code !== 'string' || typeof state !== 'string') {
                console.error('Invalid code or state');
                router.push('/login');
                return;
            }

            const storedState = sessionStorage.getItem('auth_state');
            if (state !== storedState) {
                console.error('State mismatch');
                router.push('/login');
                return;
            }

            try {
                await authService.handleCallback('google', code, state);
                router.push('/'); // Redirect to home page or dashboard
            } catch (error) {
                console.error('Error handling callback:', error);
                router.push('/login');
            }
        };

        handleCallback();
    }, [router, searchParams]);

    return <div>Processing login...</div>;
};

export default CallbackPage;