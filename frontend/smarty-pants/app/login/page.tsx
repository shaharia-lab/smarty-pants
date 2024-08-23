"use client";

import React from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import authService from "@/services/authService";
import SVGLogo from "@/components/SVGLogo";

const LoginPage: React.FC = () => {
    const router = useRouter();

    const handleLogin = async (provider: string) => {
        try {
            const response = await authService.initiateAuth(provider);
            const { auth_redirect_url, state } = response.auth_flow;

            // Store the state in sessionStorage for later verification
            sessionStorage.setItem('auth_state', state);

            // Redirect to the provider's login page
            window.location.href = auth_redirect_url;
        } catch (error) {
            console.error(`Error initiating ${provider} login:`, error);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8">
                <div className="text-center">
                    <Link href="/" className="flex flex-col items-center">
                        <SVGLogo
                            width={80}
                            height={80}
                            leftBrainColor="black"
                            rightBrainColor="black"
                            centerSquareColor="white"
                            centerSquareBlinkColor="blue"
                            onHoverRotationDegree={15}
                        />
                        <span className="mt-2 text-gray-900 font-bold text-2xl">SmartyPants</span>
                    </Link>
                    <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                        Sign in to your account
                    </h2>
                </div>
                <div className="bg-white shadow-md rounded-lg p-8 space-y-4">
                    <button
                        onClick={() => handleLogin('google')}
                        className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition ease-in-out duration-150"
                    >
                        Sign in with Google
                    </button>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;
