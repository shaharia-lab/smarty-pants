import React from 'react';
import {render, screen, fireEvent, waitFor, act} from '@testing-library/react';
import '@testing-library/jest-dom';
import SettingsPage from './page';
import { createApiService } from "@/services/apiService";
import {HeaderConfig} from "@/components/Header";

// Mock the API service
jest.mock("@/services/apiService", () => ({
    createApiService: jest.fn(),
}));

// Mock the AuthService
jest.mock("@/services/authService", () => ({
    __esModule: true,
    default: {},
}));

// Mock the Navbar and Header components
jest.mock('../../components/Navbar', () => () => <div data-testid="navbar">Navbar</div>);
jest.mock('../../components/Header', () => ({ config }: { config: HeaderConfig }) => <div data-testid="header">{config.title}</div>);

describe('SettingsPage', () => {
    const mockSettingsApi = {
        getSettings: jest.fn(),
        updateSettings: jest.fn(),
    };

    beforeEach(() => {
        (createApiService as jest.Mock).mockReturnValue({ settingsApi: mockSettingsApi });
        mockSettingsApi.getSettings.mockResolvedValue({
            general: { application_name: 'Test App' },
            debugging: { log_level: 'info', log_format: 'json', log_output: 'stdout' },
            search: { per_page: 10 },
        });
    });

    it('renders the settings page with correct title', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });
        expect(screen.getByTestId('header')).toHaveTextContent('Settings');
    });

    it('loads and displays settings', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        await waitFor(() => {
            expect(screen.getByLabelText('Application Name')).toHaveValue('Test App');
            expect(screen.getByLabelText('Log Level')).toHaveValue('info');
            expect(screen.getByLabelText('Log Format')).toHaveValue('json');
            expect(screen.getByLabelText('Log Output')).toHaveValue('stdout');
            expect(screen.getByLabelText('Results Per Page')).toHaveValue(10);
        });
    });

    it('handles input changes', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        await act(async () => {
            fireEvent.change(screen.getByLabelText('Application Name'), { target: { value: 'New App Name' } });
        });

        expect(screen.getByLabelText('Application Name')).toHaveValue('New App Name');
    });

    it('submits updated settings', async () => {
        mockSettingsApi.updateSettings.mockResolvedValue({});

        await act(async () => {
            render(<SettingsPage />);
        });

        await act(async () => {
            fireEvent.change(screen.getByLabelText('Application Name'), { target: { value: 'New App Name' } });
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Save Settings'));
        });

        await waitFor(() => {
            expect(mockSettingsApi.updateSettings).toHaveBeenCalledWith(expect.objectContaining({
                general: expect.objectContaining({ application_name: 'New App Name' })
            }));
            expect(screen.getByText('Settings updated successfully')).toBeInTheDocument();
        });
    });

    it('handles API errors when loading settings', async () => {
        mockSettingsApi.getSettings.mockRejectedValue(new Error('API Error'));

        await act(async () => {
            render(<SettingsPage />);
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to load settings. Please try again later.')).toBeInTheDocument();
        });
    });

    it('handles API errors when updating settings', async () => {
        mockSettingsApi.updateSettings.mockRejectedValue(new Error('API Error'));

        await act(async () => {
            render(<SettingsPage />);
        });

        await act(async () => {
            fireEvent.click(screen.getByText('Save Settings'));
        });

        await waitFor(() => {
            expect(screen.getByText('Failed to update settings. Please try again.')).toBeInTheDocument();
        });
    });

    it('resets settings when Reset button is clicked', async () => {
        await act(async () => {
            render(<SettingsPage />);
        });

        await act(async () => {
            fireEvent.change(screen.getByLabelText('Application Name'), { target: { value: 'New App Name' } });
        });

        expect(screen.getByLabelText('Application Name')).toHaveValue('New App Name');

        await act(async () => {
            fireEvent.click(screen.getByText('Reset'));
        });

        await waitFor(() => {
            expect(screen.getByLabelText('Application Name')).toHaveValue('Test App');
        });
    });
});