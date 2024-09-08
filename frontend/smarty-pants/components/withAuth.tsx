// components/withAuth.tsx
import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import authService from '@/services/authService';
import LoadingSpinner from '@/components/LoadingSpinner';

const withAuth = <P extends object>(WrappedComponent: React.ComponentType<P>) => {
    const AuthWrapper: React.FC<P> = (props) => {
        const router = useRouter();
        const [isAuthenticated, setIsAuthenticated] = useState(false);

        useEffect(() => {
            const checkAuth = async () => {
                const auth = await authService.isAuthenticated();
                if (!auth) {
                    router.push('/login');
                } else {
                    setIsAuthenticated(true);
                }
            };

            checkAuth();
        }, [router]);

        if (!isAuthenticated) {
            return <LoadingSpinner />;
        }

        return <WrappedComponent {...props} />;
    };

    return AuthWrapper;
};

export default withAuth;