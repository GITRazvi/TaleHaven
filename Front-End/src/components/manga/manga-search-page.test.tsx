import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import MangaSearchPage from './manga-search-page';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('MangaSearchPage covers', () => {
  const searchAniListManga = jest.fn();
  const getMangaDexBrowse = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    searchAniListManga.mockResolvedValue({
      data: [{
        id: 1,
        title: { english: 'Test Manga', userPreferred: '', romaji: '', native: '' },
        description: 'A test description.',
        coverImage: { large: 'https://example.test/test-manga.jpg', medium: '', color: '' },
      }],
    });
    getMangaDexBrowse.mockResolvedValue({
      result: 'ok',
      data: [{ id: 'mangadex-a', title: 'A MangaDex Title', description: 'A MangaDex description.', coverUrl: 'https://example.test/mangadex.jpg' }],
      total: 92880,
      limit: 24,
      offset: 0,
    });
    mockedBackend.mockImplementation(() => ({ searchAniListManga, getAniListReadingSource: jest.fn(), getMangaDexBrowse }));
  });

  it('loads MangaDex browse results with compact direct page navigation', async () => {
    render(<MangaSearchPage />, { wrapper: MemoryRouter });

    expect(await screen.findByRole('img', { name: 'Cover for A MangaDex Title' })).toBeInTheDocument();
    expect(getMangaDexBrowse).toHaveBeenCalledWith(24, 0);
    expect(screen.getByRole('button', { name: 'Page 1' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Page 3870' })).toBeInTheDocument();

    for (let step = 0; step < 5; step += 1) await userEvent.click(screen.getByRole('button', { name: 'Next page' }));

    await waitFor(() => expect(getMangaDexBrowse).toHaveBeenLastCalledWith(24, 120));
    expect(screen.getByRole('button', { name: 'Page 4' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Page 5' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Page 6' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('button', { name: 'Page 7' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Page 8' })).toBeInTheDocument();
    expect(screen.getAllByText('…')).toHaveLength(2);
  });

  it('renders the cover returned for a search result', async () => {
    render(<MangaSearchPage />, {
      wrapper: ({ children }) => <MemoryRouter initialEntries={['/manga/search?title=test']}>{children}</MemoryRouter>,
    });

    const cover = await screen.findByRole('img', { name: 'Cover for Test Manga' });
    expect(cover).toHaveAttribute('src', 'https://example.test/test-manga.jpg');
    expect(cover.parentElement).toHaveClass('flex-row');
    expect(cover.closest('.col-lg-4')).not.toBeNull();
  });

  it('restores filter-only searches from the URL and sends them to AniList', async () => {
    render(<MangaSearchPage />, {
      wrapper: ({ children }) => <MemoryRouter initialEntries={['/manga/search?tag=Isekai&year=2024&format=MANGA&status=RELEASING&sort=SCORE_DESC']}>{children}</MemoryRouter>,
    });

    expect(await screen.findByRole('img', { name: 'Cover for Test Manga' })).toBeInTheDocument();
    expect(searchAniListManga).toHaveBeenCalledWith({
      title: '', genre: '', tag: 'Isekai', year: 2024, format: 'MANGA', status: 'RELEASING', sort: 'SCORE_DESC', page: 1, perPage: 20,
    });
    expect(screen.queryByLabelText('Genre')).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Filters' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Search filters')).toBeInTheDocument();
    expect(screen.getByLabelText('Genre')).toHaveValue('Isekai');
    expect(screen.getByLabelText('Release year')).toHaveValue(2024);
    expect(screen.getByLabelText('Format')).toHaveValue('MANGA');
    expect(screen.getByLabelText('Publication status')).toHaveValue('RELEASING');
    expect(screen.getByLabelText('Sort by')).toHaveValue('SCORE_DESC');
  });

  it('submits selected filters instead of filtering loaded search cards', async () => {
    render(<MangaSearchPage />, { wrapper: MemoryRouter });

    await userEvent.click(screen.getByRole('button', { name: 'Filters' }));
    await userEvent.selectOptions(screen.getByLabelText('Genre'), 'Magic');
    await userEvent.selectOptions(screen.getByLabelText('Sort by'), 'SCORE_DESC');
    await userEvent.click(screen.getByRole('button', { name: 'Done' }));
    await userEvent.click(screen.getByRole('button', { name: 'Search' }));

    expect(await screen.findByRole('img', { name: 'Cover for Test Manga' })).toBeInTheDocument();
    expect(searchAniListManga).toHaveBeenCalledWith({
      title: '', genre: '', tag: 'Magic', year: undefined, format: '', status: '', sort: 'SCORE_DESC', page: 1, perPage: 20,
    });
  });
});
