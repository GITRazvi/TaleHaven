import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import GuidedRecommendationChat from './guided-recommendation-chat';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

describe('GuidedRecommendationChat', () => {
  const searchAniListManga = jest.fn();
  const getAniListMyLibrary = jest.fn();
  const getAniListMangaDetails = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockedUseAuth.mockReturnValue({ user: null, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);
    searchAniListManga.mockResolvedValue({ result: 'ok', data: [{ id: 2, title: { english: 'Action Pick', userPreferred: '', romaji: '', native: '' }, description: '', coverImage: { large: '', medium: 'https://example.test/action.jpg', color: '' }, format: 'MANGA', status: 'FINISHED', chapters: 1, volumes: 1, genres: ['Action'] }], page: 1, perPage: 12, total: 1, hasNextPage: false });
    mockedBackend.mockImplementation(() => ({ searchAniListManga, getAniListMyLibrary, getAniListMangaDetails }));
  });

  it('turns a preset genre answer into a score-sorted AniList search', async () => {
    render(<GuidedRecommendationChat />, { wrapper: MemoryRouter });

    await userEvent.click(screen.getByRole('button', { name: 'A genre' }));
    await userEvent.click(screen.getByRole('button', { name: 'Action' }));

    expect(await screen.findByRole('img', { name: 'Cover for Action Pick' })).toBeInTheDocument();
    expect(searchAniListManga).toHaveBeenCalledWith({ genre: 'Action', page: 1, perPage: 12, sort: 'SCORE_DESC' });
  });

  it('omits titles already in the signed-in reader’s library', async () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', role: 'user', displayName: 'Reader', profileImagePath: '', created_at: '' }, loading: false, sessionExpired: false } as ReturnType<typeof useAuth>);
    getAniListMyLibrary.mockResolvedValue({ data: [{ anilistId: 1 }] });
    searchAniListManga.mockResolvedValue({ result: 'ok', data: [
      { id: 1, title: { english: 'Saved Pick', userPreferred: '', romaji: '', native: '' }, description: '', coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: 'FINISHED', chapters: 1, volumes: 1, genres: [] },
      { id: 2, title: { english: 'New Pick', userPreferred: '', romaji: '', native: '' }, description: '', coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: 'FINISHED', chapters: 1, volumes: 1, genres: [] },
    ], page: 1, perPage: 12, total: 2, hasNextPage: false });
    render(<GuidedRecommendationChat />, { wrapper: MemoryRouter });
    await waitFor(() => expect(getAniListMyLibrary).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByRole('button', { name: 'A genre' })).not.toBeDisabled());

    await userEvent.click(screen.getByRole('button', { name: 'A genre' }));
    await userEvent.click(screen.getByRole('button', { name: 'Action' }));

    expect(await screen.findByText('New Pick')).toBeInTheDocument();
    expect(screen.queryByText('Saved Pick')).not.toBeInTheDocument();
  });

  it('turns a typed filter request into a multi-genre AniList search', async () => {
    render(<GuidedRecommendationChat />, { wrapper: MemoryRouter });

    await userEvent.type(screen.getByLabelText('Ask Chatbot'), 'action romance one-shot completed');
    await userEvent.click(screen.getByRole('button', { name: 'Ask' }));

    expect(await screen.findByRole('img', { name: 'Cover for Action Pick' })).toBeInTheDocument();
    expect(searchAniListManga).toHaveBeenCalledWith({ genres: ['Action', 'Romance'], format: 'ONE_SHOT', status: 'FINISHED', page: 1, perPage: 12, sort: 'SCORE_DESC' });
  });

  it('resolves a same-author request then excludes the referenced manga', async () => {
    const source = { id: 7, title: { english: 'Berserk', userPreferred: '', romaji: '', native: '' }, description: '', coverImage: { large: '', medium: '', color: '' }, format: 'MANGA', status: 'FINISHED', chapters: 1, volumes: 1, genres: ['Action'] };
    const otherByAuthor = { ...source, id: 8, title: { ...source.title, english: 'Other Manga' } };
    searchAniListManga.mockReset()
      .mockResolvedValueOnce({ result: 'ok', data: [source], page: 1, perPage: 5, total: 1, hasNextPage: false })
      .mockResolvedValueOnce({ result: 'ok', data: [source, otherByAuthor], page: 1, perPage: 12, total: 2, hasNextPage: false });
    getAniListMangaDetails.mockResolvedValue({ result: 'ok', data: { ...source, staff: { edges: [{ role: 'Story & Art', node: { id: 42, name: { full: 'Author', native: '', userPreferred: 'Author' } } }] } } });
    render(<GuidedRecommendationChat />, { wrapper: MemoryRouter });

    await userEvent.type(screen.getByLabelText('Ask Chatbot'), 'same author as Berserk');
    await userEvent.click(screen.getByRole('button', { name: 'Ask' }));

    expect(await screen.findByText('Other Manga')).toBeInTheDocument();
    expect(screen.queryByText('Berserk')).not.toBeInTheDocument();
    expect(searchAniListManga).toHaveBeenNthCalledWith(1, { title: 'Berserk', page: 1, perPage: 5, sort: 'SEARCH_MATCH' });
    expect(searchAniListManga).toHaveBeenNthCalledWith(2, { staffId: 42, page: 1, perPage: 12, sort: 'SCORE_DESC' });
  });
});
