import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import MangaSearchPage from './manga-search-page';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('MangaSearchPage covers', () => {
  const searchAniListManga = jest.fn();

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
    mockedBackend.mockImplementation(() => ({ searchAniListManga, getAniListReadingSource: jest.fn() }));
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
