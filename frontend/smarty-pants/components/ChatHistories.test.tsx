import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatHistories from './ChatHistories';
import { createApiService } from "@/services/apiService";
import AuthService from "@/services/authService";
import { PaginatedInteractionsResponse, Interaction } from "@/types/api";

jest.mock("@/services/apiService");
jest.mock("@/services/authService");
jest.mock("@/utils/common", () => ({
    truncateMessage: jest.fn((msg) => msg),
}));

describe('ChatHistories', () => {
    const mockOnSelectInteraction = jest.fn();
    const mockInteractions: Interaction[] = [
        { uuid: '1', query: 'Query 1', conversations: [{ role: 'user', text: 'User message 1' }] },
        { uuid: '2', query: 'Query 2', conversations: [{ role: 'user', text: 'User message 2' }] },
    ];
    const mockPaginatedResponse: PaginatedInteractionsResponse = {
        interactions: mockInteractions,
        total: 2,
        page: 1,
        per_page: 10,
        total_pages: 1
    };
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
            () => new Promise(resolve => setTimeout(() => resolve(mockPaginatedResponse), 100))
        );

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        expect(screen.getByText('Loading...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
        });
    });

    it('fetches and renders chat histories', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue(mockPaginatedResponse);

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getAllByRole('listitem')).toHaveLength(2);
        });

        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('calls onSelectInteraction when a chat is clicked', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue(mockPaginatedResponse);

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getAllByRole('listitem')).toHaveLength(2);
        });

        fireEvent.click(screen.getAllByRole('listitem')[0]);

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
        mockApiService.chatHisories.getChatHistories.mockResolvedValue(mockPaginatedResponse);
        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);
        expect(screen.getByText('Chat Histories')).toBeInTheDocument();
    });

    it('renders empty list when no histories are returned', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue({...mockPaginatedResponse, interactions: []});

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.queryByRole('listitem')).not.toBeInTheDocument();
        });
    });

    it('applies correct CSS classes', async () => {
        mockApiService.chatHisories.getChatHistories.mockResolvedValue(mockPaginatedResponse);

        render(<ChatHistories onSelectInteraction={mockOnSelectInteraction} />);

        await waitFor(() => {
            expect(screen.getByText('Chat Histories').parentElement).toHaveClass('bg-white shadow-md rounded-lg overflow-hidden');
            expect(screen.getAllByRole('listitem')[0]).toHaveClass('p-4 hover:bg-gray-50 cursor-pointer');
        });
    });
});