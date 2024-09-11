import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatHistories from './ChatHistories';
import { createApiService } from "@/services/apiService";
import AuthService from "@/services/authService";

jest.mock("@/services/apiService");
jest.mock("@/services/authService");

describe('ChatHistories', () => {
    const mockOnSelectInteraction = jest.fn();
    const mockInteractions = [
        { uuid: '1', title: 'Chat 1' },
        { uuid: '2', title: 'Chat 2' },
    ];
    const mockApiService = {
        chatHisories: {
            getChatHistories: jest.fn(),
        },
    };

    beforeEach(() => {
        jest.resetAllMocks();
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
    });

    it('renders loading state initially', async () => {
        mockApiService.chatHisories.getChatHistories.mockImplementation(
            () => new Promise(resolve => setTimeout(() => resolve({ interactions: [] }), 100))
        );

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        expect(screen.getByText('Loading...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });
    });

    it('fetches and renders chat histories', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({ interactions: mockInteractions });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat 1')).toBeInTheDocument();
            expect(screen.getByText('Chat 2')).toBeInTheDocument();
        });

        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('calls onSelectInteraction when a chat is clicked', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({ interactions: mockInteractions });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat 1')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Chat 1'));

        expect(mockOnSelectInteraction).toHaveBeenCalledWith('1');
    });

    it('handles fetch error gracefully', async () => {
        console.error = jest.fn();
        mockApiService.chatHisories.getChatHistories.mockRejectedValue(new Error('API error'));

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(console.error).toHaveBeenCalledWith('Error fetching chat histories:', expect.any(Error));
        });

        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('renders correct heading', () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({ interactions: [] });
        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);
        expect(screen.getByText('Chat Histories')).toBeInTheDocument();
    });

    it('renders empty list when no histories are returned', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({ interactions: [] });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.queryByRole('listitem')).not.toBeInTheDocument();
        });
    });

    it('applies correct CSS classes', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({ interactions: mockInteractions });

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat Histories').parentElement).toHaveClass('bg-white shadow-md rounded-lg overflow-hidden');
            expect(screen.getByText('Chat 1').parentElement).toHaveClass('p-4 hover:bg-gray-50 cursor-pointer');
        });
    });
});