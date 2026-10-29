import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import LibraryPage from './library-page';

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

beforeEach(() => {
  window.localStorage.clear();
});

it('groups saved AniList titles in My Profile and updates their book-level status and rating', async () => {
	const getAniListMyLibrary = jest.fn().mockResolvedValue({ total: 2, stats: { total: 2, byStatus: { PLAN_TO_READ: 0, CURRENTLY_READING: 1, ON_HOLD: 0, READ: 0, DROPPED: 1 }, topGenres: ['Fantasy'], chaptersRead: 3, lastMangaCompleted: { anilistId: 1, title: 'Last Completed Title', coverUrl: 'https://example.test/last-completed-cover.jpg' } }, data: [{ anilistId: 1, addedAt: '2026-10-22T00:00:00Z', readingStatus: 'CURRENTLY_READING', personalRating: null, manga: { title: { english: 'Saved Title', userPreferred: '', romaji: '', native: '' }, coverImage: { large: '', medium: '', color: '' }, averageScore: 80 } }, { anilistId: 2, addedAt: '2026-10-22T00:00:00Z', readingStatus: 'DROPPED', personalRating: null, manga: { title: { english: 'Dropped Title', userPreferred: '', romaji: '', native: '' }, coverImage: { large: '', medium: '', color: '' }, averageScore: 70 } }] });
  const removeAniListMangaFromMyLibrary = jest.fn().mockResolvedValue({ removed: true });
  const getAniListFavoriteIDs = jest.fn().mockResolvedValue({ data: [1] });
  const addAniListMangaToFavorites = jest.fn().mockResolvedValue({ added: true });
  const removeAniListMangaFromFavorites = jest.fn().mockResolvedValue({ removed: true });
  const getMangaDexFavorites = jest.fn().mockResolvedValue({ data: ['mangadex-favorite-1'], manga: [{ mangaId: 'mangadex-favorite-1', title: 'Readable Favorite', coverUrl: 'https://example.test/readable-favorite-cover.jpg' }] });
  const removeMangaFromFavorites = jest.fn().mockResolvedValue({ removed: true });
  const getAniListCharacterFavorites = jest.fn().mockResolvedValue({ data: [{ characterId: 10, mangaId: 1, role: 'MAIN', name: { full: 'Main Hero', native: '', userPreferred: '' }, image: { large: 'https://example.test/main-hero.jpg', medium: '' } }] });
  const removeAniListCharacterFromFavorites = jest.fn().mockResolvedValue({ removed: true });
  const getAniListCharacterDetails = jest.fn().mockResolvedValue({ data: { id: 10, name: { full: 'Main Hero', native: '', userPreferred: '' }, image: { large: 'https://example.test/main-hero.jpg', medium: '' }, description: 'A heroic lead.', gender: '', dateOfBirth: { year: 0, month: 0, day: 0 }, age: '', bloodType: '', siteUrl: '', favourites: 0, media: { pageInfo: { currentPage: 1, hasNextPage: false, perPage: 25, total: 1 }, edges: [{ characterRole: 'MAIN', node: { id: 1, type: 'MANGA', title: { english: 'Saved Title', native: '', romaji: '', userPreferred: '' }, coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: '', siteUrl: '' } }] } } });
	const getProfileCommentsEnabled = jest.fn().mockResolvedValue({ enabled: true });
	const getProfileComments = jest.fn().mockResolvedValue({ data: [{ id: 30, profileUserId: 1, userId: 2, authorDisplayName: 'Another Reader', profileImagePath: '', body: 'Welcome to TaleHaven!', createdAt: '2026-10-28T00:00:00Z' }] });
	const createProfileComment = jest.fn().mockResolvedValue({ comment: { id: 31, profileUserId: 1, userId: 1, authorDisplayName: 'Reader', profileImagePath: '', body: 'My profile note', createdAt: '2026-10-28T00:00:00Z' } });
	mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '2026-10-22T00:00:00Z' } } as ReturnType<typeof useAuth>);
  mockedBackend.mockImplementation(() => ({ getAniListMyLibrary, removeAniListMangaFromMyLibrary, getAniListFavoriteIDs, addAniListMangaToFavorites, removeAniListMangaFromFavorites, getMangaDexFavorites, removeMangaFromFavorites, getAniListCharacterFavorites, removeAniListCharacterFromFavorites, getAniListCharacterDetails, getProfileCommentsEnabled, getProfileComments, createProfileComment }));

  render(<LibraryPage />, { wrapper: MemoryRouter });

	expect(await screen.findByRole('heading', { name: 'Reader' })).toBeInTheDocument();
	expect(screen.getByText('My Profile')).toBeInTheDocument();
	expect(await screen.findByRole('heading', { name: 'Comments' })).toBeInTheDocument();
	expect(await screen.findByText('Welcome to TaleHaven!')).toBeInTheDocument();
	expect(screen.getByRole('link', { name: "View Another Reader's profile" })).toHaveAttribute('href', '/profile/2');
	await userEvent.type(screen.getByRole('textbox', { name: 'Leave a comment' }), 'My profile note');
	await userEvent.keyboard('{Enter}');
	await waitFor(() => expect(createProfileComment).toHaveBeenCalledWith(1, 'My profile note'));
	expect(await screen.findByRole('heading', { name: 'Currently Reading' })).toBeInTheDocument();
  expect((await screen.findAllByText('Saved Title')).length).toBeGreaterThan(0);
  expect(screen.getByText('2')).toBeInTheDocument();
  expect(screen.getByText('Fantasy')).toBeInTheDocument();
  expect(screen.getByText('Chapters read')).toBeInTheDocument();
  expect(screen.getByText('3')).toBeInTheDocument();
  expect(screen.getByText('Last manga completed')).toBeInTheDocument();
  expect(screen.getByText('Last Completed Title')).toBeInTheDocument();
  expect(screen.getByRole('img', { name: 'Cover of Last Completed Title' })).toHaveAttribute('src', 'https://example.test/last-completed-cover.jpg');
  expect(screen.getByRole('img', { name: 'Manga status distribution' })).toBeInTheDocument();
  expect(getAniListMyLibrary).toHaveBeenCalledWith({ limit: 100, offset: 0 });
  expect(await screen.findByRole('heading', { name: 'Favorites' })).toBeInTheDocument();
  expect(screen.getByRole('heading', { name: 'Manga' })).toBeInTheDocument();
  expect(screen.getByRole('heading', { name: 'Characters' })).toBeInTheDocument();
  expect(screen.getByText('Readable Favorite')).toBeInTheDocument();
  expect(screen.getByRole('img', { name: 'Readable Favorite' })).toHaveAttribute('src', 'https://example.test/readable-favorite-cover.jpg');
  expect(screen.getByRole('img', { name: 'Main Hero' })).toHaveAttribute('src', 'https://example.test/main-hero.jpg');
  await userEvent.click(screen.getByRole('button', { name: 'View Main Hero details' }));
  expect(await screen.findByText('A heroic lead.')).toBeInTheDocument();

  await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Filter books by status' }), 'DROPPED');
  expect(screen.getByText('Dropped Title')).toBeInTheDocument();
  expect(screen.queryByRole('combobox', { name: 'Reading status for Saved Title' })).not.toBeInTheDocument();

  await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Filter books by status' }), 'ALL');

  await userEvent.click(screen.getAllByRole('button', { name: 'Remove Saved Title from My Library' })[0]);
  await waitFor(() => expect(removeAniListMangaFromMyLibrary).toHaveBeenCalledWith(1));

  await userEvent.click(screen.getByRole('button', { name: 'Remove Readable Favorite from Favorites' }));
  await waitFor(() => expect(removeMangaFromFavorites).toHaveBeenCalledWith('mangadex-favorite-1'));
  await userEvent.click(screen.getByRole('button', { name: 'Remove Main Hero from favorite characters' }));
  await waitFor(() => expect(removeAniListCharacterFromFavorites).toHaveBeenCalledWith(10));
});

