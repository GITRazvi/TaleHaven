import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, Card, Col, Container, Row, Spinner } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import { readBrowserCache, writeBrowserCache } from '../../app/browser-cache';
import Backend from '../../app/clients/backend/backend';
import {
  AniListDiscoverySection,
  AniListLibraryItem,
  AniListManga,
  PopularManga,
  ReadingProgressAllResponse,
  UserLibraryManga,
} from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';

const DISCOVERY_PAGE_SIZE = 12;
const READ_NOW_PAGE_SIZE = 12;
const CONTINUE_READING_LIMIT = 6;
const DISCOVERY_CACHE_MAX_AGE_MS = 15 * 60 * 1000;

type DiscoveryDefinition = {
  key: AniListDiscoverySection;
  title: string;
  emptyMessage: string;
};

const discoverySections: DiscoveryDefinition[] = [
  { key: 'trending', title: 'Trending Now', emptyMessage: 'No trending manga is available right now.' },
  { key: 'popular', title: 'All-Time Popular', emptyMessage: 'No popular manga is available right now.' },
  { key: 'top-rated', title: 'Top Rated', emptyMessage: 'No top-rated manga is available right now.' },
  { key: 'recent', title: 'New Releases', emptyMessage: 'No new releases are available right now.' },
];

type SectionState = {
  manga: AniListManga[];
  page: number;
  hasNextPage: boolean;
  isLoading: boolean;
  isLoadingMore: boolean;
  isRefreshing: boolean;
  isUsingFallback: boolean;
  error: string;
};

type CachedDiscoverySection = Pick<
  SectionState,
  'manga' | 'page' | 'hasNextPage'
>;

type ReadNowState = {
  manga: PopularManga[];
  nextOffset: number;
  hasMore: boolean;
  isLoading: boolean;
  isLoadingMore: boolean;
  error: string;
};

type ContinueReadingItem = {
  manga: UserLibraryManga;
  chapterId: string;
  currentPage: number;
  totalPages?: number;
  isResumable: boolean;
  lastReadAt: string;
  completedChapterIds: Set<string>;
  nextChapter?: {
    id: string;
    number: string;
    language: string;
    scanlationGroupId: string;
  };
};

type ContinueReadingState = {
  started: ContinueReadingItem[];
  planned: AniListLibraryItem[];
  isLoading: boolean;
  error: string;
};

const discoveryCacheKey = (section: AniListDiscoverySection) =>
  `talehaven:anilist:discovery:${section}`;

const initialSectionState = (
  section: AniListDiscoverySection,
): SectionState => {
  const cached = readBrowserCache<CachedDiscoverySection>(
    discoveryCacheKey(section),
    DISCOVERY_CACHE_MAX_AGE_MS,
  );

  return {
    manga: cached?.manga || [],
    page: cached?.page || 0,
    hasNextPage: cached?.hasNextPage || false,
    isLoading: !cached,
    isLoadingMore: false,
    isRefreshing: false,
    isUsingFallback: Boolean(cached),
    error: '',
  };
};

const initialSections = (): Record<AniListDiscoverySection, SectionState> => ({
  trending: initialSectionState('trending'),
  popular: initialSectionState('popular'),
  'top-rated': initialSectionState('top-rated'),
  recent: initialSectionState('recent'),
});

const initialReadNowState = (): ReadNowState => ({
  manga: [],
  nextOffset: 0,
  hasMore: false,
  isLoading: true,
  isLoadingMore: false,
  error: '',
});

const initialContinueReadingState = (): ContinueReadingState => ({
  started: [],
  planned: [],
  isLoading: false,
  error: '',
});

const displayTitle = (manga: AniListManga) => (
  manga.title.english || manga.title.userPreferred || manga.title.romaji || manga.title.native || 'Untitled manga'
);

const progressTimestamp = (lastReadAt: string | undefined) => {
  const timestamp = Date.parse(lastReadAt || '');
  return Number.isNaN(timestamp) ? 0 : timestamp;
};

