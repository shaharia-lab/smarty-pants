import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ChatHeader from '../ChatHeader';

describe('ChatHeader', () => {
    const mockProps = {
        title: 'Test Chat',
        branding: { name: 'Test Brand' },
        onClose: jest.fn(),
    };

    it('renders the title', () => {
        render(<ChatHeader {...mockProps} />);
        expect(screen.getByText('Test Chat')).toBeInTheDocument();
    });

    it('renders the SVG logo when no custom logo is provided', () => {
        render(<ChatHeader {...mockProps} />);
        expect(screen.getByRole('img', { name: 'Brand logo' })).toBeInTheDocument();
    });

    it('renders a custom logo when provided', () => {
        const propsWithLogo = {
            ...mockProps,
            branding: { ...mockProps.branding, logo: 'test-logo.png' },
        };
        render(<ChatHeader {...propsWithLogo} />);
        expect(screen.getByAltText('Test Brand')).toBeInTheDocument();
    });

    it('calls onClose when close button is clicked', () => {
        render(<ChatHeader {...mockProps} />);
        fireEvent.click(screen.getByRole('button'));
        expect(mockProps.onClose).toHaveBeenCalled();
    });
});