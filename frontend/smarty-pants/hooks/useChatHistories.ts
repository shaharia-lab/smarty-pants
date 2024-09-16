// hooks/useChatHistories.ts
import { useState, useEffect } from 'react';
import { createApiService } from "@/services/apiService";
import AuthService from "@/services/authService";
import axios from "axios";
import { Interaction } from '@/types/api';

export const useChatHistories = () => {
    const [histories, setHistories] = useState<Interaction[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const apiService = createApiService(AuthService);
        const source = axios.CancelToken.source();

        const fetchHistories = async () => {
            setIsLoading(true);
            try {
                const data = await apiService.chatHisories.getChatHistories(source.token);
                setHistories(data.interactions);
            } catch (error) {
                console.error('Error fetching chat histories:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchHistories();

        return () => {
            source.cancel('Component unmounted');
        };
    }, []);

    return { histories, isLoading };
};