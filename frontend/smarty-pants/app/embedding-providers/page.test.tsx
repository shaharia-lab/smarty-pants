import React, {act} from 'react';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';
import EmbeddingProvidersPage from "@/app/embedding-providers/page";

// Mock the fetch function
global.fetch = jest.fn();

// Mock the environment variable
process.env.NEXT_PUBLIC_API_BASE_URL = 'http://test-api.com';

describe('EmbeddingProvidersPage', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders loading state initially', async () => {
        // Mock fetch to return a promise that doesn't resolve immediately
        (global.fetch as jest.Mock).mockImplementationOnce(() => new Promise(() => {
        }));

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays embedding providers', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            json: async () => ({embedding_providers: mockProviders, total: 1, page: 1, per_page: 10, total_pages: 1}),
        });

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('API error'));

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load embedding providers. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({embedding_providers: [], total: 0, page: 1, per_page: 10, total_pages: 0}),
            });

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Embedding provider deleted successfully')).toBeInTheDocument();
        });
    });

    it('handles activate provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'inactive'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({message: 'Embedding provider activated successfully'}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: [{...mockProviders[0], status: 'active'}],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Activate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Embedding provider activated successfully')).toBeInTheDocument();
        });
    });

    it('handles deactivate provider', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({message: 'Embedding provider deactivated successfully'}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: [{...mockProviders[0], status: 'inactive'}],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Deactivate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Embedding provider deactivated successfully')).toBeInTheDocument();
        });
    });

    // Additional tests to improve coverage
    it('handles API error on delete', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockRejectedValueOnce(new Error('Delete failed'));

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Provider')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to delete embedding provider. Please try again.')).toBeInTheDocument();
        });
    });

    it('handles API error on activate', async () => {
        const mockProviders = [
            {uuid: '1', name: 'Test Provider', provider: 'test', status: 'inactive'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockRejectedValueOnce(new Error('Activation failed'));

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
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

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    embedding_providers: mockProviders,
                    total: 1,
                    page: 1,
                    per_page: 10,
                    total_pages: 1
                }),
            })
            .mockRejectedValueOnce(new Error('Deactivation failed'));

        await act(async () => {
            render(<EmbeddingProvidersPage/>);
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