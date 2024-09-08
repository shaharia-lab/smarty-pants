import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatToggleButton from '../ChatToggleButton';

describe('ChatToggleButton', () => {
    const mockOnClick = jest.fn();

    beforeEach(() => {
        mockOnClick.mockClear();
    });

    it('renders the button with SVG logo', () => {
        render(<ChatToggleButton onClick={mockOnClick} />);
        expect(screen.getByRole('button', { name: 'Toggle chat' })).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Chat logo' })).toBeInTheDocument();
    });

    it('calls onClick when button is clicked', () => {
        render(<ChatToggleButton onClick={mockOnClick} />);
        fireEvent.click(screen.getByRole('button', { name: 'Toggle chat' }));
        expect(mockOnClick).toHaveBeenCalled();
    });
});