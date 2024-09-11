import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import OpenAIEmbeddingProviderForm from './OpenAIEmbeddingProviderForm';
import { createApiService } from "@/services/apiService";
import { useRouter } from 'next/navigation';
import {HeaderConfig} from "@/components/Header";

// Mock the necessary dependencies
jest.mock('@/services/apiService');
jest.mock('@/services/authService');
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
}));
jest.mock('./Navbar', () => () => <div data-testid="navbar">Navbar</div>);
jest.mock('./Header', () => ({ config }: { config: HeaderConfig }) => <h1>{config.title}</h1>);


describe('OpenAIEmbeddingProviderForm', () => {
    const mockRouter = {
        push: jest.fn(),
    };
    const mockEmbeddingProviderApi = {
        getEmbeddingProvider: jest.fn(),
        createEmbeddingProvider: jest.fn(),
        updateEmbeddingProvider: jest.fn(),
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue(mockRouter);
        (createApiService as jest.Mock).mockReturnValue({ embeddingProvider: mockEmbeddingProviderApi });
    });

    it('renders the form in add mode', () => {
        render(<OpenAIEmbeddingProviderForm />);

        expect(screen.getByText('Configure OpenAI Embedding Provider')).toBeInTheDocument();
        expect(screen.getByLabelText('Name')).toBeInTheDocument();
        expect(screen.getByLabelText('API Key')).toBeInTheDocument();
        expect(screen.getByLabelText('Model ID')).toBeInTheDocument();
        expect(screen.getByText('Save Provider')).toBeInTheDocument();
        expect(screen.getByText('Validate')).toBeInTheDocument();
    });

    it('renders the form in edit mode', async () => {
        const mockProvider = {
            name: 'Test Provider',
            configuration: {
                api_key: 'test-api-key',
                model_id: 'text-embedding-3-small',
            },
        };
        mockEmbeddingProviderApi.getEmbeddingProvider.mockResolvedValue(mockProvider);

        render(<OpenAIEmbeddingProviderForm providerId="test-id" />);

        await waitFor(() => {
            // Check if there are two headings with the expected text
            const headings = screen.getAllByRole('heading', { level: 1 });
            const editHeadings = headings.filter(heading =>
                heading.textContent?.includes('Edit OpenAI Embedding Provider')
            );
            expect(editHeadings.length).toBe(2);

            // Check for the form fields and buttons
            expect(screen.getByLabelText('Name')).toHaveValue('Test Provider');
            expect(screen.getByLabelText('API Key')).toHaveValue('test-api-key');
            expect(screen.getByLabelText('Model ID')).toHaveValue('text-embedding-3-small');
            expect(screen.getByRole('button', { name: 'Update Provider' })).toBeInTheDocument();
            expect(screen.queryByRole('button', { name: 'Validate' })).not.toBeInTheDocument();
        });
    });

    it('handles form submission in add mode', async () => {
        render(<OpenAIEmbeddingProviderForm />);

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'New Provider' } });
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'new-api-key' } });
        fireEvent.change(screen.getByLabelText('Model ID'), { target: { value: 'text-embedding-3-large' } });

        fireEvent.click(screen.getByText('Save Provider'));

        await waitFor(() => {
            expect(mockEmbeddingProviderApi.createEmbeddingProvider).toHaveBeenCalledWith({
                name: 'New Provider',
                provider: 'openai',
                configuration: {
                    api_key: 'new-api-key',
                    model_id: 'text-embedding-3-large',
                    encoding_format: 'float',
                    dimensions: 1536,
                },
                status: 'active',
            });
            expect(mockRouter.push).toHaveBeenCalledWith('/embedding-providers');
        });
    });

    it('handles form submission in edit mode', async () => {
        mockEmbeddingProviderApi.getEmbeddingProvider.mockResolvedValue({
            name: 'Test Provider',
            configuration: {
                api_key: 'test-api-key',
                model_id: 'text-embedding-ada-002',
            },
        });

        render(<OpenAIEmbeddingProviderForm providerId="test-id" />);

        await waitFor(() => {
            fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Updated Provider' } });
            fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'updated-api-key' } });
            fireEvent.change(screen.getByLabelText('Model ID'), { target: { value: 'text-embedding-3-small' } });
        });

        fireEvent.click(screen.getByText('Update Provider'));

        await waitFor(() => {
            expect(mockEmbeddingProviderApi.updateEmbeddingProvider).toHaveBeenCalledWith('test-id', {
                name: 'Updated Provider',
                provider: 'openai',
                configuration: {
                    api_key: 'updated-api-key',
                    model_id: 'text-embedding-3-small',
                    encoding_format: 'float',
                    dimensions: 1536,
                },
                status: 'active',
            });
            expect(mockRouter.push).toHaveBeenCalledWith('/embedding-providers');
        });
    });

    it('handles validation button click', async () => {
        render(<OpenAIEmbeddingProviderForm />);

        const validateButton = screen.getByText('Validate');
        fireEvent.click(validateButton);

        await waitFor(() => {
            expect(validateButton).toBeDisabled();
            expect(validateButton).toHaveTextContent('Validated');
        });
    });

    it('displays error message on API failure', async () => {
        mockEmbeddingProviderApi.createEmbeddingProvider.mockRejectedValue(new Error('API Error'));

        render(<OpenAIEmbeddingProviderForm />);

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'New Provider' } });
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'new-api-key' } });
        fireEvent.click(screen.getByText('Save Provider'));

        await waitFor(() => {
            expect(screen.getByText('Failed to add embedding provider. Please try again.')).toBeInTheDocument();
        });
    });
});