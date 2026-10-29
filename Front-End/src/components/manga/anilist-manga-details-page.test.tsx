import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import AniListMangaDetailsPage from './anilist-manga-details-page';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mangaResponse = {
  data: {
    id: 1, title: { english: 'Fantasy title', romaji: '', native: '', userPreferred: '' },
    coverImage: { large: '', medium: '', color: '' }, averageScore: 0, description: '', format: 'MANGA', status: '', chapters: 0, volumes: 0,
    genres: ['Fantasy'], popularity: 0, siteUrl: '', isAdult: false,
    characters: { pageInfo: { currentPage: 1, hasNextPage: false, perPage: 10, total: 2 }, edges: [
      { role: 'MAIN', node: { id: 10, name: { full: 'Main Hero', native: '', userPreferred: '' }, image: { large: 'https://example.test/main-hero.jpg', medium: '' } } },
      { role: 'SUPPORTING', node: { id: 11, name: { full: 'Helpful Friend', native: '', userPreferred: '' }, image: { large: '', medium: '' } } },
    ] },
  },
};

const emptyGenreRecommendations = () => jest.fn().mockResolvedValue({ data: [] });

describe('AniListMangaDetailsPage', () => {
  it('orders main characters before supporting and other roles', async () => {
    const orderedCharactersResponse = {
      data: {
        ...mangaResponse.data,
        characters: {
          ...mangaResponse.data.characters,
          edges: [
            { role: 'SUPPORTING', node: { id: 11, name: { full: 'Helpful Friend', native: '', userPreferred: '' }, image: { large: '', medium: '' } } },
            { role: 'BACKGROUND', node: { id: 12, name: { full: 'Background Guest', native: '', userPreferred: '' }, image: { large: '', medium: '' } } },
            { role: 'MAIN', node: { id: 10, name: { full: 'Main Hero', native: '', userPreferred: '' }, image: { large: '', medium: '' } } },
          ],
        },
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(orderedCharactersResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    await screen.findByRole('heading', { name: 'Characters' });
    expect(screen.getAllByRole('button', { name: /View .* details/ }).map((button) => button.getAttribute('aria-label'))).toEqual([
      'View Main Hero details',
      'View Helpful Friend details',
      'View Background Guest details',
    ]);
  });

  it('shows distinct alternative titles without repeating the primary title', async () => {
    const alternateTitlesResponse = {
      data: {
        ...mangaResponse.data,
        title: { english: 'Primary title', romaji: 'Romanized title', native: 'Native title', userPreferred: 'Primary title' },
        synonyms: ['Romanized title', 'Another known title', 'Primary title'],
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(alternateTitlesResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    expect((await screen.findByText('Alternative titles:')).parentElement).toHaveTextContent('Alternative titles: Romanized title · Native title · Another known title');
  });

  it('shows standard genres and searchable AniList tags together', async () => {
    const taggedMangaResponse = {
      data: {
        ...mangaResponse.data,
        startDate: { year: 2024, month: 4, day: 3 },
        format: 'MANGA',
        status: 'RELEASING',
        volumes: 0,
        chapters: 12,
        averageScore: 87,
        popularity: 97822,
        favourites: 4579,
        rankings: [
          { rank: 802, type: 'RATED', allTime: true, context: 'highest rated all time' },
          { rank: 181, type: 'POPULAR', allTime: true, context: 'most popular all time' },
        ],
        staff: { edges: [{ role: 'Story & Art', node: { id: 50, name: { full: 'Creator Name', native: '', userPreferred: '' } } }] },
        tags: [
          { name: 'Isekai', rank: 92, isMediaSpoiler: false },
          { name: 'Time Skip', rank: 90, isMediaSpoiler: false },
          { name: 'Hidden reveal', rank: 99, isMediaSpoiler: true },
        ],
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(taggedMangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    const information = await screen.findByLabelText('Manga information');
    expect(information).toHaveTextContent('Fantasy title');
    expect(information).toHaveTextContent('Release date: 2024-04-03');
    expect(information).toHaveTextContent('TypeManga');
    expect(information).toHaveTextContent('VolumesUnknown');
    expect(information).toHaveTextContent('Chapters12');
    expect(information).toHaveTextContent('StatusPublishing');
    expect(information).toHaveTextContent('AuthorsCreator Name (Story & Art)');
    const genres = screen.getByLabelText('Genres');
    const statistics = screen.getByRole('heading', { name: 'Statistics' }).parentElement;
    expect(genres.compareDocumentPosition(statistics!) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(statistics).toHaveTextContent('Score8.7/10');
    expect(statistics).toHaveTextContent('Ranked#802');
    expect(statistics).toHaveTextContent('Popularity#181');
    expect(statistics).toHaveTextContent('Members97,822');
    expect(statistics).toHaveTextContent('Favorites4,579');
    expect(genres).toHaveTextContent('Fantasy');
    expect(genres).toHaveTextContent('Isekai');
    expect(screen.getByRole('link', { name: 'Isekai' })).toHaveAttribute('href', '/manga/search?tag=Isekai');
    expect(screen.getByRole('link', { name: 'Fantasy' })).toHaveAttribute('href', '/anilist/genre/Fantasy');
    expect(screen.queryByText('Time Skip')).not.toBeInTheDocument();
    expect(screen.queryByText('Hidden reveal')).not.toBeInTheDocument();
  });

  it('shows only manga relations regardless of their chapter or volume totals', async () => {
    const relatedMangaResponse = {
      data: {
        ...mangaResponse.data,
        relations: {
          edges: [
            { relationType: 'SEQUEL', node: { id: 2, type: 'MANGA', title: { english: 'Long Sequel', native: '', romaji: '', userPreferred: '' }, coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: 'RELEASING', siteUrl: '' } },
            { relationType: 'ADAPTATION', node: { id: 3, type: 'ANIME', title: { english: 'Anime Adaptation', native: '', romaji: '', userPreferred: '' }, coverImage: { large: '', medium: '', color: '' }, format: 'TV', status: 'FINISHED', siteUrl: '' } },
          ],
        },
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(relatedMangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: 'Related manga' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Open Long Sequel' })).toHaveAttribute('href', '/anilist/manga/2');
    expect(screen.getByText('Sequel')).toBeInTheDocument();
    expect(screen.queryByText('Anime Adaptation')).not.toBeInTheDocument();
  });

  it('loads another ten characters only when the reader requests them', async () => {
    const getAniListMangaCharacters = jest.fn().mockResolvedValue({
      data: {
        pageInfo: { currentPage: 2, hasNextPage: false, perPage: 10, total: 11 },
        edges: [{ role: 'SUPPORTING', node: { id: 12, name: { full: 'Loaded Character', native: '', userPreferred: '' }, image: { large: '', medium: '' } } }],
      },
    });
    const pagedCharacterResponse = {
      data: {
        ...mangaResponse.data,
        characters: {
          ...mangaResponse.data.characters,
          pageInfo: { currentPage: 1, hasNextPage: true, perPage: 10, total: 11 },
        },
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(pagedCharacterResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaCharacters,
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    await userEvent.click(await screen.findByRole('button', { name: 'Load more characters' }));
    expect(getAniListMangaCharacters).toHaveBeenCalledWith(1, 2);
    expect(await screen.findByText('Loaded Character')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Load more characters' })).not.toBeInTheDocument();
  });

  it('shows valid AniList streaming links in the Read on section', async () => {
    const externalLinksResponse = {
      data: {
        ...mangaResponse.data,
        externalLinks: [
          { id: 1, site: 'Official Reader', url: 'https://reader.example.test/series', type: 'STREAMING', language: 'English' },
          { id: 2, site: 'Information only', url: 'https://info.example.test/series', type: 'INFO', language: 'English' },
          { id: 3, site: 'Unsafe Reader', url: 'javascript:alert(1)', type: 'STREAMING', language: 'English' },
        ],
      },
    };
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(externalLinksResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: 'Read on' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Read Fantasy title on Official Reader (English)' })).toHaveAttribute('href', 'https://reader.example.test/series');
    expect(screen.queryByText('Information only')).not.toBeInTheDocument();
    expect(screen.queryByText('Unsafe Reader')).not.toBeInTheDocument();
    expect(screen.queryByText('Chapters are not currently available on TaleHaven.')).not.toBeInTheDocument();
  });

  it('shows top-rated manga from the current title\'s first genre', async () => {
    const getAniListMangaByGenre = jest.fn().mockResolvedValue({
      data: [
        mangaResponse.data,
        { ...mangaResponse.data, id: 2, title: { english: 'Top Fantasy Pick', native: '', romaji: '', userPreferred: '' }, averageScore: 95 },
      ],
    });
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(mangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre,
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    const recommendationsHeading = await screen.findByRole('heading', { name: 'Top-rated in Fantasy' });
    expect(recommendationsHeading).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Characters' }).compareDocumentPosition(recommendationsHeading) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(getAniListMangaByGenre).toHaveBeenCalledWith('Fantasy', 1, 7);
    expect(await screen.findByRole('link', { name: 'Open Top Fantasy Pick' })).toHaveAttribute('href', '/anilist/manga/2');
    expect(screen.queryByRole('link', { name: 'Open Fantasy title' })).not.toBeInTheDocument();
  });

  it('links each genre directly to its genre results page', async () => {
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(mangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /><Route path="/anilist/genre/:genre" element={<div>Genre results</div>} /></Routes></MemoryRouter>);

    await userEvent.click(await screen.findByRole('link', { name: 'Fantasy' }));
    expect(await screen.findByText('Genre results')).toBeInTheDocument();
  });

  it('shows and updates reading settings on the metadata page for a saved manga', async () => {
    const getAniListMyLibrary = jest.fn().mockResolvedValue({ data: [{ anilistId: 1, addedAt: '', readingStatus: 'PLAN_TO_READ', personalRating: null, manga: mangaResponse.data }] });
    const getAniListCharacterFavorites = jest.fn().mockResolvedValue({ data: [] });
    const addAniListCharacterToFavorites = jest.fn().mockResolvedValue({ added: true });
    const getAniListCharacterDetails = jest.fn().mockResolvedValue({ data: { id: 10, name: { full: 'Main Hero', native: '', userPreferred: '' }, image: { large: 'https://example.test/main-hero.jpg', medium: '' }, description: 'A heroic lead.', gender: 'Male', dateOfBirth: { year: 2000, month: 4, day: 2 }, age: '18', bloodType: 'O', siteUrl: 'https://anilist.co/character/10', favourites: 12, media: { pageInfo: { currentPage: 1, hasNextPage: false, perPage: 25, total: 1 }, edges: [{ characterRole: 'MAIN', node: { id: 1, type: 'MANGA', title: { english: 'Fantasy title', native: '', romaji: '', userPreferred: '' }, coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: '', siteUrl: '' } }] } } });
    const updateAniListMangaReadingStatus = jest.fn().mockResolvedValue({ readingStatus: 'CURRENTLY_READING' });
    const updateAniListMangaPersonalRating = jest.fn().mockResolvedValue({ personalRating: 9 });
    const getAniListReadingSource = jest.fn().mockResolvedValue({ available: true, mangaId: 'manga-1' });
    const getAniListFavoriteIDs = jest.fn().mockResolvedValue({ data: [] });
    const addAniListMangaToFavorites = jest.fn().mockResolvedValue({ added: true });
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({ getAniListMangaDetails: jest.fn().mockResolvedValue(mangaResponse), getAniListReadingSource, getAniListMangaByGenre: emptyGenreRecommendations(), getAniListMyLibrary, getAniListFavoriteIDs, addAniListMangaToFavorites, getAniListCharacterFavorites, getAniListCharacterDetails, addAniListCharacterToFavorites, updateAniListMangaReadingStatus, updateAniListMangaPersonalRating, getAniListMangaComments: jest.fn().mockResolvedValue({ data: [] }) }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    await screen.findByRole('heading', { name: 'Your library settings' });
    expect(screen.getByRole('button', { name: 'Remove Fantasy title from My Library' })).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Add Fantasy title to Favorites' }));
    await waitFor(() => expect(addAniListMangaToFavorites).toHaveBeenCalledWith(1));
    expect(await screen.findByText('Chapters available on MangaDex')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Characters' })).toBeInTheDocument();
    expect(screen.getByText('Main Hero')).toBeInTheDocument();
    expect(screen.getByText('Main')).toBeInTheDocument();
    expect(screen.getByText('Helpful Friend')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Main Hero' })).toHaveAttribute('src', 'https://example.test/main-hero.jpg');
    await userEvent.click(screen.getByRole('button', { name: 'Add Main Hero to favorite characters' }));
    await waitFor(() => expect(addAniListCharacterToFavorites).toHaveBeenCalledWith(10, 1, 'MAIN'));
    expect(await screen.findByRole('button', { name: 'Remove Main Hero from favorite characters' })).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'View Main Hero details' }));
    expect(await screen.findByText('A heroic lead.')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Appears in' })).toBeInTheDocument();
    expect(screen.getAllByText('Fantasy title').length).toBeGreaterThan(1);
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Reading status' }), 'CURRENTLY_READING');
    await waitFor(() => expect(updateAniListMangaReadingStatus).toHaveBeenCalledWith(1, 'CURRENTLY_READING'));
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Your rating' }), '9');
    await waitFor(() => expect(updateAniListMangaPersonalRating).toHaveBeenCalledWith(1, 9));
  });

  it('shows the favorite control only after adding a manga to the library', async () => {
    const addAniListMangaToMyLibrary = jest.fn().mockResolvedValue({ added: true });
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(mangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
      getAniListMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getAniListFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
      getAniListCharacterFavorites: jest.fn().mockResolvedValue({ data: [] }),
      getAniListMangaComments: jest.fn().mockResolvedValue({ data: [] }),
      addAniListMangaToMyLibrary,
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    const addButton = await screen.findByRole('button', { name: 'Add Fantasy title to My Library' });
    expect(screen.queryByRole('button', { name: 'Add Fantasy title to Favorites' })).not.toBeInTheDocument();
    await userEvent.click(addButton);
    await waitFor(() => expect(addAniListMangaToMyLibrary).toHaveBeenCalledWith(1));
    expect(await screen.findByRole('button', { name: 'Add Fantasy title to Favorites' })).toBeInTheDocument();
  });

  it('shows manga comments with links to the authors public profiles and only lets the author delete their own', async () => {
    const createAniListMangaComment = jest.fn().mockResolvedValue({ comment: { id: 3, anilistId: 1, userId: 1, authorDisplayName: 'Reader', profileImagePath: '', body: 'A new thought', createdAt: '2026-10-28T00:00:00Z' } });
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getAniListMangaDetails: jest.fn().mockResolvedValue(mangaResponse),
      getAniListReadingSource: jest.fn().mockResolvedValue({ available: false }),
      getAniListMangaByGenre: emptyGenreRecommendations(),
      getAniListMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getAniListFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
      getAniListCharacterFavorites: jest.fn().mockResolvedValue({ data: [] }),
      getAniListMangaComments: jest.fn().mockResolvedValue({ data: [
        { id: 1, anilistId: 1, userId: 2, authorDisplayName: 'Friend', profileImagePath: '', body: 'Great chapter.', createdAt: '2026-10-27T00:00:00Z' },
        { id: 2, anilistId: 1, userId: 1, authorDisplayName: 'Reader', profileImagePath: '', body: 'My note.', createdAt: '2026-10-27T01:00:00Z' },
      ] }),
      createAniListMangaComment,
      deleteAniListMangaComment: jest.fn().mockResolvedValue({ result: 'ok' }),
    }));

    render(<MemoryRouter initialEntries={['/anilist/manga/1']}><Routes><Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} /></Routes></MemoryRouter>);

    expect((await screen.findByRole('link', { name: 'Friend' }))).toHaveAttribute('href', '/profile/2');
    expect(screen.getByRole('link', { name: 'Reader' })).toHaveAttribute('href', '/profile/1');
    expect(screen.getAllByRole('button', { name: 'Delete' })).toHaveLength(1);
    await userEvent.type(screen.getByRole('textbox', { name: 'Share your thoughts' }), 'A new thought');
    await userEvent.keyboard('{Enter}');
    await waitFor(() => expect(createAniListMangaComment).toHaveBeenCalledWith(1, 'A new thought'));
    expect(await screen.findByText('A new thought')).toBeInTheDocument();
  });
});
