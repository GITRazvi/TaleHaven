import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import MangaDetailsPage from './manga-details-page';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

describe('MangaDetailsPage scanlation preferences', () => {
  it('restores and saves a signed-in reader\'s group for the current manga and language', async () => {
    const getPreferredScanlationGroup = jest.fn().mockResolvedValue({ scanlationGroup: 'group-1' });
    const updatePreferredScanlationGroup = jest.fn().mockResolvedValue({ scanlationGroup: 'group-2' });
    const getMangaChapters = jest.fn().mockResolvedValue({
      data: [{ id: 'chapter-1', chapter: '1', title: 'Start', translatedLanguage: 'en', scanlationGroup: 'First Group' }],
      total: 1,
      availableLanguages: ['en'],
      scanlationGroups: [
        { id: 'group-1', name: 'First Group', chapterCount: 1 },
        { id: 'group-2', name: 'Second Group', chapterCount: 1 },
      ],
    });
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getReadingSettings: jest.fn().mockResolvedValue({ preferredTranslationLanguage: 'en' }),
      getPreferredScanlationGroup,
      updatePreferredScanlationGroup,
      getManga: jest.fn().mockResolvedValue({ data: { id: 'manga-1', title: 'Test Manga', description: 'Description' } }),
      getMangaChapters,
      getMangaCover: jest.fn().mockResolvedValue({ coverUrl: '' }),
      getMangaProgress: jest.fn().mockResolvedValue(null),
      getMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getMangaDexFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
    }));

    render(<MemoryRouter initialEntries={['/manga/manga-1']}><Routes><Route path="/manga/:mangaId" element={<MangaDetailsPage />} /></Routes></MemoryRouter>);

    const groupSelector = await screen.findByRole('combobox', { name: 'Scanlation group' });
    await waitFor(() => expect(groupSelector).toHaveValue('group-1'));
    expect(getPreferredScanlationGroup).toHaveBeenCalledWith('manga-1', 'en');
    expect(getMangaChapters).toHaveBeenCalledWith('manga-1', expect.objectContaining({ language: 'en', scanlationGroup: 'group-1' }));

    await userEvent.selectOptions(groupSelector, 'group-2');
    await waitFor(() => expect(updatePreferredScanlationGroup).toHaveBeenCalledWith('manga-1', { language: 'en', scanlationGroup: 'group-2' }));
  });

  it('keeps completed chapters read without showing their final-page tracker', async () => {
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getReadingSettings: jest.fn().mockResolvedValue({ preferredTranslationLanguage: 'en' }),
      getPreferredScanlationGroup: jest.fn().mockResolvedValue({ scanlationGroup: 'all' }),
      updatePreferredScanlationGroup: jest.fn(),
      getManga: jest.fn().mockResolvedValue({ data: { id: 'manga-1', title: 'Test Manga', description: 'Description' } }),
      getMangaChapters: jest.fn().mockResolvedValue({
        data: [{ id: 'chapter-1', chapter: '1', title: 'Finish', translatedLanguage: 'en', scanlationGroup: 'First Group' }],
        total: 1,
        availableLanguages: ['en'],
        scanlationGroups: [],
      }),
      getMangaCover: jest.fn().mockResolvedValue({ coverUrl: '' }),
      getMangaProgress: jest.fn().mockResolvedValue({
        data: { chapters: { 'chapter-1': { currentPage: 38, totalPages: 38, lastReadAt: '2026-10-25T00:00:00Z' } } },
      }),
      getMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getMangaDexFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
    }));

    render(<MemoryRouter initialEntries={['/manga/manga-1']}><Routes><Route path="/manga/:mangaId" element={<MangaDetailsPage />} /></Routes></MemoryRouter>);

    expect(await screen.findByText(/Chapter 1 - Finish/)).toBeInTheDocument();
    expect(screen.queryByText('Last read: Page 38')).not.toBeInTheDocument();
  });

  it('preserves a saved group when its latest chapter response has no current releases', async () => {
    const updatePreferredScanlationGroup = jest.fn();
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getReadingSettings: jest.fn().mockResolvedValue({ preferredTranslationLanguage: 'en' }),
      getPreferredScanlationGroup: jest.fn().mockResolvedValue({ scanlationGroup: 'saved-group' }),
      updatePreferredScanlationGroup,
      getManga: jest.fn().mockResolvedValue({ data: { id: 'manga-1', title: 'Test Manga', description: 'Description' } }),
      getMangaChapters: jest.fn().mockResolvedValue({ data: [], total: 0, availableLanguages: ['en'], scanlationGroups: [] }),
      getMangaCover: jest.fn().mockResolvedValue({ coverUrl: '' }),
      getMangaProgress: jest.fn().mockResolvedValue(null),
      getMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getMangaDexFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
    }));

    render(<MemoryRouter initialEntries={['/manga/manga-1']}><Routes><Route path="/manga/:mangaId" element={<MangaDetailsPage />} /></Routes></MemoryRouter>);

    const groupSelector = await screen.findByRole('combobox', { name: 'Scanlation group' });
    await waitFor(() => expect(groupSelector).toHaveValue('saved-group'));
    expect(screen.getByRole('option', { name: 'Saved group (no chapters currently available) — 0 chapters' })).toBeInTheDocument();
    expect(updatePreferredScanlationGroup).not.toHaveBeenCalled();
  });

  it('does not render the manual completion sentinel as a last-read page', async () => {
    const saveReadingProgress = jest.fn().mockResolvedValue({ result: 'ok' });
    const syncReadingProgress = jest.fn().mockResolvedValue({ result: 'ok' });
    mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    mockedBackend.mockImplementation(() => ({
      getReadingSettings: jest.fn().mockResolvedValue({ preferredTranslationLanguage: 'en' }),
      getPreferredScanlationGroup: jest.fn().mockResolvedValue({ scanlationGroup: 'all' }),
      updatePreferredScanlationGroup: jest.fn(),
      getManga: jest.fn().mockResolvedValue({ data: { id: 'manga-1', title: 'Test Manga', description: 'Description' } }),
      getMangaChapters: jest.fn().mockResolvedValue({
        data: [{ id: 'chapter-1', chapter: '1', title: 'Finish', translatedLanguage: 'en', scanlationGroup: 'First Group' }],
        total: 1, availableLanguages: ['en'], scanlationGroups: [],
      }),
      getMangaCover: jest.fn().mockResolvedValue({ coverUrl: '' }),
      getMangaProgress: jest.fn().mockResolvedValue(null),
      getMyLibrary: jest.fn().mockResolvedValue({ data: [] }),
      getMangaDexFavoriteIDs: jest.fn().mockResolvedValue({ data: [] }),
      saveReadingProgress,
      syncReadingProgress,
    }));

    render(<MemoryRouter initialEntries={['/manga/manga-1']}><Routes><Route path="/manga/:mangaId" element={<MangaDetailsPage />} /></Routes></MemoryRouter>);

    await screen.findByText(/Chapter 1 - Finish/);
    await userEvent.click(screen.getByRole('button', { name: 'Select chapters' }));
    await userEvent.click(screen.getByRole('button', { name: 'Select chapter 1: Finish' }));
    await userEvent.click(screen.getByRole('button', { name: 'Mark as read' }));

    await waitFor(() => expect(saveReadingProgress).toHaveBeenCalledWith({
      mangaId: 'manga-1', chapterId: 'chapter-1', currentPage: 2147483647, totalPages: 2147483647,
    }));
    expect(screen.queryByText('Last read: Page 2147483647')).not.toBeInTheDocument();
  });
});
