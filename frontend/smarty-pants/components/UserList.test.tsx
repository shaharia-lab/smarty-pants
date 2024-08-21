import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import UserList from './UserList';
import { User, UserRole, UserStatus } from '@/types/user';

// Mock fetch function
global.fetch = jest.fn();

const mockUsers: User[] = [
    {
        uuid: '1',
        name: 'John Doe',
        email: 'john@example.com',
        status: 'active',
        roles: ['user'],
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-01-01T00:00:00Z',
    },
    {
        uuid: '2',
        name: 'Jane Smith',
        email: 'jane@example.com',
        status: 'inactive',
        roles: ['developer', 'admin'],
        created_at: '2023-01-02T00:00:00Z',
        updated_at: '2023-01-02T00:00:00Z',
    },
];

const mockProps = {
    users: mockUsers,
    currentPage: 1,
    totalPages: 1,
    onPageChange: jest.fn(),
};

describe('UserList Component', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('renders user list correctly', () => {
        render(<UserList {...mockProps} />);
        expect(screen.getByText('John Doe')).toBeInTheDocument();
        expect(screen.getByText('jane@example.com')).toBeInTheDocument();
    });

    test('filters users by name', () => {
        render(<UserList {...mockProps} />);
        const nameFilter = screen.getByPlaceholderText('Filter by name');
        fireEvent.change(nameFilter, { target: { value: 'John' } });
        expect(screen.getByText('John Doe')).toBeInTheDocument();
        expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument();
    });

    test('filters users by email', () => {
        render(<UserList {...mockProps} />);
        const emailFilter = screen.getByPlaceholderText('Filter by email');
        fireEvent.change(emailFilter, { target: { value: 'jane' } });
        expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
        expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });

    test('filters users by role', () => {
        render(<UserList {...mockProps} />);
        const roleFilter = screen.getAllByRole('combobox', { name: '' })[1];
        fireEvent.change(roleFilter, { target: { value: 'admin' } });
        expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
        expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });

    test('expands user details', () => {
        render(<UserList {...mockProps} />);
        const showDetailsButton = screen.getAllByText('Show Details')[0];
        fireEvent.click(showDetailsButton);
        expect(screen.getByText('User Details')).toBeInTheDocument();
        expect(screen.getByText('Manage Status')).toBeInTheDocument();
        expect(screen.getByText('Manage Roles')).toBeInTheDocument();
    });

    test('updates user status', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({ ok: true });
        render(<UserList {...mockProps} />);
        const showDetailsButton = screen.getAllByText('Show Details')[0];
        fireEvent.click(showDetailsButton);
        const deactivateButton = screen.getByText('Deactivate');
        fireEvent.click(deactivateButton);
        await waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/v1/users/1/deactivate'),
                expect.objectContaining({ method: 'PUT' })
            );
        });
    });

    test('updates user roles', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({ ok: true });
        render(<UserList {...mockProps} />);
        const showDetailsButton = screen.getAllByText('Show Details')[0];
        fireEvent.click(showDetailsButton);
        const developerCheckbox = screen.getByLabelText('developer');
        fireEvent.click(developerCheckbox);
        await waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/v1/users/1'),
                expect.objectContaining({
                    method: 'PUT',
                    body: JSON.stringify({ roles: ['user', 'developer'] }),
                })
            );
        });
    });
});