import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import AddOpenAIEmbeddingProviderPage from '../../../../app/embedding-providers/openai/add/page';
import OpenAIEmbeddingProviderForm from '../../../../components/OpenAIEmbeddingProviderForm';

// Mock the OpenAIEmbeddingProviderForm component
jest.mock('../../../../components/OpenAIEmbeddingProviderForm', () => {
    return jest.fn(() => <div data-testid="openai-embedding-provider-form">Mocked OpenAIEmbeddingProviderForm</div>);
});

describe('AddOpenAIEmbeddingProviderPage', () => {
    it('renders the OpenAIEmbeddingProviderForm', () => {
        render(<AddOpenAIEmbeddingProviderPage />);

        // Check if the OpenAIEmbeddingProviderForm is rendered
        const formElement = screen.getByTestId('openai-embedding-provider-form');
        expect(formElement).toBeInTheDocument();
        expect(formElement).toHaveTextContent('Mocked OpenAIEmbeddingProviderForm');

        // Verify that the OpenAIEmbeddingProviderForm component was called
        expect(OpenAIEmbeddingProviderForm).toHaveBeenCalled();
    });

    it('renders the OpenAIEmbeddingProviderForm without props', () => {
        render(<AddOpenAIEmbeddingProviderPage />);

        // Verify that the OpenAIEmbeddingProviderForm component was called without props
        expect(OpenAIEmbeddingProviderForm).toHaveBeenCalledWith({}, {});
    });
});