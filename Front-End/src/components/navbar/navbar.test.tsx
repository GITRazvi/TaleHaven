import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import HorizontalNavbar from './navbar';
import { useAuth } from '../../contexts/AuthContext';
import Backend from '../../app/clients/backend/backend';

const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));
jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../home/guided-recommendation-chat', () => () => <div aria-label="Manga recommendation assistant">Guided recommender</div>);

describe('HorizontalNavbar recommendations', () => {
  beforeEach(() => {
    mockedBackend.mockImplementation(() => ({ getNotificationUnreadCount: jest.fn().mockResolvedValue({ unread: 0 }) }));
  });
  it('opens the guided recommendation chat in a right-side panel', async () => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(() => ({ matches: false, addListener: jest.fn(), removeListener: jest.fn(), addEventListener: jest.fn(), removeEventListener: jest.fn(), dispatchEvent: jest.fn() })),
    });
    mockedUseAuth.mockReturnValue({ user: null, loading: false, sessionExpired: false, logout: jest.fn() } as unknown as ReturnType<typeof useAuth>);
    render(
      <HorizontalNavbar
        searchQuery=""
        setSearchQuery={jest.fn()}
        siteTheme="forest"
        siteThemeStyle={{ '--theme-accent': '#77d6a8' } as React.CSSProperties}
        hasSiteThemeImage={false}
      />,
      { wrapper: MemoryRouter },
    );

    await userEvent.click(screen.getByRole('button', { name: 'Chatbot' }));

    expect(await screen.findByLabelText('Manga recommendation assistant')).toBeInTheDocument();
    expect(document.querySelector('.recommendation-offcanvas')).toHaveClass('chatbot-theme', 'site-theme-forest');
  });
});
