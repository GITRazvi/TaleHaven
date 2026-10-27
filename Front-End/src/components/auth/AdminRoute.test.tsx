import React from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { render, screen } from '@testing-library/react';
import AdminRoute from './AdminRoute';
import { useAuth } from '../../contexts/AuthContext';

jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

const renderAdminRoute = () => render(
    <MemoryRouter initialEntries={['/admin']}>
        <Routes>
            <Route path="/admin" element={<AdminRoute><h1>Administrator dashboard</h1></AdminRoute>} />
            <Route path="/login" element={<h1>Login</h1>} />
        </Routes>
    </MemoryRouter>,
);

describe('AdminRoute', () => {
    it('redirects a signed-out visitor to login', () => {
        mockedUseAuth.mockReturnValue({ user: null, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);

        renderAdminRoute();

        expect(screen.getByRole('heading', { name: 'Login' })).toBeInTheDocument();
    });

    it('blocks a non-administrator from the page', () => {
        mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '' }, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);

        renderAdminRoute();

        expect(screen.getByRole('alert')).toHaveTextContent('Administrator access is required');
        expect(screen.queryByRole('heading', { name: 'Administrator dashboard' })).not.toBeInTheDocument();
    });

    it('allows an administrator to open the page', () => {
        mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'admin@example.test', role: 'admin', displayName: 'Admin', profileImagePath: '', created_at: '' }, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);

        renderAdminRoute();

        expect(screen.getByRole('heading', { name: 'Administrator dashboard' })).toBeInTheDocument();
    });
});
