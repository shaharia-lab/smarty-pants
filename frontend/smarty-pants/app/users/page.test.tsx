import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import UsersPage from './page';
import { User, PaginatedUsers } from '@/types/user';
import { useRouter } from 'next/navigation';
import AuthService from '@/services/authService';
import { createApiService } from "@/services/apiService";

// Mock Next.js modules
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
    usePathname: jest.fn(),
}));

// Mock auth service
jest.mock('@/services/authService', () => ({
    isAuthenticated: jest.fn(),
    getAuthenticatedAxiosInstance: jest.fn(),
}));

// Mock createApiService
jest.mock("@/services/apiService", () => ({
    createApiService: jest.fn(),
}));

// Mock the components
jest.mock('@/components/Navbar', () => {
    const MockNavbar = () => <div data-testid="navbar">Navbar</div>;
    MockNavbar.displayName = 'MockNavbar';
    return MockNavbar;
});

jest.mock('@/components/Header', () => {
    const MockHeader = ({ config }: { config: { title: string } }) => <div data-testid="header">{config.title}</div>;
    MockHeader.displayName = 'MockHeader';
    return MockHeader;
});

jest.mock('../../components/UserList', () => {
    const MockUserList = ({ users }: { users: User[] }) => (
        <div data-testid="user-list">
            {users.map(user => <div key={user.uuid}>{user.name}</div>)}
        </div>
    );
    MockUserList.displayName = 'MockUserList';
    return MockUserList;
});

describe('UsersPage Component', () => {
    const mockRouter = { push: jest.fn() };
    const mockApiService = {
        usersApi: {
            getUsers: jest.fn(),
            updateUserStatus: jest.fn(),
            updateUserRoles: jest.fn(),
        },
    };

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

    beforeEach(() => {
        jest.clearAllMocks();
        (useRouter as jest.Mock).mockReturnValue(mockRouter);
        (AuthService.isAuthenticated as jest.Mock).mockReturnValue(true);
        (createApiService as jest.Mock).mockReturnValue(mockApiService);
    });

    it('renders loading state initially', async () => {
        mockApiService.usersApi.getUsers.mockReturnValue(new Promise(() => {}));

        await act(async () => {
            render(<UsersPage />);
        });

        expect(screen.getByText('Loading users...')).toBeInTheDocument();
    });

    it('renders UsersPage correctly with users', async () => {
        mockApiService.usersApi.getUsers.mockResolvedValue(mockPaginatedUsers);

        await act(async () => {
            render(<UsersPage />);
        });

        expect(screen.getByTestId('navbar')).toBeInTheDocument();
        expect(screen.getByTestId('header')).toHaveTextContent('User Management');

        await waitFor(() => {
            expect(screen.getByTestId('user-list')).toBeInTheDocument();
        });

        expect(screen.getByText('John Doe')).toBeInTheDocument();
        expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });

    it('handles fetch error', async () => {
        mockApiService.usersApi.getUsers.mockRejectedValue(new Error('API Error'));

        await act(async () => {
            render(<UsersPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Error fetching users. Please try again.')).toBeInTheDocument();
        });
    });

    it('fetches users on initial render', async () => {
        mockApiService.usersApi.getUsers.mockResolvedValue(mockPaginatedUsers);

        await act(async () => {
            render(<UsersPage />);
        });

        await waitFor(() => {
            expect(mockApiService.usersApi.getUsers).toHaveBeenCalledWith(1, 10, expect.anything());
        });
    });
});