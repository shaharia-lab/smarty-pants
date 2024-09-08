import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import EmbeddingProvidersPage from "@/app/embedding-providers/page";
import { useRouter } from 'next/navigation';
import AuthService from '@/services/authService';
import { createApiService } from "@/services/apiService";

// Mock Next.js modules
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));

// Mock auth service
jest.mock('@/services/authService', () => ({
    isAuthenticated: jest.fn(),
    getAuthenticatedAxiosInstance: jest.fn(),
}));

// Mock createApiService
jest.mock("@/services/apiService", () => ({
    createApiService: jest.fn(),
}));

describe('EmbeddingProvidersPage', () => {
    const mockRouter = { push: jest.fn() };
    const mockApiService = {
        embeddingProvider: {
            getEmbeddingProviders: jest.fn(),
            deleteEmbeddingProvider: jest.fn(),
            activateEmbeddingProvider: jest.fn(),
            deactivateEmbeddingProvider: jest.fn(),
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue(mockRouter);
        (AuthService.isAuthenticated as jest.Mock).mockReturnValue(true);
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
    });

    it('renders loading state initially', async () => {
        mockApiService.embeddingProvider.getEmbeddingProviders.mockImplementation(() => new Promise(() => {}));

        render(<EmbeddingProvidersPage />);

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays embedding providers', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'active' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders.mockResolvedValue({
            embedding_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        mockApiService.embeddingProvider.getEmbeddingProviders.mockRejectedValue(new Error('API error'));

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Failed to load embedding providers. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete provider', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'active' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders
            .mockResolvedValueOnce({
                embedding_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                embedding_providers: [],
                total: 0,
                page: 1,
                per_page: 10,
                total_pages: 0
            });

        mockApiService.embeddingProvider.deleteEmbeddingProvider.mockResolvedValue(undefined);

        window.confirm = jest.fn().mockImplementation(() => true);

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Delete'));

        await waitFor(() => {
            expect(screen.getByText('Embedding provider deleted successfully')).toBeInTheDocument();
        });
    });

    it('handles activate provider', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'inactive' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders
            .mockResolvedValueOnce({
                embedding_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                embedding_providers: [{ ...mockProviders[0], status: 'active' }],
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            });

        mockApiService.embeddingProvider.activateEmbeddingProvider.mockResolvedValue({ message: 'Embedding provider activated successfully' });

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Activate'));

        await waitFor(() => {
            expect(screen.getByText('Embedding provider activated successfully')).toBeInTheDocument();
        });
    });

    it('handles deactivate provider', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'active' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders
            .mockResolvedValueOnce({
                embedding_providers: mockProviders,
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            })
            .mockResolvedValueOnce({
                embedding_providers: [{ ...mockProviders[0], status: 'inactive' }],
                total: 1,
                page: 1,
                per_page: 10,
                total_pages: 1
            });

        mockApiService.embeddingProvider.deactivateEmbeddingProvider.mockResolvedValue({ message: 'Embedding provider deactivated successfully' });

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Deactivate'));

        await waitFor(() => {
            expect(screen.getByText('Embedding provider deactivated successfully')).toBeInTheDocument();
        });
    });

    // Additional tests for error handling
    it('handles API error on delete', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'active' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders.mockResolvedValue({
            embedding_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockApiService.embeddingProvider.deleteEmbeddingProvider.mockRejectedValue(new Error('Delete failed'));

        window.confirm = jest.fn().mockImplementation(() => true);

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Delete'));

        await waitFor(() => {
            // Look for a more general error message
            expect(screen.getByText(/failed/i)).toBeInTheDocument();
        });
    });

    it('handles API error on activate', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'inactive' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders.mockResolvedValue({
            embedding_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockApiService.embeddingProvider.activateEmbeddingProvider.mockRejectedValue(new Error('Activation failed'));

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Activate'));

        await waitFor(() => {
            expect(screen.getByText('Activation failed')).toBeInTheDocument();
        });
    });

    it('handles API error on deactivate', async () => {
        const mockProviders = [
            { uuid: '1', name: 'Test Provider', provider: 'test', status: 'active' }
        ];

        mockApiService.embeddingProvider.getEmbeddingProviders.mockResolvedValue({
            embedding_providers: mockProviders,
            total: 1,
            page: 1,
            per_page: 10,
            total_pages: 1
        });

        mockApiService.embeddingProvider.deactivateEmbeddingProvider.mockRejectedValue(new Error('Deactivation failed'));

        render(<EmbeddingProvidersPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Deactivate'));

        await waitFor(() => {
            expect(screen.getByText('Deactivation failed')).toBeInTheDocument();
        });
    });
});