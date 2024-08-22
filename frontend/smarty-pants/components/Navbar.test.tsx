// File: components/Navbar.test.tsx
import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import Navbar from './Navbar';

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

// Mock window.location
const mockPathname = jest.fn();
Object.defineProperty(window, 'location', {
    value: { pathname: mockPathname },
    writable: true,
});

describe('Navbar', () => {
    it('renders without crashing', () => {
        render(<Navbar />);
        expect(screen.getByText('SmartyPants')).toBeInTheDocument();
    });

    it('displays all top-level navigation items', () => {
        render(<Navbar />);

        const topLevelItems = ['Home', 'Assistant', 'Datasources', 'AI Providers', 'Management'];

        topLevelItems.forEach(item => {
            const elements = screen.getAllByText(item);
            const topLevelElement = elements.find(el =>
                el.tagName === 'BUTTON' || (el.tagName === 'A' && el.getAttribute('href'))
            );
            expect(topLevelElement).toBeInTheDocument();
        });
    });

    it('highlights the current active route', () => {
        render(<Navbar initialPath="/" />);
        const homeLink = screen.getByText('Home').closest('a');
        expect(homeLink).toHaveClass('bg-gray-900');
        expect(homeLink).toHaveClass('text-white');
    });

    it('opens dropdown menu when clicked', () => {
        render(<Navbar />);
        const assistantDropdown = screen.getByText('Assistant');
        fireEvent.click(assistantDropdown);
        expect(screen.getByText('Conversation')).toBeVisible();
    });

    it('highlights parent item when child route is active', () => {
        render(<Navbar initialPath="/ask" />);
        const assistantDropdown = screen.getByText('Assistant').closest('button');
        expect(assistantDropdown).toHaveClass('bg-gray-900');
        expect(assistantDropdown).toHaveClass('text-white');
    });

    it('handles window popstate event', () => {
        render(<Navbar />);
        act(() => {
            window.dispatchEvent(new PopStateEvent('popstate'));
        });
        // This test just ensures that the event listener doesn't throw an error
        expect(true).toBeTruthy();
    });

    it('removes popstate event listener on unmount', () => {
        const removeEventListenerSpy = jest.spyOn(window, 'removeEventListener');
        const { unmount } = render(<Navbar />);
        unmount();
        expect(removeEventListenerSpy).toHaveBeenCalledWith('popstate', expect.any(Function));
    });
});