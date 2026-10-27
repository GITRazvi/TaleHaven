import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import MangaReaderPage from './manga-reader-page';

const mockedBackend = Backend as unknown as jest.Mock;

class MockEventSource {
  static instances: MockEventSource[] = [];

  readonly addEventListener = jest.fn((event: string, listener: EventListener) => {
    const listeners = this.listeners.get(event) ?? [];
    listeners.push(listener);
    this.listeners.set(event, listeners);
  });
  readonly close = jest.fn();
  onerror: ((event: Event) => void) | null = null;
  private readonly listeners = new Map<string, EventListener[]>();

  constructor(readonly url: string) {
    MockEventSource.instances.push(this);
  }

  emit(event: string, data: unknown) {
    const message = new MessageEvent(event, { data: JSON.stringify(data) });
    this.listeners.get(event)?.forEach((listener) => listener(message));
  }
}

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('MangaReaderPage progress', () => {
  const getManga = jest.fn();
  const getChapter = jest.fn();
  const getMangaChapters = jest.fn();
  const getChapterProgress = jest.fn();
  const getChapterPageCount = jest.fn();
  const getChapterPage = jest.fn();
  const getReadingSettings = jest.fn();
  const getImageUrl = jest.fn();
  const saveReadingProgress = jest.fn();
  const syncReadingProgress = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    MockEventSource.instances = [];
    (global as typeof globalThis & { EventSource: typeof EventSource }).EventSource = MockEventSource as unknown as typeof EventSource;
    Element.prototype.scrollIntoView = jest.fn();

    getManga.mockResolvedValue({ data: { title: 'Test Manga' } });
    getChapter.mockResolvedValue({ data: { chapter: '1' } });
    getMangaChapters.mockResolvedValue({ data: [] });
    getChapterProgress.mockRejectedValue({ status: 404 });
    getChapterPageCount.mockResolvedValue({ totalPages: 2 });
    getChapterPage.mockResolvedValue({ pageNumber: 1, totalPages: 2, pageUrl: '/images/page-1.jpg' });
    getReadingSettings.mockResolvedValue({ preferredTranslationLanguage: 'all', preferredReadingMode: 'vertical' });
    getImageUrl.mockImplementation((path: string) => `http://localhost:3001${path}`);
    saveReadingProgress.mockResolvedValue({ result: 'ok' });
    syncReadingProgress.mockResolvedValue({ result: 'ok' });
    mockedBackend.mockImplementation(() => ({
      getBaseUrl: () => 'http://localhost:3001',
      getManga,
      getChapter,
      getMangaChapters,
      getChapterProgress,
      getChapterPageCount,
      getChapterPage,
      getReadingSettings,
      getImageUrl,
      saveReadingProgress,
      syncReadingProgress,
    }));
  });

  it('saves and synchronizes the visible page before the stream completes', async () => {
    render(
      <MemoryRouter initialEntries={[{ pathname: '/manga/manga-1/chapter/chapter-1', state: { chapterIds: ['chapter-1'], startAtPage: 1 } }]}>
        <Routes>
          <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderPage />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(MockEventSource.instances).toHaveLength(1));
    const stream = MockEventSource.instances[0];

    await act(async () => {
      stream.emit('start', { totalPages: 3, startPage: 1 });
    });
    await screen.findByText('No pages found for this chapter.');

    await act(async () => {
      stream.emit('page', { pageNumber: 1, pageUrl: '/images/page-1.jpg' });
    });

    const page = await screen.findByAltText('Chapter page 1');
    await act(async () => {
      page.dispatchEvent(new Event('load'));
    });
    await waitFor(() => {
      expect(saveReadingProgress).toHaveBeenCalledWith({
        mangaId: 'manga-1',
        chapterId: 'chapter-1',
        currentPage: 1,
        totalPages: 3,
      });
      expect(syncReadingProgress).toHaveBeenCalledWith('chapter-1');
    });
    expect(stream.close).not.toHaveBeenCalled();
  });

  it('disables next chapter at a numbered gap in a selected scanlation group', async () => {
    render(
      <MemoryRouter initialEntries={[{
        pathname: '/manga/manga-1/chapter/chapter-1',
        state: {
          chapterIds: ['chapter-1', 'chapter-3'],
          chapterNavigation: [{ id: 'chapter-1', number: '1' }, { id: 'chapter-3', number: '3' }],
          chapterFilters: { language: 'en', scanlationGroup: 'group-1' },
        },
      }]}>
        <Routes>
          <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderPage />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(MockEventSource.instances).toHaveLength(1));
    await act(async () => {
      MockEventSource.instances[0].emit('start', { totalPages: 1, startPage: 1 });
    });

    expect(await screen.findByRole('button', { name: 'Next Chapter →' })).toBeDisabled();
    expect(getMangaChapters).not.toHaveBeenCalled();
  });

  it('replays a manually completed chapter from page 1 without overwriting completion', async () => {
    getChapterProgress.mockResolvedValue({ data: { currentPage: 2147483647, totalPages: 2147483647 } });
    render(
      <MemoryRouter initialEntries={['/manga/manga-1/chapter/chapter-1']}>
        <Routes>
          <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderPage />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(MockEventSource.instances).toHaveLength(1));
    const stream = MockEventSource.instances[0];
    expect(stream.url).toContain('startPage=1');

    await act(async () => {
      stream.emit('start', { totalPages: 3, startPage: 1 });
    });
    await screen.findByText('No pages found for this chapter.');
    await act(async () => {
      stream.emit('page', { pageNumber: 1, pageUrl: '/images/page-1.jpg' });
    });
    const page = await screen.findByAltText('Chapter page 1');
    await act(async () => {
      page.dispatchEvent(new Event('load'));
    });

    expect(saveReadingProgress).not.toHaveBeenCalled();
    expect(syncReadingProgress).not.toHaveBeenCalled();
  });

  it('uses the left half of the reader to advance in right-to-left mode', async () => {
    getReadingSettings.mockResolvedValue({ preferredTranslationLanguage: 'all', preferredReadingMode: 'rtl' });
    getChapterProgress.mockResolvedValue({ data: { currentPage: 2, totalPages: 2 } });
    getChapterPage.mockImplementation((_chapterId: string, page: number) => Promise.resolve({
      pageNumber: page,
      totalPages: 2,
      pageUrl: `/images/page-${page}.jpg`,
    }));
    render(
      <MemoryRouter initialEntries={[{ pathname: '/manga/manga-1/chapter/chapter-1', state: { chapterIds: ['chapter-1'] } }]}>
        <Routes>
          <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderPage />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => expect(getChapterPage).toHaveBeenCalledWith('chapter-1', 1, 'data'));
    expect(getChapterPageCount).toHaveBeenCalledWith('chapter-1', 'data');
    expect(MockEventSource.instances).toHaveLength(0);
    const firstPage = await screen.findByAltText('Chapter page 1');
    expect(screen.getByLabelText('Page navigation')).toBeInTheDocument();
    expect(screen.getAllByLabelText('Page navigation')).toHaveLength(1);

    await act(async () => {
      firstPage.dispatchEvent(new Event('load'));
    });

    fireEvent.click(firstPage, { clientX: 0, clientY: 200 });

    await waitFor(() => {
      expect(getChapterPage).toHaveBeenCalledWith('chapter-1', 2, 'data');
    });
    const secondPage = await screen.findByAltText('Chapter page 2');
    await act(async () => {
      secondPage.dispatchEvent(new Event('load'));
    });

    fireEvent.change(screen.getByLabelText('Select page'), { target: { value: '1' } });
    await waitFor(() => {
      expect(getChapterPage).toHaveBeenLastCalledWith('chapter-1', 1, 'data');
    });
    expect(saveReadingProgress).not.toHaveBeenCalled();
    expect(syncReadingProgress).not.toHaveBeenCalled();
  });
});
