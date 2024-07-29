import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import AvailableProviders from './AvailableProviders';

// Mock Next.js components
jest.mock('next/image', () => ({
    __esModule: true,
    default: (props: any) => {
        return <img {...props} />;
    },
}));

jest.mock('next/link', () => ({
    __esModule: true,
    default: ({ children, href }: { children: React.ReactNode; href: string }) => {
        return <a href={href}>{children}</a>;
    },
}));

describe('AvailableProvidersComponent', () => {
    const mockProviders = [
        {
            id: '1',
            name: 'Provider 1',
            description: 'Description 1',
            imageUrl: '/provider1.png',
            configurationUrl: '/configure/1',
        },
        {
            id: '2',
            name: 'Provider 2',
            description: 'Description 2',
            imageUrl: '/provider2.png',
            configurationUrl: '/configure/2',
        },
    ];

    it('renders the title correctly', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        expect(screen.getByText('Available Providers')).toBeInTheDocument();
    });

    it('renders all providers', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        expect(screen.getAllByRole('heading', { level: 3 })).toHaveLength(2);
    });

    it('displays correct provider information', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        expect(screen.getByText('Provider 1')).toBeInTheDocument();
        expect(screen.getByText('Description 1')).toBeInTheDocument();
        expect(screen.getByText('Provider 2')).toBeInTheDocument();
        expect(screen.getByText('Description 2')).toBeInTheDocument();
    });

    it('renders images with correct src and alt text', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        const images = screen.getAllByRole('img');
        expect(images[0]).toHaveAttribute('src', '/provider1.png');
        expect(images[0]).toHaveAttribute('alt', 'Provider 1 icon');
        expect(images[1]).toHaveAttribute('src', '/provider2.png');
        expect(images[1]).toHaveAttribute('alt', 'Provider 2 icon');
    });

    it('renders configure links with correct hrefs', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        const links = screen.getAllByText('Configure');
        expect(links[0]).toHaveAttribute('href', '/configure/1');
        expect(links[1]).toHaveAttribute('href', '/configure/2');
    });

    it('uses default image when imageUrl is not provided', () => {
        const providerWithoutImage = [
            {
                id: '3',
                name: 'Provider 3',
                description: 'Description 3',
                imageUrl: '',
                configurationUrl: '/configure/3',
            },
        ];
        render(<AvailableProviders title="Available Providers" providers={providerWithoutImage} />);
        const image = screen.getByRole('img');
        expect(image).toHaveAttribute('src', '/default-provider-icon.png');
    });

    it('renders correctly with no providers', () => {
        render(<AvailableProviders title="No Providers" providers={[]} />);
        expect(screen.getByText('No Providers')).toBeInTheDocument();
        expect(screen.queryByRole('heading', { level: 3 })).not.toBeInTheDocument();
    });

    it('applies correct CSS classes', () => {
        render(<AvailableProviders title="Available Providers" providers={mockProviders} />);
        expect(screen.getByText('Available Providers').parentElement).toHaveClass('w-1/2');
    });
});