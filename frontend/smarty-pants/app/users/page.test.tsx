import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import UsersPage from './page';
import { User, PaginatedUsers } from '@/types/user';

// Mock the components and fetch function
jest.mock('@/components/Navbar', () => () => <div data-testid="navbar">Navbar</div>);
jest.mock('@/components/Header', () => ({ config }: { config: { title: string } }) => <div data-testid="header">{config.title}</div>);
jest.mock('../../components/UserList', () => ({ users }: { users: User[] }) => (
    <div data-testid="user-list">
        {users.map(user => <div key={user.uuid}>{user.name}</div>)}
    </div>
));

global.fetch = jest.fn();

const mockUsers: User[] = [
    { uuid: '1', name: 'John Doe', email: 'john@example.com', status: 'active', roles: ['user'], created_at: '2023-01-01T00:00:00Z', updated_at: '2023-01-01T00:00:00Z' },
    { uuid: '2', name: 'Jane Smith', email: 'jane@example.com', status: 'inactive', roles: ['admin'], created_at: '2023-01-02T00:00:00Z', updated_at: '2023-01-02T00:00:00Z' },
];

const mockPaginatedUsers: PaginatedUsers = {
    users: mockUsers,
    total_pages: 2,
    page: 1,
    per_page: 10,
    total: 15,
};

describe('UsersPage Component', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (global.fetch as jest.Mock).mockResolvedValue({
            ok: true,
            json: async () => mockPaginatedUsers,
        });
    });

    test('renders UsersPage correctly', async () => {
        render(<UsersPage />);

        expect(screen.getByTestId('navbar')).toBeInTheDocument();
        expect(screen.getByTestId('header')).toHaveTextContent('User Management');

        await waitFor(() => {
            expect(screen.getByTestId('user-list')).toBeInTheDocument();
        });

        expect(screen.getByText('John Doe')).toBeInTheDocument();
        expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });

    test('displays loading state', () => {
        (global.fetch as jest.Mock).mockImplementationOnce(() => new Promise(() => {}));
        render(<UsersPage />);
        expect(screen.getByText('Loading users...')).toBeInTheDocument();
    });

    test('handles fetch error', async () => {
        (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('API Error'));
        render(<UsersPage />);

        await waitFor(() => {
            expect(screen.getByText('Error fetching users. Please try again.')).toBeInTheDocument();
        });
    });

    test('fetches users on initial render', async () => {
        render(<UsersPage />);

        await waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/v1/users?page=1&per_page=10')
            );
        });
    });
});