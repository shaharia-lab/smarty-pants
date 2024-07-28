import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ListComponent from './ListComponent';
import '@testing-library/jest-dom';


// Mock Next.js Link component
jest.mock('next/link', () => {
    return ({ children, href }: { children: React.ReactNode; href: string }) => (
        <a href={href}>{children}</a>
    );
});

// Mock Next.js Image component
jest.mock('next/image', () => ({
    __esModule: true,
    default: (props: any) => {
        return <img {...props} />;
    },
}));

describe('ListComponent', () => {
    const mockItems = [
        {
            id: '1',
            name: 'Test Item 1',
            sourceType: "openai",
            status: 'active',
            imageUrl: '/test-image-1.png',
            onDelete: jest.fn(),
            onDeactivate: jest.fn(),
        },
        {
            id: '2',
            name: 'Test Item 2',
            status: 'inactive',
            sourceType: "openai",
            imageUrl: '/test-image-2.png',
            onDelete: jest.fn(),
            onActivate: jest.fn(),
            additionalInfo: 'Additional Info',
        },
    ];

    it('renders the component title', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        expect(screen.getByText('Test List')).toBeInTheDocument();
    });

    it('displays loading message when loading prop is true', () => {
        render(
            <ListComponent
                title="Test List"
                items={[]}
                loading={true}
                error={null}
                type="llm"
            />
        );
        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('displays error message when error prop is provided', () => {
        render(
            <ListComponent
                title="Test List"
                items={[]}
                loading={false}
                error="Test error message"
                type="llm"
            />
        );
        expect(screen.getByText('Test error message')).toBeInTheDocument();
    });

    it('displays "No items to display" message when items array is empty', () => {
        render(
            <ListComponent
                title="Test List"
                items={[]}
                loading={false}
                error={null}
                type="llm"
            />
        );
        expect(screen.getByText('No items to display.')).toBeInTheDocument();
    });

    it('renders list items correctly', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        expect(screen.getByText('Test Item 1')).toBeInTheDocument();
        expect(screen.getByText('Test Item 2')).toBeInTheDocument();
        expect(screen.getByText('Status: active')).toBeInTheDocument();
        expect(screen.getByText('Status: inactive')).toBeInTheDocument();
    });

    it('renders correct buttons based on item status', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        expect(screen.getByText('Deactivate')).toBeInTheDocument();
        expect(screen.getByText('Activate')).toBeInTheDocument();
        expect(screen.getAllByText('Delete')).toHaveLength(2);
    });

    it('calls onDelete function when Delete button is clicked', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        fireEvent.click(screen.getAllByText('Delete')[0]);
        expect(mockItems[0].onDelete).toHaveBeenCalledWith('1');
    });

    it('calls onActivate function when Activate button is clicked', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        fireEvent.click(screen.getByText('Activate'));
        expect(mockItems[1].onActivate).toHaveBeenCalledWith('2');
    });

    it('calls onDeactivate function when Deactivate button is clicked', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        fireEvent.click(screen.getByText('Deactivate'));
        expect(mockItems[0].onDeactivate).toHaveBeenCalledWith('1');
    });

    it('renders correct edit link based on type prop', () => {
        render(
            <ListComponent
                title="Test List"
                items={mockItems}
                loading={false}
                error={null}
                type="llm"
            />
        );
        expect(screen.getAllByText('Edit')[0]).toHaveAttribute('href', '/llm-providers/openai/1');
    });
});