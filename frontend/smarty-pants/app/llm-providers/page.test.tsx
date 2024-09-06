import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import LLMProvidersPage from "@/app/llm-providers/page";
import { useRouter } from "next/navigation";
import AuthService from "@/services/authService";
import { createApiService } from "@/services/apiService";

jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));

jest.mock('@/services/authService', () => ({
    isAuthenticated: jest.fn(),
}));

jest.mock('@/services/apiService', () => ({
    createApiService: jest.fn(),
}));

const mockLLMProviderApi = {
    getLLMProviders: jest.fn(),
    deleteLLMProvider: jest.fn(),
    activateLLMProvider: jest.fn(),
    deactivateLLMProvider: jest.fn(),
};

describe('LLMProvidersPage', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue({ push: jest.fn() });
        (AuthService.isAuthenticated as jest.Mock).mockReturnValue(true);
        (createApiService as jest.Mock).mockReturnValue({ llmProvider: mockLLMProviderApi });
    });

    it('renders loading state initially', async () => {
        mockLLMProviderApi.getLLMProviders.mockReturnValue(new Promise(() => {}));

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays LLM providers', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test LLM Provider', provider: 'test', status: 'active'}
        ];

        mockLLMProviderApi.getLLMProviders.mockResolvedValue({
            llm_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test LLM Provider')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        mockLLMProviderApi.getLLMProviders.mockRejectedValue(new Error('API error'));

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load LLM providers. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test LLM Provider', provider: 'test', status: 'active'}
        ];

        mockLLMProviderApi.getLLMProviders
            .mockResolvedValueOnce({
                llm_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                llm_providers: [],
                total: 0,
                page: 1,
                per_page: 10,
                total_pages: 0
            });

        mockLLMProviderApi.deleteLLMProvider.mockResolvedValue(undefined);

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test LLM Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('LLM provider deleted successfully')).toBeInTheDocument();
        });
    });

    it('handles activate provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test LLM Provider', provider: 'test', status: 'inactive'}
        ];

        mockLLMProviderApi.getLLMProviders
            .mockResolvedValueOnce({
                llm_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                llm_providers: [{...mockProviders[0], status: 'active'}],
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            });

        mockLLMProviderApi.activateLLMProvider.mockResolvedValue({ message: 'LLM provider activated successfully' });

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test LLM Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Activate'));
        });

        await waitFor(() => {
            expect(screen.getByText('LLM provider activated successfully')).toBeInTheDocument();
        });
    });

    it('handles deactivate provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test LLM Provider', provider: 'test', status: 'active'}
        ];

        mockLLMProviderApi.getLLMProviders
            .mockResolvedValueOnce({
                llm_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                llm_providers: [{...mockProviders[0], status: 'inactive'}],
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            });

        mockLLMProviderApi.deactivateLLMProvider.mockResolvedValue({ message: 'LLM provider deactivated successfully' });

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test LLM Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Deactivate'));
        });

        await waitFor(() => {
            expect(screen.getByText('LLM provider deactivated successfully')).toBeInTheDocument();
        });
    });

    it('handles API error on delete', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test LLM Provider', provider: 'test', status: 'active'}
        ];

        mockLLMProviderApi.getLLMProviders.mockResolvedValue({
            llm_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockLLMProviderApi.deleteLLMProvider.mockRejectedValue(new Error('Delete failed'));

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test LLM Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to delete LLM provider. Please try again.')).toBeInTheDocument();
        });
    });

    it('handles API error on activate', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'inactive'}
        ];

        mockLLMProviderApi.getLLMProviders.mockResolvedValue({
            llm_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockLLMProviderApi.activateLLMProvider.mockRejectedValue(new Error('Activation failed'));

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Activate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Activation failed')).toBeInTheDocument();
        });
    });

    it('handles API error on deactivate', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'active'}
        ];

        mockLLMProviderApi.getLLMProviders.mockResolvedValue({
            llm_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockLLMProviderApi.deactivateLLMProvider.mockRejectedValue(new Error('Deactivation failed'));

        await act(async () => {
            render(<LLMProvidersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Deactivate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Deactivation failed')).toBeInTheDocument();
        });
    });
});