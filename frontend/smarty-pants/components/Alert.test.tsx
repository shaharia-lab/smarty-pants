import React from 'react';
import { render, screen } from '@testing-library/react';
import { Alert, AlertDescription } from './Alert';  // Adjust the import path as needed

describe('Alert Component', () => {
    test('renders with default variant', () => {
        render(<Alert>Test Alert</Alert>);
        const alertElement = screen.getByRole('alert');
        expect(alertElement).toHaveTextContent('Test Alert');
        expect(alertElement).toHaveClass('bg-green-100', 'text-green-800', 'border', 'border-green-300');
    });

    test('renders with destructive variant', () => {
        render(<Alert variant="destructive">Destructive Alert</Alert>);
        const alertElement = screen.getByRole('alert');
        expect(alertElement).toHaveTextContent('Destructive Alert');
        expect(alertElement).toHaveClass('bg-red-100', 'text-red-800', 'border', 'border-red-300');
    });

    test('renders with AlertDescription', () => {
        render(
            <Alert>
                Alert Title
                <AlertDescription>Alert description</AlertDescription>
            </Alert>
        );
        const alertElement = screen.getByRole('alert');
        expect(alertElement).toHaveTextContent('Alert Title');
        expect(alertElement).toHaveTextContent('Alert description');
        const descriptionElement = screen.getByText('Alert description');
        expect(descriptionElement).toHaveClass('text-sm');
    });

    test('applies base classes to all variants', () => {
        const { rerender } = render(<Alert>Base Classes Test</Alert>);
        let alertElement = screen.getByRole('alert');
        expect(alertElement).toHaveClass('px-4', 'py-3', 'rounded-md', 'mb-4', 'text-sm');

        rerender(<Alert variant="destructive">Base Classes Test</Alert>);
        alertElement = screen.getByRole('alert');
        expect(alertElement).toHaveClass('px-4', 'py-3', 'rounded-md', 'mb-4', 'text-sm');
    });
});