import React from 'react';
import { render, screen } from '@testing-library/react';
import AdminDashboardPage from './admin-dashboard-page';
import Backend from '../../app/clients/backend/backend';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('AdminDashboardPage', () => {
    it('shows aggregate-only dashboard statistics', async () => {
        const getAdminDashboard = jest.fn().mockResolvedValue({
            result: 'ok',
            stats: {
                users: 12,
                aniListLibrary: 34,
                mangaDexLibrary: 5,
                favorites: 17,
                trackedChapters: 89,
            },
        });
        mockedBackend.mockImplementation(() => ({ getAdminDashboard }));

        render(<AdminDashboardPage />);

        expect(await screen.findByText('12')).toBeInTheDocument();
        expect(screen.getByText('AniList library entries')).toBeInTheDocument();
        expect(screen.getByText('34')).toBeInTheDocument();
        expect(screen.getByText('Tracked chapters')).toBeInTheDocument();
        expect(screen.getByText('89')).toBeInTheDocument();
        expect(getAdminDashboard).toHaveBeenCalledTimes(1);
    });

    it('shows an error and a retry action when the dashboard request fails', async () => {
        const getAdminDashboard = jest.fn().mockRejectedValue(new Error('Request failed with status code 500'));
        mockedBackend.mockImplementation(() => ({ getAdminDashboard }));

        render(<AdminDashboardPage />);

        expect(await screen.findByText(/Dashboard statistics could not be loaded/)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument();
    });
});
