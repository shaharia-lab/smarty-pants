import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import UserList from './UserList';
import { User, UserRole, UserStatus } from '@/types/user';
import { createApiService } from "@/services/apiService";

// Mock createApiService
jest.mock("@/services/apiService", () => ({
    createApiService: jest.fn(),
}));

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
    const mockApiService = {
        usersApi: {
            updateUserStatus: jest.fn(),
            updateUserRoles: jest.fn(),
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
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

    test('filters users by status', () => {
        render(<UserList {...mockProps} />);
        const statusFilter = screen.getByTestId('status-filter') as HTMLSelectElement;
        fireEvent.change(statusFilter, { target: { value: 'inactive' } });
        expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
        expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });

    test('filters users by role', () => {
        render(<UserList {...mockProps} />);
        const roleFilter = screen.getByTestId('role-filter') as HTMLSelectElement;
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
        mockApiService.usersApi.updateUserStatus.mockResolvedValueOnce({ ...mockUsers[0], status: 'inactive' });
        render(<UserList {...mockProps} />);
        const showDetailsButton = screen.getAllByText('Show Details')[0];
        fireEvent.click(showDetailsButton);
        const deactivateButton = screen.getByText('Deactivate');
        fireEvent.click(deactivateButton);
        await waitFor(() => {
            expect(mockApiService.usersApi.updateUserStatus).toHaveBeenCalledWith('1', 'inactive');
        });
    });

    test('updates user roles', async () => {
        mockApiService.usersApi.updateUserRoles.mockResolvedValueOnce({ ...mockUsers[0], roles: ['user', 'developer'] });
        render(<UserList {...mockProps} />);
        const showDetailsButton = screen.getAllByText('Show Details')[0];
        fireEvent.click(showDetailsButton);
        const developerCheckbox = screen.getByLabelText('developer');
        fireEvent.click(developerCheckbox);
        await waitFor(() => {
            expect(mockApiService.usersApi.updateUserRoles).toHaveBeenCalledWith('1', ['user', 'developer']);
        });
    });

    test('calls onPageChange when pagination is used', () => {
        render(<UserList {...mockProps} totalPages={2} />);

        // Try to find the "Next" button in the main pagination controls
        const nextPageButtons = screen.getAllByRole('button', { name: /next/i });

        // Choose the first enabled "Next" button
        const enabledNextButton = nextPageButtons.find(button => !(button as HTMLButtonElement).disabled);

        if (enabledNextButton) {
            fireEvent.click(enabledNextButton);
            expect(mockProps.onPageChange).toHaveBeenCalledWith(2);
        } else {
            console.error('No enabled "Next" button found');
            throw new Error('No enabled "Next" button found');
        }
    });
});