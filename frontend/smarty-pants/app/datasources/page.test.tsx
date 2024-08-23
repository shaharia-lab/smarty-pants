import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import DatasourcesPage from "@/app/datasources/page";
import { useRouter } from 'next/navigation';
import authService from '@/services/authService';

// Mock Next.js modules
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));

// Mock auth service
jest.mock('@/services/authService', () => ({
    isAuthenticated: jest.fn(),
}));

// Mock the fetch function
global.fetch = jest.fn();

// Mock the environment variable
process.env.NEXT_PUBLIC_API_BASE_URL = 'http://test-api.com';

describe('DatasourcesPage', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue({ push: jest.fn() });
        (authService.isAuthenticated as jest.Mock).mockReturnValue(true);
    });

    it('renders loading state initially', async () => {
        (global.fetch as jest.Mock).mockImplementationOnce(() => new Promise(() => {
        }));

        await act(async () => {
            render(<DatasourcesPage/>);
        });
        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays datasources', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
        });

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('API error'));

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load datasources. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete datasource', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockResolvedValueOnce({
                ok: true,
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: [], total: 0, page: 1, per_page: 10, total_pages: 0}),
            });

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Datasource deleted successfully')).toBeInTheDocument();
        });
    });

    it('handles activate datasource', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'inactive'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({message: 'Datasource activated successfully'}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    datasources: [{...mockDatasources[0], status: 'active'}],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Activate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Datasource activated successfully')).toBeInTheDocument();
        });
    });

    it('handles deactivate datasource', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({message: 'Datasource deactivated successfully'}),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    datasources: [{...mockDatasources[0], status: 'inactive'}],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Deactivate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Datasource deactivated successfully')).toBeInTheDocument();
        });
    });

    // Additional tests to improve coverage
    it('handles API error on delete', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockRejectedValueOnce(new Error('Delete failed'));

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to delete datasource. Please try again.')).toBeInTheDocument();
        });
    });

    it('handles API error on activate', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'inactive'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockRejectedValueOnce(new Error('Activation failed'));

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Activate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Activation failed')).toBeInTheDocument();
        });
    });

    it('handles API error on deactivate', async () => {
        const mockDatasources = [
            {uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active'}
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1}),
            })
            .mockRejectedValueOnce(new Error('Deactivation failed'));

        await act(async () => {
            render(<DatasourcesPage/>);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Deactivate'));
        });

        await waitFor(() => {
            expect(screen.getByText('Deactivation failed')).toBeInTheDocument();
        });
    });
});