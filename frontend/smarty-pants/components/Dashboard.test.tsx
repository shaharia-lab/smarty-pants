import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import Dashboard from './Dashboard';

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

// Mock the global fetch function with proper typing
const mockFetch = jest.fn() as jest.MockedFunction<typeof fetch>;
global.fetch = mockFetch;

// Helper function to create a mock Response
const createMockResponse = (data: any): Response => {
    return {
        json: jest.fn().mockResolvedValue(data),
        ok: true,
        status: 200,
        statusText: 'OK',
        headers: new Headers(),
        redirected: false,
        type: 'basic',
        url: 'http://test.com',
        clone: jest.fn(),
        body: null,
        bodyUsed: false,
        arrayBuffer: jest.fn(),
        blob: jest.fn(),
        formData: jest.fn(),
        text: jest.fn(),
    } as Response;
};

describe('Dashboard Component', () => {
    const mockAnalyticsData = {
        embedding_providers: {
            total_providers: 2,
            total_active_providers: 1,
            active_provider: {
                name: 'OpenAI',
                type: 'API',
                model: 'text-embedding-ada-002',
            },
        },
        llm_providers: {
            total_providers: 3,
            total_active_providers: 1,
            active_provider: {
                name: 'OpenAI',
                type: 'API',
                model: 'gpt-3.5-turbo',
            },
        },
        datasources: {
            configured_datasources: [
                {
                    name: 'Web Crawler',
                    type: 'web',
                    status: 'active',
                    created_at: '2023-01-01T00:00:00Z',
                },
            ],
            total_datasources: 1,
            total_datasources_by_type: { web: 1 },
            total_datasources_by_status: { active: 1 },
            total_documents_fetched_by_datasource_type: { web: 100 },
        },
    };

    beforeEach(() => {
        jest.resetAllMocks();
    });

    test('renders error state when API call fails', async () => {
        mockFetch.mockRejectedValueOnce(new Error('API Error'));

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load analytics data')).toBeInTheDocument();
        });
    });

    test('renders dashboard with correct data after successful API call', async () => {
        mockFetch.mockResolvedValueOnce(createMockResponse(mockAnalyticsData));

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getByText('Embedding Providers')).toBeInTheDocument();
            expect(screen.getByText('2')).toBeInTheDocument(); // Total Embedding Providers

            // Check for OpenAI in Embedding Providers section
            const embeddingProvidersSection = screen.getByText('Embedding Providers').closest('div');
            expect(embeddingProvidersSection).toBeInTheDocument();
            expect(embeddingProvidersSection).toHaveTextContent('OpenAI');
            expect(embeddingProvidersSection).toHaveTextContent('Model: text-embedding-ada-002');

            expect(screen.getByText('LLM Providers')).toBeInTheDocument();
            expect(screen.getByText('3')).toBeInTheDocument(); // Total LLM Providers

            // Check for OpenAI in LLM Providers section
            const llmProvidersSection = screen.getByText('LLM Providers').closest('div');
            expect(llmProvidersSection).toBeInTheDocument();
            expect(llmProvidersSection).toHaveTextContent('OpenAI');
            expect(llmProvidersSection).toHaveTextContent('Model: gpt-3.5-turbo');

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
        mockFetch.mockResolvedValueOnce(createMockResponse(modifiedMockData));

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
        mockFetch.mockResolvedValueOnce(createMockResponse(modifiedMockData));

        await act(async () => {
            render(<Dashboard />);
        });

        await waitFor(() => {
            expect(screen.getAllByText('No data available')).toHaveLength(2);
        });
    });
});