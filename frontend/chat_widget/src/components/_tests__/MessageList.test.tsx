import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import MessageList from '../MessageList';

describe('MessageList', () => {
    const mockMessages = [
        { text: 'Hello', isUser: true },
        { text: 'Hi there!', isUser: false },
    ];

    it('renders messages correctly', () => {
        render(<MessageList messages={mockMessages} />);
        expect(screen.getByText('Hello')).toBeInTheDocument();
        expect(screen.getByText('Hi there!')).toBeInTheDocument();
    });

    it('applies correct classes for user and bot messages', () => {
        render(<MessageList messages={mockMessages} />);
        expect(screen.getByText('Hello').closest('.chat-message')).toHaveClass('user-message');
        expect(screen.getByText('Hi there!').closest('.chat-message')).toHaveClass('bot-message');
    });
});