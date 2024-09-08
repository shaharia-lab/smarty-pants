import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import MessageInput from '../MessageInput';

describe('MessageInput', () => {
    const mockSendMessage = jest.fn();
    const mockProps = {
        onSendMessage: mockSendMessage,
        primaryColor: '#000000',
    };

    beforeEach(() => {
        mockSendMessage.mockClear();
    });

    it('renders input and send button', () => {
        render(<MessageInput {...mockProps} />);
        expect(screen.getByPlaceholderText('Type a message...')).toBeInTheDocument();
        expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('calls onSendMessage when form is submitted with non-empty input', () => {
        render(<MessageInput {...mockProps} />);
        const input = screen.getByPlaceholderText('Type a message...');
        fireEvent.change(input, { target: { value: 'Hello' } });
        fireEvent.click(screen.getByRole('button'));
        expect(mockSendMessage).toHaveBeenCalledWith('Hello');
    });

    it('does not call onSendMessage when form is submitted with empty input', () => {
        render(<MessageInput {...mockProps} />);
        fireEvent.click(screen.getByRole('button'));
        expect(mockSendMessage).not.toHaveBeenCalled();
    });

    it('clears input after sending a message', () => {
        render(<MessageInput {...mockProps} />);
        const input = screen.getByPlaceholderText('Type a message...');
        fireEvent.change(input, { target: { value: 'Hello' } });
        fireEvent.click(screen.getByRole('button'));
        expect(input).toHaveValue('');
    });
});