import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { act } from 'react-dom/test-utils';
import DatasourcesPage from "@/app/datasources/page";

// Mock the fetch function
global.fetch = jest.fn();

// Mock the environment variable
process.env.NEXT_PUBLIC_API_BASE_URL = 'http://test-api.com';

describe('DatasourcesPage', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders loading state initially', () => {
        render(<DatasourcesPage />);
        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays datasources', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            json: async () => ({ datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1 }),
        });

        render(<DatasourcesPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('API error'));

        render(<DatasourcesPage />);

        await waitFor(() => {
            expect(screen.getByText('Failed to load datasources. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete datasource', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1 }),
            })
            .mockResolvedValueOnce({
                ok: true,
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ datasources: [], total: 0, page: 1, per_page: 10, total_pages: 0 }),
            });

        window.confirm = jest.fn().mockImplementation(() => true);

        render(<DatasourcesPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Delete'));

        await waitFor(() => {
            expect(screen.getByText('Datasource deleted successfully')).toBeInTheDocument();
        });
    });

    it('handles activate datasource', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'inactive' }
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1 }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ message: 'Datasource activated successfully' }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    datasources: [{ ...mockDatasources[0], status: 'active' }],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        render(<DatasourcesPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Activate'));

        await waitFor(() => {
            expect(screen.getByText('Datasource activated successfully')).toBeInTheDocument();
        });
    });

    it('handles deactivate datasource', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ datasources: mockDatasources, total: 1, page: 1, per_page: 10, total_pages: 1 }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({ message: 'Datasource deactivated successfully' }),
            })
            .mockResolvedValueOnce({
                ok: true,
                json: async () => ({
                    datasources: [{ ...mockDatasources[0], status: 'inactive' }],
                    total: 1, page: 1, per_page: 10, total_pages: 1
                }),
            });

        render(<DatasourcesPage />);

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Deactivate'));

        await waitFor(() => {
            expect(screen.getByText('Datasource deactivated successfully')).toBeInTheDocument();
        });
    });
});