import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import Navbar from './Navbar';
import { useRouter, usePathname } from 'next/navigation';
import authService from '../services/authService';

// Mock the next/navigation hooks
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));

// Mock the authService
jest.mock('../services/authService', () => ({
    isAuthenticated: jest.fn(),
    logout: jest.fn(),
}));

// Mock the next/link component
jest.mock('next/link', () => {
    const MockedLink = React.forwardRef<HTMLAnchorElement, { children: React.ReactNode; href: string; className: string }>(
        ({ children, href, className }, ref) => (
            <a href={href} className={className} ref={ref}>
                {children}
            </a>
        )
    );
    MockedLink.displayName = 'MockedLink';
    return MockedLink;
});

jest.mock('./SVGLogo', () => {
    return function DummySVGLogo(props: any) {
        return <div data-testid="svg-logo" data-props={JSON.stringify(props)}>SVG Logo</div>;
    };
});

// Mock window.location
const mockPathname = jest.fn();
Object.defineProperty(window, 'location', {
    value: { pathname: mockPathname },
    writable: true,
});

describe('Navbar', () => {
    beforeEach(() => {
        (useRouter as jest.Mock).mockReturnValue({ push: jest.fn() });
        (usePathname as jest.Mock).mockReturnValue('/');
        (authService.isAuthenticated as jest.Mock).mockReturnValue(false);
    });

    it('renders without crashing', () => {
        render(<Navbar />);
        expect(screen.getByText('SmartyPants')).toBeInTheDocument();
    });

    it('displays all top-level navigation items', () => {
        render(<Navbar />);
        const topLevelItems = ['Home', 'Assistant', 'Datasources', 'AI Providers', 'Management'];
        topLevelItems.forEach(item => {
            expect(screen.getByText(item)).toBeInTheDocument();
        });
    });

    it('highlights the current active route', () => {
        (usePathname as jest.Mock).mockReturnValue('/');
        render(<Navbar />);
        const homeLink = screen.getByText('Home').closest('a');
        expect(homeLink).toHaveClass('bg-gray-900 text-white');
    });

    it('opens dropdown menu when clicked', () => {
        render(<Navbar />);
        fireEvent.click(screen.getByText('Assistant'));
        expect(screen.getByText('Conversation')).toBeVisible();
    });

    it('highlights parent item when child route is active', () => {
        (usePathname as jest.Mock).mockReturnValue('/ask');
        render(<Navbar />);
        const assistantDropdown = screen.getByText('Assistant').closest('button');
        expect(assistantDropdown).toHaveClass('bg-gray-900 text-white');
    });

    it('renders login button when unauthenticated', () => {
        render(<Navbar />);
        expect(screen.getByText('Login')).toBeInTheDocument();
    });

    it('renders logout button when authenticated', () => {
        (authService.isAuthenticated as jest.Mock).mockReturnValue(true);
        render(<Navbar />);
        expect(screen.getByText('Logout')).toBeInTheDocument();
    });

    it('calls logout function and redirects on logout button click', () => {
        (authService.isAuthenticated as jest.Mock).mockReturnValue(true);
        const routerPush = jest.fn();
        (useRouter as jest.Mock).mockReturnValue({ push: routerPush });
        render(<Navbar />);
        fireEvent.click(screen.getByText('Logout'));
        expect(authService.logout).toHaveBeenCalled();
        expect(routerPush).toHaveBeenCalledWith('/login');
    });

    it('renders the logo and brand name together', () => {
        render(<Navbar />);
        const logoContainer = screen.getByTestId('svg-logo').closest('a');
        expect(logoContainer).toContainElement(screen.getByText('SmartyPants'));
    });
});