import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import HomePage from './home-page';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mangaFor = (id: number, title: string) => ({
  id,
  title: { english: title, romaji: title, native: '', userPreferred: title },
  coverImage: { large: '', medium: '', color: '' },
  averageScore: 90,
  description: '',
  format: 'MANGA',
  status: 'FINISHED',
  chapters: 1,
  volumes: 1,
  genres: [],
  popularity: 1,
  siteUrl: '',
  isAdult: false,
});

describe('HomePage AniList discovery', () => {
  const getAniListDiscoverySection = jest.fn();
  const getPopularManga = jest.fn();
  const getAniListMyLibrary = jest.fn();
  const getMyLibrary = jest.fn();
  const getReadingProgress = jest.fn();
  const getChapter = jest.fn();
  const getMangaChapters = jest.fn();
  const addMangaToMyLibrary = jest.fn();
  const removeMangaFromMyLibrary = jest.fn();
  const getAniListFavoriteIDs = jest.fn();
  const getMangaDexFavoriteIDs = jest.fn();
  const addMangaToFavorites = jest.fn();
  const removeMangaFromFavorites = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    window.localStorage.clear();
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    getAniListDiscoverySection.mockImplementation((section: string) => Promise.resolve({
      data: [mangaFor(section.length, `${section} manga`)],
      page: 1,
      perPage: 12,
      total: 1,
      hasNextPage: false,
    }));
    getPopularManga.mockResolvedValue({
      data: [{ id: 'read-now-1', title: 'readable manga', description: '', coverUrl: '', authors: [], artists: [], status: 'ongoing', year: 0, tags: [] }],
      total: 1,
      limit: 12,
      offset: 0,
    });
    getAniListMyLibrary.mockResolvedValue({ data: [] });
    getMyLibrary.mockResolvedValue({ data: [] });
    getReadingProgress.mockResolvedValue({ data: { mangas: {} } });
    addMangaToMyLibrary.mockResolvedValue({ added: true });
    removeMangaFromMyLibrary.mockResolvedValue({ removed: true });
    getAniListFavoriteIDs.mockResolvedValue({ data: [] });
    getMangaDexFavoriteIDs.mockResolvedValue({ data: [] });
    addMangaToFavorites.mockResolvedValue({ added: true });
    removeMangaFromFavorites.mockResolvedValue({ removed: true });
    mockedBackend.mockImplementation(() => ({ getAniListDiscoverySection, getPopularManga, getAniListMyLibrary, getMyLibrary, getReadingProgress, getChapter, getMangaChapters, addMangaToMyLibrary, removeMangaFromMyLibrary, getAniListFavoriteIDs, getMangaDexFavoriteIDs, addMangaToFavorites, removeMangaFromFavorites }));
  });

  it('loads and renders all four AniList discovery sections', async () => {
    render(<HomePage />, { wrapper: MemoryRouter });

    expect(await screen.findByRole('heading', { name: 'Trending Now' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'All-Time Popular' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Top Rated' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'New Releases' })).toBeInTheDocument();

    await waitFor(() => {
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('trending', 1, 12);
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('popular', 1, 12);
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('top-rated', 1, 12);
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('recent', 1, 12);
    });
    expect(await screen.findByText('trending manga')).toBeInTheDocument();
    expect(await screen.findByText('popular manga')).toBeInTheDocument();
    expect(await screen.findByText('readable manga')).toBeInTheDocument();
    expect(getPopularManga).toHaveBeenCalledWith({ limit: 12, offset: 0, orderBy: 'followedCount' });
    expect(getReadingProgress).not.toHaveBeenCalled();
  });

  it('loads the next page for only the requested section', async () => {
    getAniListDiscoverySection.mockImplementation((section: string, page: number) => Promise.resolve({
      data: [mangaFor(page === 1 ? 1 : 2, `${section} manga ${page}`)],
      page,
      perPage: 12,
      total: 2,
      hasNextPage: page === 1,
    }));
    render(<HomePage />, { wrapper: MemoryRouter });

    await screen.findByText('trending manga 1');
    const trendingSection = screen.getByRole('heading', { name: 'Trending Now' }).closest('section');
    if (!trendingSection) {
      throw new Error('Trending section was not rendered');
    }
    await userEvent.click(within(trendingSection).getByRole('button', { name: 'Load more' }));

    await waitFor(() => {
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('trending', 2, 12);
      expect(screen.getByText('trending manga 2')).toBeInTheDocument();
    });
  });

  it('loads another MangaDex page only when its Read on TaleHaven section asks for it', async () => {
    getPopularManga.mockImplementation(({ offset }: { offset: number }) => Promise.resolve({
      data: [{ id: `read-now-${offset + 1}`, title: `readable manga ${offset + 1}`, description: '', coverUrl: '', authors: [], artists: [], status: 'ongoing', year: 0, tags: [] }],
      total: 2,
      limit: 12,
      offset,
    }));
    render(<HomePage />, { wrapper: MemoryRouter });

    await screen.findByText('readable manga 1');
    const readNowSection = screen.getByRole('heading', { name: 'Read on TaleHaven' }).closest('section');
    if (!readNowSection) {
      throw new Error('Read on TaleHaven section was not rendered');
    }
    await userEvent.click(within(readNowSection).getByRole('button', { name: 'Load more' }));

    await waitFor(() => {
      expect(getPopularManga).toHaveBeenCalledWith({ limit: 12, offset: 1, orderBy: 'followedCount' });
      expect(screen.getByText('readable manga 2')).toBeInTheDocument();
    });
  });

  it('shows and toggles the MangaDex library heart for signed-in readers', async () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.com', created_at: '' } } as ReturnType<typeof useAuth>);
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'read-now-1', title: 'readable manga', coverUrl: '', addedAt: '' }] });
    render(<HomePage />, { wrapper: MemoryRouter });

    const heart = await screen.findByRole('button', { name: 'Remove readable manga from My Library' });
    await userEvent.click(heart);

    await waitFor(() => {
      expect(removeMangaFromMyLibrary).toHaveBeenCalledWith('read-now-1');
      expect(screen.getByRole('button', { name: 'Add readable manga to My Library' })).toBeInTheDocument();
    });
  });

  it('shows the MangaDex favorite heart only after the title is saved to My Library', async () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.com', created_at: '' } } as ReturnType<typeof useAuth>);
    getMangaDexFavoriteIDs.mockResolvedValue({ data: ['read-now-1'] });
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'read-now-1', title: 'readable manga', coverUrl: '', addedAt: '' }] });
    render(<HomePage />, { wrapper: MemoryRouter });

    const heart = await screen.findByRole('button', { name: 'Remove readable manga from Favorites' });
    await userEvent.click(heart);

    await waitFor(() => {
      expect(removeMangaFromFavorites).toHaveBeenCalledWith('read-now-1');
      expect(screen.getByRole('button', { name: 'Add readable manga to Favorites' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Remove readable manga from My Library' })).toBeInTheDocument();
    });
  });

  it('hides the MangaDex favorite heart until the title is in My Library', async () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.com', created_at: '' } } as ReturnType<typeof useAuth>);
    getMangaDexFavoriteIDs.mockResolvedValue({ data: ['read-now-1'] });
    render(<HomePage />, { wrapper: MemoryRouter });

    await screen.findByText('readable manga');
    expect(screen.queryByRole('button', { name: 'Remove readable manga from Favorites' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add readable manga to Favorites' })).not.toBeInTheDocument();
  });

  it('shows saved titles only after a completed chapter and includes AniList Plan to Read titles', async () => {
    mockedUseAuth.mockReturnValue({
      user: { id: 1, email: 'reader@example.com', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '' },
    } as ReturnType<typeof useAuth>);
    getMyLibrary.mockResolvedValue({
      data: [
        { mangaId: 'started-manga', title: 'Started manga', coverUrl: '', addedAt: '' },
        { mangaId: 'unfinished-manga', title: 'Unfinished manga', coverUrl: '', addedAt: '' },
      ],
    });
    getAniListMyLibrary.mockResolvedValue({
      data: [{ anilistId: 99, readingStatus: 'PLAN_TO_READ', personalRating: null, addedAt: '', manga: mangaFor(99, 'Planned manga') }],
    });
    getReadingProgress.mockResolvedValue({
      data: {
        mangas: {
          'started-manga': {
            mangaId: 'started-manga',
            updatedAt: '2026-10-26T12:00:00Z',
            chapters: {
              completed: { currentPage: 12, totalPages: 12, lastReadAt: '2026-10-25T12:00:00Z' },
              current: { currentPage: 4, totalPages: 12, lastReadAt: '2026-10-26T12:00:00Z' },
            },
          },
          'unfinished-manga': {
            mangaId: 'unfinished-manga',
            updatedAt: '2026-10-26T12:00:00Z',
            chapters: {
              current: { currentPage: 2, totalPages: 10, lastReadAt: '2026-10-26T12:00:00Z' },
            },
          },
        },
      },
    });
    getChapter.mockResolvedValue({
      data: { id: 'current', chapter: '2', translatedLanguage: 'en', scanlationGroupId: 'reader-group' },
    });
    getMangaChapters.mockResolvedValue({
      data: [
        { id: 'current', chapter: '2' },
        { id: 'same-group-next', chapter: '3' },
      ],
      total: 2,
      limit: 100,
      offset: 0,
    });

    render(<HomePage />, { wrapper: MemoryRouter });

    expect(await screen.findByRole('heading', { name: 'Continue Reading' })).toBeInTheDocument();
    expect(await screen.findByText('Started manga')).toBeInTheDocument();
    expect(screen.getByText('Chapter in progress · Page 4 of 12')).toBeInTheDocument();
    expect(screen.getByText('Unread Chapter 3 in your last scanlation group')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Resume' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Plan to Read' })).toBeInTheDocument();
    expect(screen.getByText('Planned manga')).toBeInTheDocument();
    expect(screen.queryByText('Unfinished manga')).not.toBeInTheDocument();
    expect(getReadingProgress).toHaveBeenCalledTimes(1);
    expect(getMangaChapters).toHaveBeenCalledWith('started-manga', {
      limit: 100,
      offset: 0,
      language: 'en',
      scanlationGroup: 'reader-group',
    });
  });

  it('does not recommend a fully completed title when its scanlation group has no newer chapter', async () => {
    mockedUseAuth.mockReturnValue({
      user: { id: 1, email: 'reader@example.com', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '' },
    } as ReturnType<typeof useAuth>);
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'finished-manga', title: 'Finished manga', coverUrl: '', addedAt: '' }] });
    getAniListMyLibrary.mockResolvedValue({ data: [] });
    getReadingProgress.mockResolvedValue({
      data: {
        mangas: {
          'finished-manga': {
            mangaId: 'finished-manga', updatedAt: '2026-10-26T12:00:00Z',
            chapters: { finished: { currentPage: 20, totalPages: 20, lastReadAt: '2026-10-26T12:00:00Z' } },
          },
        },
      },
    });
    getChapter.mockResolvedValue({ data: { id: 'finished', chapter: '7', translatedLanguage: 'en', scanlationGroupId: 'reader-group' } });
    getMangaChapters.mockResolvedValue({
      data: [{ id: 'finished', chapter: '7' }], total: 1, limit: 100, offset: 0,
    });

    render(<HomePage />, { wrapper: MemoryRouter });

    expect(await screen.findByText('Save a manga and finish a chapter to see it here, or add one to Plan to Read.')).toBeInTheDocument();
    expect(screen.queryByText('Finished manga')).not.toBeInTheDocument();
  });

  it('recommends a skipped unread chapter from the last-used scanlation group', async () => {
    mockedUseAuth.mockReturnValue({
      user: { id: 1, email: 'reader@example.com', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '' },
    } as ReturnType<typeof useAuth>);
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'gapped-manga', title: 'Gapped manga', coverUrl: '', addedAt: '' }] });
    getAniListMyLibrary.mockResolvedValue({ data: [] });
    getReadingProgress.mockResolvedValue({
      data: {
        mangas: {
          'gapped-manga': {
            mangaId: 'gapped-manga', updatedAt: '2026-10-26T12:00:00Z',
            chapters: {
              one: { currentPage: 10, totalPages: 10, lastReadAt: '2026-10-25T12:00:00Z' },
              three: { currentPage: 10, totalPages: 10, lastReadAt: '2026-10-26T12:00:00Z' },
            },
          },
        },
      },
    });
    getChapter.mockResolvedValue({ data: { id: 'three', chapter: '3', translatedLanguage: 'en', scanlationGroupId: 'reader-group' } });
    getMangaChapters.mockResolvedValue({
      data: [{ id: 'one', chapter: '1' }, { id: 'two', chapter: '2' }, { id: 'three', chapter: '3' }],
      total: 3, limit: 100, offset: 0,
    });

    render(<HomePage />, { wrapper: MemoryRouter });

    expect(await screen.findByText('Gapped manga')).toBeInTheDocument();
    expect(screen.getByText('Unread Chapter 2 in your last scanlation group')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Read Chapter 2' })).toBeInTheDocument();
  });

  it('keeps cached AniList discovery results visible when AniList is unavailable', async () => {
    window.localStorage.setItem('talehaven:anilist:discovery:trending', JSON.stringify({
      savedAt: Date.now(),
      value: {
        manga: [mangaFor(99, 'cached trending manga')],
        page: 1,
        hasNextPage: false,
      },
    }));
    getAniListDiscoverySection.mockRejectedValue(new Error('AniList rate limit reached'));

    render(<HomePage />, { wrapper: MemoryRouter });

    const trendingSection = screen.getByRole('heading', { name: 'Trending Now' }).closest('section');
    if (!trendingSection) {
      throw new Error('Trending section was not rendered');
    }

    expect(await within(trendingSection).findByText('cached trending manga')).toBeInTheDocument();
    expect(await within(trendingSection).findByText(/Showing saved trending now from this browser/)).toBeInTheDocument();
    await userEvent.click(within(trendingSection).getByRole('button', { name: 'Retry' }));
    await waitFor(() => {
      expect(getAniListDiscoverySection).toHaveBeenCalledWith('trending', 1, 12);
      expect(getAniListDiscoverySection).toHaveBeenCalledTimes(5);
    });
  });
});
