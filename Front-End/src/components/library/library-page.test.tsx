import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import LibraryPage from './library-page';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('LibraryPage pagination', () => {
  const getMyLibrary = jest.fn();
  const getReadingProgress = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    getMyLibrary
      .mockResolvedValueOnce({
        result: 'ok',
        count: 1,
        total: 2,
        limit: 24,
        offset: 0,
        data: [{
          mangaId: 'manga-1',
          title: 'First Manga',
          coverUrl: '',
          addedAt: '2026-10-20T00:00:00.000Z',
        }],
      })
      .mockResolvedValueOnce({
        result: 'ok',
        count: 1,
        total: 2,
        limit: 24,
        offset: 1,
        data: [{
          mangaId: 'manga-2',
          title: 'Second Manga',
          coverUrl: '',
          addedAt: '2026-10-19T00:00:00.000Z',
        }],
      });
    getReadingProgress.mockResolvedValue({ data: { mangas: {} } });
    mockedBackend.mockImplementation(() => ({ getMyLibrary, getReadingProgress }));
  });

  it('requests another page only when Load more is clicked', async () => {
    render(<LibraryPage />, { wrapper: MemoryRouter });

    expect(await screen.findByText('First Manga')).toBeInTheDocument();
    expect(screen.queryByText('Second Manga')).not.toBeInTheDocument();
    expect(getMyLibrary).toHaveBeenCalledWith({ limit: 24, offset: 0 });

    await userEvent.click(screen.getByRole('button', { name: 'Load more (1 of 2)' }));

    await waitFor(() => {
      expect(getMyLibrary).toHaveBeenLastCalledWith({ limit: 24, offset: 1 });
      expect(screen.getByText('Second Manga')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: /Load more/ })).not.toBeInTheDocument();
  });
});
