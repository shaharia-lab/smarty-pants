"use client";

import React from 'react';
import { useRouter } from 'next/navigation';
import authService from "@/services/authService";

const LoginPage: React.FC = () => {
    const router = useRouter();

    const handleGoogleLogin = async () => {
        try {
            const response = await authService.initiateAuth('google');
            const { auth_redirect_url, state } = response.auth_flow;

            // Store the state in sessionStorage for later verification
            sessionStorage.setItem('auth_state', state);

            // Redirect to Google's login page
            window.location.href = auth_redirect_url;
        } catch (error) {
            console.error('Error initiating Google login:', error);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8">
                <div>
                    <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                        Sign in to your account
                    </h2>
                </div>
                <div>
                    <button
                        onClick={handleGoogleLogin}
                        className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                    >
                        Sign in with Google
                    </button>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;