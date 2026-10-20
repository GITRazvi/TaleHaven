import { act, render, screen, waitFor } from '@testing-library/react';
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
    getImageUrl.mockImplementation((path: string) => `http://localhost:3001${path}`);
    saveReadingProgress.mockResolvedValue({ result: 'ok' });
    syncReadingProgress.mockResolvedValue({ result: 'ok' });
    mockedBackend.mockImplementation(() => ({
      getBaseUrl: () => 'http://localhost:3001',
      getManga,
      getChapter,
      getMangaChapters,
      getChapterProgress,
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
});
