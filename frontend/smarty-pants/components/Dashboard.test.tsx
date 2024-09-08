import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import Dashboard from './Dashboard';
import { createApiService } from "@/services/apiService";
import { AnalyticsOverview } from "@/types/api";

// Mock the createApiService function
jest.mock("@/services/apiService", () => ({
    createApiService: jest.fn(),
}));

// Mock the Chart.js library
jest.mock('chart.js', () => ({
    Chart: {
        register: jest.fn(),
    },
    ArcElement: jest.fn(),
    Tooltip: jest.fn(),
    Legend: jest.fn(),
    CategoryScale: jest.fn(),
    LinearScale: jest.fn(),
    BarElement: jest.fn(),
    Title: jest.fn(),
}));

// Mock the react-chartjs-2 library
jest.mock('react-chartjs-2', () => ({
    Pie: () => <div data-testid="mock-pie-chart">Mocked Pie Chart</div>,
}));

describe('Dashboard Component', () => {
    const mockAnalyticsData: AnalyticsOverview = {
        embedding_providers: {
            total_providers: 2,
            active_provider: {
                name: 'OpenAI',
                model: 'text-embedding-ada-002',
                type: 'OpenAI',
            },
            total_active_providers: 1,
        },
        llm_providers: {
            total_providers: 3,
            active_provider: {
                name: 'Example 2',
                model: 'gpt-3.5-turbo',
                type: 'OpenAI',
            },
            total_active_providers: 1,
        },
        datasources: {
            configured_datasources: [
                {
                    name: 'Web Crawler',
                    type: 'web',
                    status: 'active',
                    created_at: '2021-10-01T00:00:00Z',
                },
            ],
            total_datasources: 1,
            total_datasources_by_status: { active: 1 },
            total_datasources_by_type: { web: 1 },
            total_documents_fetched_by_datasource_type: { web: 100 },
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (createApiService as jest.Mock).mockReturnValue({
            analytics: {
                getAnalyticsOverview: jest.fn(),
            },
        });
    });

    test('renders error state when API call fails', async () => {
        const mockError = new Error('API Error');
        (createApiService as jest.Mock)().analytics.getAnalyticsOverview.mockRejectedValueOnce(mockError);

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load analytics data')).toBeInTheDocument();
        });
    });

    test('renders dashboard with correct data after successful API call', async () => {
        (createApiService as jest.Mock)().analytics.getAnalyticsOverview.mockResolvedValueOnce(mockAnalyticsData);

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getByText('Embedding Providers')).toBeInTheDocument();
            expect(screen.getByText('2')).toBeInTheDocument(); // Total Embedding Providers
            expect(screen.getByText('OpenAI')).toBeInTheDocument(); // Embedding Provider name
            expect(screen.getByText('Model: text-embedding-ada-002')).toBeInTheDocument();

            expect(screen.getByText('LLM Providers')).toBeInTheDocument();
            expect(screen.getByText('3')).toBeInTheDocument(); // Total LLM Providers
            expect(screen.getByText('Example 2')).toBeInTheDocument(); // LLM Provider name
            expect(screen.getByText('Model: gpt-3.5-turbo')).toBeInTheDocument();

            expect(screen.getByText('Datasources')).toBeInTheDocument();
            expect(screen.getByText('1')).toBeInTheDocument(); // Total Datasources
            expect(screen.getByText('Web Crawler')).toBeInTheDocument();
            expect(screen.getByText('(web)')).toBeInTheDocument();
            expect(screen.getByText('active')).toBeInTheDocument();

            expect(screen.getByText('Datasources Overview')).toBeInTheDocument();
            expect(screen.getAllByTestId('mock-pie-chart')).toHaveLength(2);
        });
    });

    test('renders "No configured datasources" when datasources array is empty', async () => {
        const modifiedMockData = {
            ...mockAnalyticsData,
            datasources: {
                ...mockAnalyticsData.datasources,
                configured_datasources: [],
            },
        };
        (createApiService as jest.Mock)().analytics.getAnalyticsOverview.mockResolvedValueOnce(modifiedMockData);

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getByText('No configured datasources')).toBeInTheDocument();
        });
    });

    test('renders "No data available" when chart data is empty', async () => {
        const modifiedMockData = {
            ...mockAnalyticsData,
            datasources: {
                ...mockAnalyticsData.datasources,
                total_datasources_by_type: {},
                total_documents_fetched_by_datasource_type: {},
            },
        };
        (createApiService as jest.Mock)().analytics.getAnalyticsOverview.mockResolvedValueOnce(modifiedMockData);

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getAllByText('No data available')).toHaveLength(2);
        });
    });
});