const isCompletedProgress = (progress: { currentPage: number; totalPages?: number }) => (
  Boolean(progress.totalPages) && progress.currentPage >= (progress.totalPages || 0)
);

const chapterField = (chapter: unknown, field: string): unknown => {
  const data = chapter as Record<string, unknown> | undefined;
  const attributes = data?.attributes as Record<string, unknown> | undefined;
  return data?.[field] ?? attributes?.[field];
};

const numericChapter = (chapter: unknown): number | null => {
  if (typeof chapter !== 'string' && typeof chapter !== 'number') return null;
  const value = Number(chapter);
  return Number.isFinite(value) ? value : null;
};

const buildStartedTitles = (
  library: UserLibraryManga[],
  progress: ReadingProgressAllResponse,
): ContinueReadingItem[] => {
  const mangaProgress = progress.data?.mangas || {};

  return library.flatMap((manga) => {
    const chapters = Object.entries(mangaProgress[manga.mangaId]?.chapters || {})
      .map(([chapterId, chapter]) => ({ chapterId, ...chapter }));
    const completedChapters = chapters.filter(isCompletedProgress);

    // A title belongs here only after the reader has completed at least one
    // chapter and explicitly saved it to their MangaDex library.
    if (completedChapters.length === 0) {
      return [];
    }

    const unfinishedChapters = chapters
      .filter((chapter) => !isCompletedProgress(chapter))
      .sort((left, right) => progressTimestamp(right.lastReadAt) - progressTimestamp(left.lastReadAt));
    const newestCompletedChapter = completedChapters
      .sort((left, right) => progressTimestamp(right.lastReadAt) - progressTimestamp(left.lastReadAt))[0];
    const resumeChapter = unfinishedChapters[0] || newestCompletedChapter;

    return [{
      manga,
      chapterId: resumeChapter.chapterId,
      currentPage: resumeChapter.currentPage,
      totalPages: resumeChapter.totalPages,
      isResumable: !isCompletedProgress(resumeChapter),
      lastReadAt: resumeChapter.lastReadAt,
      completedChapterIds: new Set(completedChapters.map((chapter) => chapter.chapterId)),
    }];
  }).sort((left, right) => progressTimestamp(right.lastReadAt) - progressTimestamp(left.lastReadAt));
};

const findSameGroupNextChapter = async (
  backend: Backend,
  item: ContinueReadingItem,
): Promise<ContinueReadingItem> => {
  try {
    const chapterResponse = await backend.getChapter(item.chapterId);
    const lastChapter = chapterResponse.data;
    const language = chapterField(lastChapter, 'translatedLanguage');
    const scanlationGroupId = chapterField(lastChapter, 'scanlationGroupId');

    if (
      typeof language !== 'string' ||
      !language ||
      typeof scanlationGroupId !== 'string' ||
      !scanlationGroupId
    ) {
      return item;
    }

    const candidates: Array<{ id: string; number: string }> = [];
    let offset = 0;

    // The API returns filtered chapter pages of at most 100 entries. Follow
    // pagination so a prolific scanlation group cannot hide a later chapter.
    do {
      const response = await backend.getMangaChapters(item.manga.mangaId, {
        limit: 100,
        offset,
        language,
        scanlationGroup: scanlationGroupId,
      });
      candidates.push(...response.data.flatMap((chapter) => {
        const id = chapterField(chapter, 'id');
        const number = chapterField(chapter, 'chapter');
        if (typeof id !== 'string' || numericChapter(number) === null) return [];
        return [{ id, number: String(number) }];
      }));

      offset += response.data.length;
      if (response.data.length === 0 || offset >= response.total) break;
    } while (true);

    // A skipped chapter is just as actionable as a newly released one. Only
    // completed chapter IDs are excluded; in-progress chapters are handled by
    // the primary Resume action above.
    const nextChapter = candidates
      .filter((chapter) => chapter.id !== item.chapterId && !item.completedChapterIds.has(chapter.id))
      .sort((left, right) => (numericChapter(left.number) || 0) - (numericChapter(right.number) || 0))[0];

    if (!nextChapter) return item;

    return {
      ...item,
      nextChapter: {
        ...nextChapter,
        language,
        scanlationGroupId,
      },
    };
  } catch {
    // Availability is helpful metadata. A transient chapter-source failure
    // must not prevent a reader from resuming their saved progress.
    return item;
  }
};

