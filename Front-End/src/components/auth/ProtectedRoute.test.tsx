import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import ProtectedRoute from './ProtectedRoute';
import { useAuth } from '../../contexts/AuthContext';

jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

const LoginDestination = () => {
  const location = useLocation();
  const expired = Boolean((location.state as { sessionExpired?: boolean } | null)?.sessionExpired);
  return <div>{expired ? 'Session expired login' : 'Login destination'}</div>;
};

const renderRoute = () => render(
  <MemoryRouter initialEntries={['/profile']}>
    <Routes>
      <Route path="/profile" element={<ProtectedRoute><div>Private profile</div></ProtectedRoute>} />
      <Route path="/login" element={<LoginDestination />} />
    </Routes>
  </MemoryRouter>,
);

describe('ProtectedRoute', () => {
  it('waits for the app-level session bootstrap', () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: true, sessionExpired: false } as ReturnType<typeof useAuth>);
    renderRoute();
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('redirects unauthenticated users to login', () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);
    renderRoute();
    expect(screen.getByText('Login destination')).toBeInTheDocument();
  });

  it('shows the expired-session login state after a protected request loses its session', () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, sessionExpired: true } as ReturnType<typeof useAuth>);
    renderRoute();
    expect(screen.getByText('Session expired login')).toBeInTheDocument();
  });

  it('renders protected content for an authenticated user', () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', role: 'user', created_at: '' }, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);
    renderRoute();
    expect(screen.getByText('Private profile')).toBeInTheDocument();
  });
});
