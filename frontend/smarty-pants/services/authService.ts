import axios, { AxiosInstance } from 'axios';
import Cookies from "js-cookie";

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
    refresh_token: string;
    expires_in: number;
}

class AuthService {
    private static instance: AuthService;
    private accessToken: string | null = null;
    private refreshToken: string | null = null;
    private expirationTime: number | null = null;

    private constructor() {
        this.loadTokens();
    }

    public static getInstance(): AuthService {
        if (!AuthService.instance) {
            AuthService.instance = new AuthService();
        }
        return AuthService.instance;
    }

    private loadTokens(): void {
        if (typeof window !== 'undefined') {
            this.accessToken = localStorage.getItem('access_token');
            this.refreshToken = localStorage.getItem('refresh_token');
            const expiration = localStorage.getItem('token_expiration');
            this.expirationTime = expiration ? parseInt(expiration, 10) : null;
        }
    }

    async initiateAuth(provider: string): Promise<AuthFlowResponse> {
        const response = await axios.post(`${API_BASE_URL}/api/v1/auth/initiate`, {
            auth_flow: { provider }
        });
        return response.data;
    }

    async verifyToken(token: string): Promise<boolean> {
        try {
            // First, check if the token matches the stored access token
            if (token === this.accessToken) {
                // If it matches, check if it's expired
                if (this.isTokenExpired()) {
                    // If expired, try to refresh
                    await this.refreshAccessToken();
                    return !!this.accessToken;
                }
                return true;
            }

            // If it doesn't match the stored token, verify with the backend
            const response = await axios.post(`${API_BASE_URL}/api/v1/auth/verify`, { token });
            return response.data.isValid;
        } catch (error) {
            console.error('Error verifying token:', error);
            return false;
        }
    }

    async handleCallback(provider: string, authCode: string, state: string): Promise<void> {
        const response = await axios.post(`${API_BASE_URL}/api/v1/auth/callback`, {
            auth_flow: { provider, auth_code: authCode, state }
        });
        this.setTokens(response.data);
    }

    setTokens(tokenResponse: AuthTokenResponse): void {
        this.accessToken = tokenResponse.access_token;
        this.refreshToken = tokenResponse.refresh_token;
        this.expirationTime = Date.now() + tokenResponse.expires_in * 1000;

        Cookies.set('auth_token', this.accessToken, { expires: 7 }); // Set cookie to expire in 7 days
        localStorage.setItem('refresh_token', this.refreshToken);
        localStorage.setItem('token_expiration', this.expirationTime.toString());
    }

    getAccessToken(): string | null {
        if (!this.accessToken) {
            this.accessToken = Cookies.get('auth_token') || null;
        }
        return this.accessToken;
    }

    private isTokenExpired(): boolean {
        return this.expirationTime ? Date.now() > this.expirationTime : true;
    }

    private async refreshAccessToken(): Promise<void> {
        /**
         * @TODO: Implement token refresh
         */
        /*try {
            const response = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
                refresh_token: this.refreshToken
            });
            this.setTokens(response.data);
        } catch (error) {
            console.error('Error refreshing token:', error);
            this.logout();
        }*/
    }

    logout(): void {
        this.accessToken = null;
        this.refreshToken = null;
        this.expirationTime = null;
        Cookies.remove('auth_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('token_expiration');
    }

    isAuthenticated(): boolean {
        return !!this.getAccessToken();
    }

    getAuthenticatedAxiosInstance(): AxiosInstance {
        const axiosInstance = axios.create({
            baseURL: API_BASE_URL,
        });

        axiosInstance.interceptors.request.use(async (config) => {
            const token = await this.getAccessToken();
            if (token) {
                config.headers['Authorization'] = `Bearer ${token}`;
            }
            return config;
        });

        axiosInstance.interceptors.response.use(
            (response) => response,
            async (error) => {
                if (error.response && error.response.status === 401 && this.refreshToken) {
                    await this.refreshAccessToken();
                    const token = this.getAccessToken();
                    if (token) {
                        error.config.headers['Authorization'] = `Bearer ${token}`;
                        return axios(error.config);
                    }
                }
                return Promise.reject(error);
            }
        );

        return axiosInstance;
    }
}

export default AuthService.getInstance();

export interface IAuthService {
    getAuthenticatedAxiosInstance(): AxiosInstance;
}