import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import AniListGenrePage from './anilist-genre-page';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

const mangaFor = (id: number, title: string) => ({
  id,
  title: { english: title, romaji: title, native: '', userPreferred: title },
  coverImage: { large: '', medium: '', color: '' },
  averageScore: 90,
  description: '', format: 'MANGA', status: 'FINISHED', chapters: 1, volumes: 1,
  genres: ['Fantasy'], popularity: 1, siteUrl: '', isAdult: false,
});

describe('AniListGenrePage', () => {
  const getAniListMangaByGenre = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockedUseAuth.mockReturnValue({ user: null } as ReturnType<typeof useAuth>);
    getAniListMangaByGenre.mockImplementation((genre: string, page: number) => Promise.resolve({
      data: [mangaFor(page, `${genre} manga ${page}`)], page, perPage: 12, total: 2, hasNextPage: page === 1,
    }));
    mockedBackend.mockImplementation(() => ({ getAniListMangaByGenre }));
  });

  it('loads the selected genre and appends only its next results page', async () => {
    render(<MemoryRouter initialEntries={['/anilist/genre/Fantasy']}><Routes><Route path="/anilist/genre/:genre" element={<AniListGenrePage />} /></Routes></MemoryRouter>);

    expect(await screen.findByRole('heading', { name: 'Manga in Fantasy' })).toBeInTheDocument();
    await waitFor(() => expect(getAniListMangaByGenre).toHaveBeenCalledWith('Fantasy', 1, 12));
    expect(await screen.findByText('Fantasy manga 1')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Load more' }));
    await waitFor(() => {
      expect(getAniListMangaByGenre).toHaveBeenCalledWith('Fantasy', 2, 12);
      expect(screen.getByText('Fantasy manga 2')).toBeInTheDocument();
    });
  });
});
