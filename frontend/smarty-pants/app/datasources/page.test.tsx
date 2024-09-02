import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import DatasourcesPage from "@/app/datasources/page";
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

describe('DatasourcesPage', () => {
    const mockRouter = { push: jest.fn() };
    const mockApiService = {
        datasource: {
            getDatasources: jest.fn(),
            deleteDatasource: jest.fn(),
            activateDatasource: jest.fn(),
            deactivateDatasource: jest.fn(),
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue(mockRouter);
        (AuthService.isAuthenticated as jest.Mock).mockReturnValue(true);
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
    });

    it('renders loading state initially', async () => {
        mockApiService.datasource.getDatasources.mockReturnValue(new Promise(() => {}));

        await act(async () => {
            render(<DatasourcesPage />);
        });

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('fetches and displays datasources', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        mockApiService.datasource.getDatasources.mockResolvedValue({ datasources: mockDatasources });

        await act(async () => {
            render(<DatasourcesPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });
    });

    it('handles fetch error', async () => {
        mockApiService.datasource.getDatasources.mockRejectedValue(new Error('API error'));

        await act(async () => {
            render(<DatasourcesPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load datasources. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles delete datasource', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        mockApiService.datasource.getDatasources
            .mockResolvedValueOnce({ datasources: mockDatasources })
            .mockResolvedValueOnce({ datasources: [] });

        mockApiService.datasource.deleteDatasource.mockResolvedValue({});

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<DatasourcesPage />);
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
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'inactive' }
        ];

        mockApiService.datasource.getDatasources
            .mockResolvedValueOnce({ datasources: mockDatasources })
            .mockResolvedValueOnce({ datasources: [{ ...mockDatasources[0], status: 'active' }] });

        mockApiService.datasource.activateDatasource.mockResolvedValue({ message: 'Datasource activated successfully' });

        await act(async () => {
            render(<DatasourcesPage />);
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
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        mockApiService.datasource.getDatasources
            .mockResolvedValueOnce({ datasources: mockDatasources })
            .mockResolvedValueOnce({ datasources: [{ ...mockDatasources[0], status: 'inactive' }] });

        mockApiService.datasource.deactivateDatasource.mockResolvedValue({ message: 'Datasource deactivated successfully' });

        await act(async () => {
            render(<DatasourcesPage />);
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
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        mockApiService.datasource.getDatasources.mockResolvedValue({ datasources: mockDatasources });
        mockApiService.datasource.deleteDatasource.mockRejectedValue(new Error('Delete failed'));

        window.confirm = jest.fn().mockImplementation(() => true);

        await act(async () => {
            render(<DatasourcesPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Test Datasource')).toBeInTheDocument();
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Delete'));
        });

        await waitFor(() => {
            expect(screen.getByText('Delete failed')).toBeInTheDocument();
        });
    });

    it('handles API error on activate', async () => {
        const mockDatasources = [
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'inactive' }
        ];

        mockApiService.datasource.getDatasources.mockResolvedValue({ datasources: mockDatasources });
        mockApiService.datasource.activateDatasource.mockRejectedValue(new Error('Activation failed'));

        await act(async () => {
            render(<DatasourcesPage />);
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
            { uuid: '1', name: 'Test Datasource', source_type: 'test', status: 'active' }
        ];

        mockApiService.datasource.getDatasources.mockResolvedValue({ datasources: mockDatasources });
        mockApiService.datasource.deactivateDatasource.mockRejectedValue(new Error('Deactivation failed'));

        await act(async () => {
            render(<DatasourcesPage />);
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