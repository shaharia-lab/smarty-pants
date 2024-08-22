import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

export interface AuthFlowResponse {
    auth_flow: {
        provider: string;
        auth_redirect_url: string;
        state: string;
    };
}

export interface AuthTokenResponse {
    access_token: string;
}

class AuthService {
    private static instance: AuthService;
    private accessToken: string | null = null;

    private constructor() {}

    public static getInstance(): AuthService {
        if (!AuthService.instance) {
            AuthService.instance = new AuthService();
        }
        return AuthService.instance;
    }

    async initiateAuth(provider: string): Promise<AuthFlowResponse> {
        const response = await axios.post(`${API_BASE_URL}/api/v1/auth/initiate`, {
            auth_flow: { provider }
        });
        return response.data;
    }

    async handleCallback(provider: string, authCode: string, state: string): Promise<void> {
        const response = await axios.post(`${API_BASE_URL}/api/v1/auth/callback`, {
            auth_flow: { provider, auth_code: authCode, state }
        });
        this.setAccessToken(response.data.access_token);
    }

    setAccessToken(token: string): void {
        this.accessToken = token;
        localStorage.setItem('access_token', token);
    }

    getAccessToken(): string | null {
        if (!this.accessToken) {
            this.accessToken = localStorage.getItem('access_token');
        }
        return this.accessToken;
    }

    logout(): void {
        this.accessToken = null;
        localStorage.removeItem('access_token');
    }

    isAuthenticated(): boolean {
        return !!this.getAccessToken();
    }

    getAuthenticatedAxiosInstance() {
        const axiosInstance = axios.create({
            baseURL: API_BASE_URL,
        });

        axiosInstance.interceptors.request.use((config) => {
            const token = this.getAccessToken();
            if (token) {
                config.headers['Authorization'] = `Bearer ${token}`;
            }
            return config;
        });

        return axiosInstance;
    }
}

export default AuthService.getInstance();