import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import PublicProfilePage from './public-profile-page';
import { useAuth } from '../../contexts/AuthContext';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

describe('PublicProfilePage', () => {
  it('loads and renders only the selected member profile', async () => {
		mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    const getPublicProfile = jest.fn().mockResolvedValue({
      profile: {
        id: 7,
        displayName: 'Chat Friend',
        profileImagePath: '/images/profiles/friend.jpg',
        createdAt: '2026-10-21T00:00:00Z',
		commentsEnabled: true,
      },
    });
		const getProfileComments = jest.fn().mockResolvedValue({ data: [{ id: 2, profileUserId: 7, userId: 3, authorDisplayName: 'Another Reader', profileImagePath: '', body: 'Hello there!', createdAt: '2026-10-22T00:00:00Z' }] });
		const createProfileComment = jest.fn().mockResolvedValue({ comment: { id: 3, profileUserId: 7, userId: 1, authorDisplayName: 'Reader', profileImagePath: '', body: 'Welcome!', createdAt: '2026-10-28T00:00:00Z' } });
    mockedBackend.mockImplementation(() => ({ getPublicProfile, getProfileComments, createProfileComment, getBaseUrl: () => 'http://localhost:3001' }));

    render(<MemoryRouter initialEntries={['/profile/7']}><Routes><Route path="/profile/:userID" element={<PublicProfilePage />} /></Routes></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: 'Chat Friend' })).toBeInTheDocument();
    expect(getPublicProfile).toHaveBeenCalledWith(7);
    expect(screen.getByRole('img', { name: "Chat Friend's profile" })).toHaveAttribute('src', 'http://localhost:3001/images/profiles/friend.jpg');
    expect(screen.getByText(/Member since/)).toBeInTheDocument();
		expect(await screen.findByText('Hello there!')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Another Reader' })).toHaveAttribute('href', '/profile/3');
		await userEvent.type(screen.getByRole('textbox', { name: 'Leave a comment' }), 'Welcome!');
		await userEvent.keyboard('{Enter}');
		await waitFor(() => expect(createProfileComment).toHaveBeenCalledWith(7, 'Welcome!'));
    expect(screen.getByRole('link', { name: '← Back to chatrooms' })).toHaveAttribute('href', '/chat');
  });
});
