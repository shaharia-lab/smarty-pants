import React from 'react';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatHistories from './ChatHistories';

// Mock the fetch function
global.fetch = jest.fn();

describe('ChatHistories', () => {
    const mockOnSelectInteraction = jest.fn();
    const mockInteractions = [
        { uuid: '1', title: 'Chat 1' },
        { uuid: '2', title: 'Chat 2' },
    ];

    beforeEach(() => {
        jest.resetAllMocks();
    });

    it('renders loading state initially', async () => {
        // Mock a delayed API response
        (global.fetch as jest.Mock).mockImplementation(
            () => new Promise(resolve => setTimeout(() => resolve({ json: () => Promise.resolve({ interactions: [] }) }), 100))
        );

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        // Check for loading state
        expect(screen.getByText('Loading...')).toBeInTheDocument();

        // Wait for the component to update
        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });
    });

    it('fetches and renders chat histories', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            json: () => Promise.resolve({ interactions: mockInteractions }),
        });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat 1')).toBeInTheDocument();
            expect(screen.getByText('Chat 2')).toBeInTheDocument();
        });

        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('calls onSelectInteraction when a chat is clicked', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            json: () => Promise.resolve({ interactions: mockInteractions }),
        });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat 1')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Chat 1'));

        expect(mockOnSelectInteraction).toHaveBeenCalledWith('1');
    });

    it('handles fetch error gracefully', async () => {
        console.error = jest.fn(); // Mock console.error to prevent error output in test logs
        (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Fetch error'));

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(console.error).toHaveBeenCalledWith('Error fetching chat histories:', expect.any(Error));
        });

        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('renders correct heading', () => {
        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);
        expect(screen.getByText('Chat Histories')).toBeInTheDocument();
    });

    it('renders empty list when no histories are returned', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            json: () => Promise.resolve({ interactions: [] }),
        });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.queryByRole('listitem')).not.toBeInTheDocument();
        });
    });

    it('uses correct API endpoint', async () => {
        process.env.NEXT_PUBLIC_API_BASE_URL = 'http://test-api.com';

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            json: () => Promise.resolve({ interactions: mockInteractions }),
        });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith('http://test-api.com/api/v1/interactions');
        });
    });

    it('applies correct CSS classes', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            json: () => Promise.resolve({ interactions: mockInteractions }),
        });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat Histories').parentElement).toHaveClass('bg-white shadow-md rounded-lg overflow-hidden');
            expect(screen.getByText('Chat 1').parentElement).toHaveClass('p-4 hover:bg-gray-50 cursor-pointer');
        });
    });
});