import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import OpenAILLMProviderForm from './OpenAILLMProviderForm';
import { createApiService } from "@/services/apiService";
import { useRouter, usePathname } from 'next/navigation';
import authService from '@/services/authService';

jest.mock("@/services/apiService");
jest.mock("next/navigation", () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));
jest.mock('@/services/authService', () => ({
    isAuthenticated: jest.fn(),
}));
jest.mock('./Navbar', () => {
    return function DummyNavbar() {
        return <div data-testid="navbar">Navbar</div>;
    };
});
jest.mock('./Header', () => {
    return function DummyHeader({ config }: { config: { title: string } }) {
        return <div data-testid="header">{config.title}</div>;
    };
});

describe('OpenAILLMProviderForm', () => {
    const mockRouter = {
        push: jest.fn(),
    };
    const mockApiService = {
        llmProvider: {
            getLLMProvider: jest.fn(),
            createLLMProvider: jest.fn(),
            updateLLMProvider: jest.fn(),
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue(mockRouter);
        (usePathname as jest.Mock).mockReturnValue('/llm-providers/openai/add');
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
        (authService.isAuthenticated as jest.Mock).mockReturnValue(true);
    });

    it('renders the form correctly', () => {
        render(<OpenAILLMProviderForm />);
        expect(screen.getByTestId('navbar')).toBeInTheDocument();
        expect(screen.getByTestId('header')).toHaveTextContent('Add OpenAI LLM Provider');
        expect(screen.getByLabelText('Name')).toBeInTheDocument();
        expect(screen.getByLabelText('API Key')).toBeInTheDocument();
        expect(screen.getByLabelText('Model ID')).toBeInTheDocument();
    });

    it('handles form submission for new provider', async () => {
        render(<OpenAILLMProviderForm />);

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Provider' } });
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'test-api-key' } });
        fireEvent.change(screen.getByLabelText('Model ID'), { target: { value: 'gpt-4' } });

        fireEvent.click(screen.getByText('Save Provider'));

        await waitFor(() => {
            expect(mockApiService.llmProvider.createLLMProvider).toHaveBeenCalledWith({
                name: 'Test Provider',
                provider: 'openai',
                status: 'active',
                configuration: {
                    api_key: 'test-api-key',
                    model_id: 'gpt-4',
                },
            });
            expect(mockRouter.push).toHaveBeenCalledWith('/llm-providers');
        });
    });

    it('handles form submission for existing provider', async () => {
        const providerId = 'existing-provider-id';
        mockApiService.llmProvider.getLLMProvider.mockResolvedValue({
            name: 'Existing Provider',
            provider: 'openai',
            status: 'active',
            configuration: {
                api_key: 'existing-api-key',
                model_id: 'gpt-3.5-turbo',
            },
        });

        render(<OpenAILLMProviderForm providerId={providerId} />);

        await waitFor(() => {
            expect(screen.getByLabelText('Name')).toHaveValue('Existing Provider');
            expect(screen.getByLabelText('API Key')).toHaveValue('existing-api-key');
            expect(screen.getByLabelText('Model ID')).toHaveValue('gpt-3.5-turbo');
        });

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Updated Provider' } });
        fireEvent.click(screen.getByText('Update Provider'));

        await waitFor(() => {
            expect(mockApiService.llmProvider.updateLLMProvider).toHaveBeenCalledWith(
                providerId,
                {
                    name: 'Updated Provider',
                    provider: 'openai',
                    status: 'active',
                    configuration: {
                        api_key: 'existing-api-key',
                        model_id: 'gpt-3.5-turbo',
                    },
                }
            );
            expect(mockRouter.push).toHaveBeenCalledWith('/llm-providers');
        });
    });

    it('displays error message on API failure', async () => {
        mockApiService.llmProvider.createLLMProvider.mockRejectedValue(new Error('API Error'));

        render(<OpenAILLMProviderForm />);

        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Provider' } });
        fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'test-api-key' } });
        fireEvent.click(screen.getByText('Save Provider'));

        await waitFor(() => {
            expect(screen.getByText('Failed to add LLM provider. Please try again.')).toBeInTheDocument();
        });
    });

    it('handles validation', () => {
        render(<OpenAILLMProviderForm />);

        const validateButton = screen.getByText('Validate');
        fireEvent.click(validateButton);

        expect(validateButton).toBeDisabled();
        expect(validateButton).toHaveTextContent('Validated');
    });
});