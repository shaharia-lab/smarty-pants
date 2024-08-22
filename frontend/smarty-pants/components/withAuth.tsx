import { useEffect } from 'react';
import { useRouter } from 'next/router';
import authService from '../services/authService';

const withAuth = <P extends object>(WrappedComponent: React.ComponentType<P>) => {
    const AuthWrapper: React.FC<P> = (props) => {
        const router = useRouter();

        useEffect(() => {
            if (!authService.isAuthenticated()) {
                router.push('/login');
            }
        }, []);

        if (!authService.isAuthenticated()) {
            return null; // or a loading spinner
        }

        return <WrappedComponent {...props} />;
    };

    return AuthWrapper;
};

export default withAuth;