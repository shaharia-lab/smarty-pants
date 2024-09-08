import React from 'react';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatWidget from '../ChatWidget';

const mockConfig = {
    position: 'bottom-right' as const,
    title: 'Test Chat',
    primaryColor: '#000000',
    backend: {
        endpoint: 'https://test.com',
        api_key: 'test-key',
        widget_id: 'test-widget'
    }
};

describe('ChatWidget', () => {
    it('renders the toggle button when closed', () => {
        render(<ChatWidget {...mockConfig} />);
        expect(screen.getByRole('button', { name: /toggle chat/i })).toBeInTheDocument();
    });

    it('opens the chat window when toggle button is clicked', async () => {
        render(<ChatWidget {...mockConfig} />);
        fireEvent.click(screen.getByRole('button', { name: /toggle chat/i }));
        await waitFor(() => {
            expect(screen.getByText('Test Chat')).toBeInTheDocument();
        });
    });

    it('sends a message and receives a response', async () => {
        render(<ChatWidget {...mockConfig} />);
        fireEvent.click(screen.getByRole('button'));

        const input = screen.getByPlaceholderText('Type a message...');
        fireEvent.change(input, { target: { value: 'Hello' } });
        fireEvent.click(screen.getByRole('button', { name: /send/i }));

        expect(await screen.findByText('Hello')).toBeInTheDocument();
        expect(await screen.findByText(/This is a dummy response/)).toBeInTheDocument();
    });
});