it('keeps a user-scoped cached profile visible when AniList is temporarily unavailable', async () => {
  window.localStorage.setItem('talehaven:profile:1', JSON.stringify({
    savedAt: Date.now(),
    value: {
      library: [{ anilistId: 7, addedAt: '2026-10-24T00:00:00Z', readingStatus: 'PLAN_TO_READ', personalRating: null, manga: { title: { english: 'Cached Profile Title', userPreferred: '', romaji: '', native: '' }, coverImage: { large: '', medium: '', color: '' }, averageScore: 80 } }],
      stats: { total: 1, byStatus: { PLAN_TO_READ: 1, CURRENTLY_READING: 0, ON_HOLD: 0, READ: 0, DROPPED: 0 }, topGenres: [], chaptersRead: 0, lastMangaCompleted: null },
      favoriteAniListIDs: [],
      mangaDexFavorites: [],
      characterFavorites: [],
    },
  }));
  const getAniListMyLibrary = jest.fn().mockRejectedValue(new Error('AniList rate limit reached'));
  const getAniListFavoriteIDs = jest.fn().mockResolvedValue({ data: [] });
  const getMangaDexFavorites = jest.fn().mockResolvedValue({ manga: [] });
  const getAniListCharacterFavorites = jest.fn().mockResolvedValue({ data: [] });
	const getProfileCommentsEnabled = jest.fn().mockResolvedValue({ enabled: false });
  mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '2026-10-24T00:00:00Z' } } as ReturnType<typeof useAuth>);
  mockedBackend.mockImplementation(() => ({ getAniListMyLibrary, getAniListFavoriteIDs, getMangaDexFavorites, getAniListCharacterFavorites, getProfileCommentsEnabled }));

  render(<LibraryPage />, { wrapper: MemoryRouter });

  expect((await screen.findAllByText('Cached Profile Title')).length).toBeGreaterThan(0);
  expect(screen.getByText(/Showing saved profile results from this browser/)).toBeInTheDocument();
  await userEvent.click(screen.getByRole('button', { name: 'Retry' }));
  await waitFor(() => {
    expect(getAniListMyLibrary).toHaveBeenCalledTimes(2);
  });
});