const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();
  const [sections, setSections] = useState<Record<AniListDiscoverySection, SectionState>>(initialSections);
  const [readNow, setReadNow] = useState<ReadNowState>(initialReadNowState);
  const [continueReading, setContinueReading] = useState<ContinueReadingState>(initialContinueReadingState);
  const [coverErrors, setCoverErrors] = useState<Set<number>>(new Set());
  const [readNowCoverErrors, setReadNowCoverErrors] = useState<Set<string>>(new Set());
  const hasLoadedInitialSections = useRef(false);
  const [savedAniListIDs, setSavedAniListIDs] = useState<Set<number>>(new Set());
  const [libraryActionID, setLibraryActionID] = useState<number | null>(null);
  const [favoriteAniListIDs, setFavoriteAniListIDs] = useState<Set<number>>(new Set());
  const [favoriteAniListActionID, setFavoriteAniListActionID] = useState<number | null>(null);
  const [savedMangaDexIDs, setSavedMangaDexIDs] = useState<Set<string>>(new Set());
  const [mangaDexLibraryActionID, setMangaDexLibraryActionID] = useState<string | null>(null);
  const [favoriteMangaDexIDs, setFavoriteMangaDexIDs] = useState<Set<string>>(new Set());
  const [favoriteMangaDexActionID, setFavoriteMangaDexActionID] = useState<string | null>(null);

  const loadSection = async (section: AniListDiscoverySection, page: number, append: boolean) => {
    setSections((current) => ({
      ...current,
      [section]: {
        ...current[section],
        isLoading: !append && current[section].manga.length === 0,
        isLoadingMore: append,
        isRefreshing: !append && current[section].manga.length > 0,
        isUsingFallback: false,
        error: '',
      },
    }));

    try {
      const response = await backend.getAniListDiscoverySection(section, page, DISCOVERY_PAGE_SIZE);
      const receivedManga = response.data || [];
      if (!append) {
        writeBrowserCache<CachedDiscoverySection>(discoveryCacheKey(section), {
          manga: receivedManga,
          page: response.page || page,
          hasNextPage: response.hasNextPage,
        });
      }
      setSections((current) => {
        const previous = current[section];
        const manga = append
          ? [...previous.manga, ...receivedManga.filter((item) => !previous.manga.some((known) => known.id === item.id))]
          : receivedManga;

        return {
          ...current,
          [section]: {
            ...previous,
            manga,
            page: response.page || page,
            hasNextPage: response.hasNextPage,
            isLoading: false,
            isLoadingMore: false,
            isRefreshing: false,
            isUsingFallback: false,
          },
        };
      });
    } catch (error) {
      const cached = readBrowserCache<CachedDiscoverySection>(
        discoveryCacheKey(section),
        DISCOVERY_CACHE_MAX_AGE_MS,
      );
      setSections((current) => {
        const currentSection = current[section];
        const fallback = cached ||
          (currentSection.manga.length > 0
            ? {
                manga: currentSection.manga,
                page: currentSection.page,
                hasNextPage: currentSection.hasNextPage,
              }
            : null);

        return {
          ...current,
          [section]: {
            ...currentSection,
            manga: fallback?.manga || [],
            page: fallback?.page || 0,
            hasNextPage: fallback?.hasNextPage || false,
            isLoading: false,
            isLoadingMore: false,
            isRefreshing: false,
            isUsingFallback: Boolean(fallback),
            error: (error as Error).message || 'Unable to load this manga list. Please try again.',
          },
        };
      });
    }
  };

  const loadReadNow = async (offset: number, append: boolean) => {
    setReadNow((current) => ({
      ...current,
      isLoading: !append,
      isLoadingMore: append,
      error: '',
    }));

    try {
      const response = await backend.getPopularManga({
        limit: READ_NOW_PAGE_SIZE,
        offset,
        orderBy: 'followedCount',
      });
      const receivedManga = response.data || [];
      setReadNow((current) => {
        const manga = append
          ? [...current.manga, ...receivedManga.filter((item) => !current.manga.some((known) => known.id === item.id))]
          : receivedManga;

        return {
          ...current,
          manga,
          nextOffset: response.offset + receivedManga.length,
          hasMore: response.offset + receivedManga.length < response.total,
          isLoading: false,
          isLoadingMore: false,
        };
      });
    } catch (error) {
      setReadNow((current) => ({
        ...current,
        isLoading: false,
        isLoadingMore: false,
        error: (error as Error).message || 'Unable to load readable manga. Please try again.',
      }));
    }
  };

  useEffect(() => {
    if (hasLoadedInitialSections.current) {
      return;
    }
    hasLoadedInitialSections.current = true;

    discoverySections.forEach(({ key }) => {
      void loadSection(key, 1, false);
    });
    void loadReadNow(0, false);
    // backend is stable because it is memoized for this component instance.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [backend]);

  useEffect(() => {
    if (!user) {
      setSavedAniListIDs(new Set());
      setSavedMangaDexIDs(new Set());
      setFavoriteAniListIDs(new Set());
      setFavoriteMangaDexIDs(new Set());
      setContinueReading(initialContinueReadingState());
      return;
    }

    let cancelled = false;
    setContinueReading((current) => ({ ...current, isLoading: true, error: '' }));

    void Promise.all([
      backend.getAniListMyLibrary({ limit: 100, offset: 0 }),
      backend.getMyLibrary({ limit: 100, offset: 0 }),
      backend.getReadingProgress(),
    ]).then(async ([aniListLibrary, mangaDexLibrary, progress]) => {
      if (cancelled) return;
      setSavedAniListIDs(new Set(aniListLibrary.data.map((item) => item.anilistId)));
      setSavedMangaDexIDs(new Set(mangaDexLibrary.data.map((item) => item.mangaId)));
      const started = buildStartedTitles(mangaDexLibrary.data, progress).slice(0, CONTINUE_READING_LIMIT);
      const startedWithAvailability = await Promise.all(
        started.map((item) => findSameGroupNextChapter(backend, item)),
      );
      if (cancelled) return;
      setContinueReading({
        // A completed title belongs here only when its last scanlation group
        // still contains an unread chapter. Otherwise there is nothing to
        // resume.
        started: startedWithAvailability.filter((item) => item.isResumable || Boolean(item.nextChapter)),
        planned: aniListLibrary.data.filter((item) => item.readingStatus === 'PLAN_TO_READ'),
        isLoading: false,
        error: '',
      });
    }).catch((error) => {
      if (cancelled) return;
      setSavedAniListIDs(new Set());
      setSavedMangaDexIDs(new Set());
      setContinueReading({
        ...initialContinueReadingState(),
        error: (error as Error).message || 'Unable to load your reading list.',
      });
    });
    void backend.getAniListFavoriteIDs()
      .then((response) => setFavoriteAniListIDs(new Set(response.data)))
      .catch(() => setFavoriteAniListIDs(new Set()));
    void backend.getMangaDexFavoriteIDs()
      .then((response) => setFavoriteMangaDexIDs(new Set(response.data)))
      .catch(() => setFavoriteMangaDexIDs(new Set()));

    return () => {
      cancelled = true;
    };
  }, [backend, user]);

  const handleCoverError = (mangaId: number) => {
    setCoverErrors((current) => new Set(current).add(mangaId));
  };

  const handleReadNowCoverError = (mangaId: string) => {
    setReadNowCoverErrors((current) => new Set(current).add(mangaId));
  };

  const toggleLibrary = async (event: React.MouseEvent<HTMLButtonElement>, manga: AniListManga) => {
    event.stopPropagation();
    if (!user) { navigate('/login'); return; }
    const saved = savedAniListIDs.has(manga.id); setLibraryActionID(manga.id);
    try { if (saved) { await backend.removeAniListMangaFromMyLibrary(manga.id); } else { await backend.addAniListMangaToMyLibrary(manga.id); }
      setSavedAniListIDs((current) => { const next = new Set(current); if (saved) next.delete(manga.id); else next.add(manga.id); return next; });
    } finally { setLibraryActionID(null); }
  };

  const toggleMangaDexLibrary = async (event: React.MouseEvent<HTMLButtonElement>, manga: PopularManga) => {
    event.stopPropagation();
    if (!user) {
      navigate('/login');
      return;
    }

    const saved = savedMangaDexIDs.has(manga.id);
    setMangaDexLibraryActionID(manga.id);
    try {
      if (saved) {
        await backend.removeMangaFromMyLibrary(manga.id);
      } else {
        await backend.addMangaToMyLibrary(manga.id);
      }
      setSavedMangaDexIDs((current) => {
        const next = new Set(current);
        if (saved) next.delete(manga.id); else next.add(manga.id);
        return next;
      });
    } finally {
      setMangaDexLibraryActionID(null);
    }
  };

  const toggleAniListFavorite = async (event: React.MouseEvent<HTMLButtonElement>, manga: AniListManga) => {
    event.stopPropagation();
    if (!user) { navigate('/login'); return; }
    const favorite = favoriteAniListIDs.has(manga.id);
    setFavoriteAniListActionID(manga.id);
    try {
      if (favorite) await backend.removeAniListMangaFromFavorites(manga.id);
      else await backend.addAniListMangaToFavorites(manga.id);
      setFavoriteAniListIDs((current) => { const next = new Set(current); if (favorite) next.delete(manga.id); else next.add(manga.id); return next; });
    } finally { setFavoriteAniListActionID(null); }
  };

  const toggleMangaDexFavorite = async (event: React.MouseEvent<HTMLButtonElement>, manga: PopularManga) => {
    event.stopPropagation();
    if (!user) { navigate('/login'); return; }
    const favorite = favoriteMangaDexIDs.has(manga.id);
    setFavoriteMangaDexActionID(manga.id);
    try {
      if (favorite) await backend.removeMangaFromFavorites(manga.id);
      else await backend.addMangaToFavorites(manga.id);
      setFavoriteMangaDexIDs((current) => { const next = new Set(current); if (favorite) next.delete(manga.id); else next.add(manga.id); return next; });
    } finally { setFavoriteMangaDexActionID(null); }
  };

  return (
    <Container className="p-5">
      <div className="text-center mb-5">
        <h1 className="display-4 mb-3">
          Welcome{user ? ` back, ${user.email.split('@')[0]}` : ' to TaleHaven'}!
        </h1>
        <p className="lead text-muted">Discover manga metadata from AniList, then read available chapters through MangaDex.</p>
        {user && (
          <Alert variant="info" className="mt-3 d-inline-block">
            🎉 Your reading progress is saved to your account.
          </Alert>
        )}
      </div>

      {user && (
        <section className="mb-5" aria-labelledby="continue-reading-heading">
          <div className="d-flex justify-content-between align-items-center mb-3">
            <div>
              <h2 id="continue-reading-heading" className="mb-0">Continue Reading</h2>
              <p className="text-muted mb-0 small">Your saved manga with completed chapters, plus your Plan to Read list.</p>
            </div>
            <Button variant="outline-primary" size="sm" onClick={() => navigate('/library')}>View My Profile</Button>
          </div>

          {continueReading.isLoading && (
            <div className="text-center p-4" aria-label="Loading continue reading">
              <Spinner animation="border" size="sm" className="me-2" />
              Loading your reading list...
            </div>
          )}

          {!continueReading.isLoading && continueReading.error && (
            <Alert variant="warning">Unable to load Continue Reading: {continueReading.error}</Alert>
          )}

          {!continueReading.isLoading && !continueReading.error && continueReading.started.length === 0 && continueReading.planned.length === 0 && (
            <Card className="text-center p-3"><Card.Body>Save a manga and finish a chapter to see it here, or add one to Plan to Read.</Card.Body></Card>
          )}

          {!continueReading.isLoading && continueReading.started.length > 0 && (
            <Row className="g-3">
              {continueReading.started.map((item) => (
                <Col key={item.manga.mangaId} xs={6} sm={4} md={3} lg={2}>
                  <Card className="profile-reading-card h-100">
                    {item.manga.coverUrl ? (
                      <Card.Img variant="top" src={item.manga.coverUrl} alt={item.manga.title} referrerPolicy="no-referrer" />
                    ) : <div className="library-cover-placeholder">No cover</div>}
                    <Card.Body className="p-2 d-flex flex-column">
                      <strong className="profile-reading-title" title={item.manga.title}>{item.manga.title}</strong>
                      <small className="text-muted mt-1">
                        {item.isResumable
                          ? `Chapter in progress · Page ${item.currentPage}${item.totalPages ? ` of ${item.totalPages}` : ''}`
                          : 'Latest saved chapter completed'}
                      </small>
                      {item.nextChapter && (
                        <small className="text-success mt-1">
                          Unread Chapter {item.nextChapter.number} in your last scanlation group
                        </small>
                      )}
                      <Button
                        size="sm"
                        variant="outline-primary"
                        className="mt-auto pt-1"
                        onClick={() => {
                          if (item.isResumable) {
                            navigate(`/manga/${item.manga.mangaId}/chapter/${item.chapterId}`);
                          } else if (item.nextChapter) {
                            navigate(`/manga/${item.manga.mangaId}/chapter/${item.nextChapter.id}`, {
                              state: {
                                chapterFilters: {
                                  language: item.nextChapter.language,
                                  scanlationGroup: item.nextChapter.scanlationGroupId,
                                },
                              },
                            });
                          } else {
                            navigate(`/manga/${item.manga.mangaId}`);
                          }
                        }}
                      >
                        {item.isResumable ? 'Resume' : item.nextChapter ? `Read Chapter ${item.nextChapter.number}` : 'Choose next chapter'}
                      </Button>
                    </Card.Body>
                  </Card>
                </Col>
              ))}
            </Row>
          )}

          {!continueReading.isLoading && continueReading.planned.length > 0 && (
            <>
              <h3 className="h5 mt-4 mb-3">Plan to Read</h3>
              <Row className="g-3">
                {continueReading.planned.map((item) => {
                  const title = displayTitle(item.manga);
                  const cover = item.manga.coverImage.large || item.manga.coverImage.medium;
                  return (
                    <Col key={item.anilistId} xs={6} sm={4} md={3} lg={2}>
                      <Card className="profile-reading-card h-100" role="button" onClick={() => navigate(`/anilist/manga/${item.anilistId}`)}>
                        {cover ? <Card.Img variant="top" src={cover} alt={title} referrerPolicy="no-referrer" /> : <div className="library-cover-placeholder">No cover</div>}
                        <Card.Body className="p-2"><strong className="profile-reading-title" title={title}>{title}</strong><small className="text-muted d-block mt-1">Plan to Read</small></Card.Body>
                      </Card>
                    </Col>
                  );
                })}
              </Row>
            </>
          )}
        </section>
      )}

      <section className="mb-5" aria-labelledby="read-now-heading">
        <div className="d-flex justify-content-between align-items-center mb-3">
          <div>
            <h2 id="read-now-heading" className="mb-0">Read on TaleHaven</h2>
            <p className="text-muted mb-0 small">Popular MangaDex titles with English chapters available.</p>
          </div>
          {readNow.hasMore && readNow.manga.length > 0 && (
            <Button
              variant="outline-primary"
              size="sm"
              onClick={() => void loadReadNow(readNow.nextOffset, true)}
              disabled={readNow.isLoadingMore}
            >
              {readNow.isLoadingMore ? 'Loading...' : 'Load more'}
            </Button>
          )}
        </div>

        {readNow.isLoading && (
          <div className="text-center p-4" aria-label="Loading readable manga">
            <Spinner animation="border" size="sm" className="me-2" />
            Loading readable manga...
          </div>
        )}

        {!readNow.isLoading && readNow.error && (
          <Alert variant="danger">Unable to load readable manga: {readNow.error}</Alert>
        )}

        {!readNow.isLoading && !readNow.error && readNow.manga.length === 0 && (
          <Card className="text-center p-3"><Card.Body>No readable manga is available right now.</Card.Body></Card>
        )}

        {!readNow.isLoading && readNow.manga.length > 0 && (
          <Row>
            {readNow.manga.map((manga) => (
              <Col key={manga.id} xs={6} sm={4} md={3} lg={2} className="mb-4">
                <Card className="manga-library-card h-100 shadow-sm" role="button" onClick={() => navigate(`/manga/${manga.id}`)}>
                  <div className="position-relative">
                    {!readNowCoverErrors.has(manga.id) && manga.coverUrl ? (
                      <Card.Img
                        variant="top"
                        src={manga.coverUrl}
                        alt={manga.title}
                        referrerPolicy="no-referrer"
                        onError={() => handleReadNowCoverError(manga.id)}
                        style={{ height: '260px', objectFit: 'cover', width: '100%' }}
                      />
                    ) : (
                      <div className="d-flex align-items-center justify-content-center text-muted" style={{ height: '260px', backgroundColor: '#f0f0f0' }}>
                        No Cover
                      </div>
                    )}
                    <button
                      type="button"
                      className="home-library-toggle"
                      onClick={(event) => void toggleMangaDexLibrary(event, manga)}
                      disabled={mangaDexLibraryActionID === manga.id}
                      aria-label={savedMangaDexIDs.has(manga.id) ? `Remove ${manga.title} from My Library` : `Add ${manga.title} to My Library`}
                    >
                      {mangaDexLibraryActionID === manga.id ? '…' : savedMangaDexIDs.has(manga.id) ? '−' : '+'}
                    </button>
                    {savedMangaDexIDs.has(manga.id) && <button
                      type="button"
                      className={`home-library-heart ${favoriteMangaDexIDs.has(manga.id) ? 'is-saved' : ''}`}
                      onClick={(event) => void toggleMangaDexFavorite(event, manga)}
                      disabled={favoriteMangaDexActionID === manga.id}
                      aria-label={favoriteMangaDexIDs.has(manga.id) ? `Remove ${manga.title} from Favorites` : `Add ${manga.title} to Favorites`}
                    >
                      {favoriteMangaDexActionID === manga.id ? '…' : favoriteMangaDexIDs.has(manga.id) ? '♥' : '♡'}
                    </button>}
                  </div>
                  <Card.Body className="p-2 d-flex flex-column">
                    <Card.Title className="small text-center mb-1" title={manga.title} style={{ overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>
                      {manga.title}
                    </Card.Title>
                    <Card.Text className="small text-muted text-center mb-0">Open chapters</Card.Text>
                  </Card.Body>
                </Card>
              </Col>
            ))}
          </Row>
        )}
      </section>

      {discoverySections.map(({ key, title, emptyMessage }) => {
        const section = sections[key];
        return (
          <section key={key} className="mb-5" aria-labelledby={`${key}-heading`}>
            <div className="d-flex justify-content-between align-items-center mb-3">
              <h2 id={`${key}-heading`} className="mb-0">{title}</h2>
              {section.hasNextPage && section.manga.length > 0 && (
                <Button
                  variant="outline-primary"
                  size="sm"
                  onClick={() => void loadSection(key, section.page + 1, true)}
                  disabled={section.isLoadingMore}
                >
                  {section.isLoadingMore ? 'Loading...' : 'Load more'}
                </Button>
              )}
            </div>

            {section.isLoading && (
              <div className="text-center p-4" aria-label={`Loading ${title}`}>
                <Spinner animation="border" size="sm" className="me-2" />
                Loading {title.toLowerCase()}...
              </div>
            )}

            {!section.isLoading && section.error && (
              <Alert variant={section.isUsingFallback ? "warning" : "danger"}>
                {section.isUsingFallback
                  ? `AniList is temporarily unavailable. Showing saved ${title.toLowerCase()} from this browser.`
                  : `Unable to load ${title.toLowerCase()}: ${section.error}`}
                <Button
                  className="ms-2"
                  variant={section.isUsingFallback ? "outline-dark" : "outline-light"}
                  size="sm"
                  onClick={() => void loadSection(key, 1, false)}
                >
                  Retry
                </Button>
              </Alert>
            )}

            {!section.isLoading && !section.error && section.isRefreshing && (
              <p className="small text-muted mb-3" role="status">
                Refreshing saved results...
              </p>
            )}

            {!section.isLoading && !section.error && section.manga.length === 0 && (
              <Card className="text-center p-3">
                <Card.Body>{emptyMessage}</Card.Body>
              </Card>
            )}

            {!section.isLoading && section.manga.length > 0 && (
              <Row>
                {section.manga.map((manga) => {
                  const titleText = displayTitle(manga);
                  const coverURL = manga.coverImage.large || manga.coverImage.medium;
                  return (
                    <Col key={manga.id} xs={6} sm={4} md={3} lg={2} className="mb-4">
                      <Card className="manga-library-card h-100 shadow-sm" role="button" onClick={() => navigate(`/anilist/manga/${manga.id}`)}>
                        <div className="position-relative">
                        {!coverErrors.has(manga.id) && coverURL ? (
                          <Card.Img
                            variant="top"
                            src={coverURL}
                            alt={titleText}
                            referrerPolicy="no-referrer"
                            onError={() => handleCoverError(manga.id)}
                            style={{ height: '260px', objectFit: 'cover', width: '100%' }}
                          />
                        ) : (
                          <div
                            className="d-flex align-items-center justify-content-center text-muted"
                            style={{ height: '260px', backgroundColor: manga.coverImage.color || '#f0f0f0' }}
                          >
                            No Cover
                          </div>
                        )}
                        <button type="button" className="home-library-toggle"
                          onClick={(event) => void toggleLibrary(event, manga)} disabled={libraryActionID === manga.id}
                          aria-label={savedAniListIDs.has(manga.id) ? `Remove ${titleText} from My Library` : `Add ${titleText} to My Library`}>
                          {libraryActionID === manga.id ? '…' : savedAniListIDs.has(manga.id) ? '−' : '+'}
                        </button>
                        {savedAniListIDs.has(manga.id) && <button type="button" className={`home-library-heart ${favoriteAniListIDs.has(manga.id) ? 'is-saved' : ''}`}
                          onClick={(event) => void toggleAniListFavorite(event, manga)} disabled={favoriteAniListActionID === manga.id}
                          aria-label={favoriteAniListIDs.has(manga.id) ? `Remove ${titleText} from Favorites` : `Add ${titleText} to Favorites`}>
                          {favoriteAniListActionID === manga.id ? '…' : favoriteAniListIDs.has(manga.id) ? '♥' : '♡'}
                        </button>}
                        </div>
                        <Card.Body className="p-2 d-flex flex-column">
                          <Card.Title
                            className="small text-center mb-1"
                            title={titleText}
                            style={{
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              display: '-webkit-box',
                              WebkitLineClamp: 2,
                              WebkitBoxOrient: 'vertical',
                            }}
                          >
                            {titleText}
                          </Card.Title>
                          {manga.averageScore > 0 && (
                            <Card.Text className="small text-muted text-center mb-0">★ {manga.averageScore}%</Card.Text>
                          )}
                        </Card.Body>
                      </Card>
                    </Col>
                  );
                })}
              </Row>
            )}
          </section>
        );
      })}
    </Container>
  );
};

export default HomePage;
