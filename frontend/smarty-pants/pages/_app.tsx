// pages/_app.tsx
import { AppProps } from 'next/app';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import authService from '@/services/authService';
import LoadingSpinner from '@/components/LoadingSpinner';

const publicRoutes = ['/login', '/register', '/forgot-password', '/auth/google/callback'];

function MyApp({ Component, pageProps }: AppProps) {
    const router = useRouter();
    const [isAuthChecking, setIsAuthChecking] = useState(true);

    useEffect(() => {
        const checkAuth = async () => {
            setIsAuthChecking(true);
            const isAuth = await authService.isAuthenticated();
            if (!isAuth && !publicRoutes.includes(router.pathname)) {
                router.push('/login');
            }
            setIsAuthChecking(false);
        };

        checkAuth();

        const handleRouteChange = async (url: string) => {
            if (!publicRoutes.includes(url) && !(await authService.isAuthenticated())) {
                router.push('/login');
            }
        };

        router.events.on('routeChangeStart', handleRouteChange);

        return () => {
            router.events.off('routeChangeStart', handleRouteChange);
        };
    }, [router]);

    if (isAuthChecking) {
        return <LoadingSpinner />;
    }

    return <Component {...pageProps} />;
}

export default MyApp;