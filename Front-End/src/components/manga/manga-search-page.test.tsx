import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import MangaSearchPage from './manga-search-page';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('MangaSearchPage covers', () => {
  const searchManga = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    searchManga.mockResolvedValue({
      data: [{
        id: 'manga-1',
        title: 'Test Manga',
        description: 'A test description.',
        coverUrl: 'https://uploads.mangadex.org/covers/manga-1/test.jpg.512.jpg',
      }],
    });
    mockedBackend.mockImplementation(() => ({ searchManga }));
  });

  it('renders the cover returned for a search result', async () => {
    render(<MangaSearchPage />, {
      wrapper: ({ children }) => <MemoryRouter initialEntries={['/manga/search?title=test']}>{children}</MemoryRouter>,
    });

    const cover = await screen.findByRole('img', { name: 'Cover for Test Manga' });
    expect(cover).toHaveAttribute('src', 'https://uploads.mangadex.org/covers/manga-1/test.jpg.512.jpg');
    expect(cover.parentElement).toHaveClass('flex-row');
    expect(cover.closest('.col-lg-4')).not.toBeNull();
  });
});
