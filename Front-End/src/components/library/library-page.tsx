import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Container, Row, Spinner } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { ReadingProgressAllResponse, UserLibraryManga } from '../../app/clients/backend/types';

type ResumeTarget = {
  chapterId: string;
  currentPage: number;
  lastReadAt: string;
};

const CHAPTERS_PER_REQUEST = 100;
const LIBRARY_PAGE_SIZE = 24;

const getChapterId = (chapter: Record<string, unknown>): string => {
  return typeof chapter.id === 'string' ? chapter.id : '';
};

const getChapterNumber = (chapter: Record<string, unknown>): string => {
  const attributes = chapter.attributes as Record<string, unknown> | undefined;
  const chapterNumber = attributes?.chapter;
  return chapterNumber === undefined || chapterNumber === null ? '' : String(chapterNumber);
};

const getResumeTarget = (mangaId: string, readingProgress: ReadingProgressAllResponse | null): ResumeTarget | null => {
  const chapterProgress = readingProgress?.data?.mangas?.[mangaId]?.chapters;
  if (!chapterProgress) {
    return null;
  }

  return Object.entries(chapterProgress).reduce<ResumeTarget | null>((latest, [chapterId, progress]) => {
    if (!latest || new Date(progress.lastReadAt).getTime() > new Date(latest.lastReadAt).getTime()) {
      return { chapterId, currentPage: progress.currentPage, lastReadAt: progress.lastReadAt };
    }
    return latest;
  }, null);
};

const LibraryPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const [library, setLibrary] = useState<UserLibraryManga[]>([]);
  const [libraryTotal, setLibraryTotal] = useState(0);
  const [hasMoreLibrary, setHasMoreLibrary] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [loadMoreError, setLoadMoreError] = useState('');
  const [coverErrors, setCoverErrors] = useState<Set<string>>(new Set());
  const [readingProgress, setReadingProgress] = useState<ReadingProgressAllResponse | null>(null);
  const [readingAction, setReadingAction] = useState<string | null>(null);
  const [readingError, setReadingError] = useState('');

  useEffect(() => {
    const loadLibrary = async () => {
      setIsLoading(true);
      setError('');
      try {
        const [libraryResponse, progressResponse] = await Promise.all([
          backend.getMyLibrary({ limit: LIBRARY_PAGE_SIZE, offset: 0 }),
          backend.getReadingProgress().catch(() => null),
        ]);
        const firstPage = libraryResponse.data || [];
        const total = libraryResponse.total ?? firstPage.length;
        setLibrary(firstPage);
        setLibraryTotal(total);
        setHasMoreLibrary(firstPage.length < total);
        setReadingProgress(progressResponse);
      } catch (err) {
        setError((err as Error).message || 'Unable to load your library.');
      } finally {
        setIsLoading(false);
      }
    };

    loadLibrary();
  }, [backend]);

  const handleLoadMore = async () => {
    setLoadMoreError('');
    setIsLoadingMore(true);
    const offset = library.length;

    try {
      const response = await backend.getMyLibrary({ limit: LIBRARY_PAGE_SIZE, offset });
      const nextPage = response.data || [];
      const total = response.total ?? offset + nextPage.length;

      setLibrary((currentLibrary) => {
        const savedMangaIds = new Set(currentLibrary.map((manga) => manga.mangaId));
        return [...currentLibrary, ...nextPage.filter((manga) => !savedMangaIds.has(manga.mangaId))];
      });
      setLibraryTotal(total);
      setHasMoreLibrary(offset + nextPage.length < total);
    } catch (err) {
      setLoadMoreError((err as Error).message || 'Unable to load more titles. Please try again.');
    } finally {
      setIsLoadingMore(false);
    }
  };

  const loadChapterList = async (mangaId: string) => {
    const chapters: Record<string, unknown>[] = [];
    let offset = 0;
    let total = Infinity;

    while (offset < total) {
      const response = await backend.getMangaChapters(mangaId, {
        limit: CHAPTERS_PER_REQUEST,
        offset,
        translatedLanguage: ['en'],
      });
      chapters.push(...response.data);
      total = response.total;

      if (response.data.length === 0) {
        break;
      }

      offset += response.data.length;
    }

    const chapterIds = chapters.map(getChapterId).filter(Boolean);

    if (chapterIds.length === 0) {
      throw new Error('No readable chapters are available for this manga yet.');
    }

    return { chapters, chapterIds };
  };

  const handleReadFromBeginning = async (event: React.MouseEvent<HTMLButtonElement>, mangaId: string) => {
    event.stopPropagation();
    setReadingError('');
    setReadingAction(`${mangaId}:beginning`);

    try {
      const { chapters, chapterIds } = await loadChapterList(mangaId);
      const chapterOne = chapters.find((chapter) => getChapterNumber(chapter) === '1');
      const firstChapterId = chapterOne ? getChapterId(chapterOne) : chapterIds[0];

      navigate(`/manga/${mangaId}/chapter/${firstChapterId}`, { state: { chapterIds, startAtPage: 1 } });
    } catch (err) {
      setReadingError((err as Error).message || 'Unable to start this manga. Please try again.');
    } finally {
      setReadingAction(null);
    }
  };

  const handleContinueReading = async (event: React.MouseEvent<HTMLButtonElement>, mangaId: string, resumeTarget: ResumeTarget | null) => {
    event.stopPropagation();
    setReadingError('');
    setReadingAction(`${mangaId}:continue`);

    try {
      const { chapters, chapterIds } = await loadChapterList(mangaId);
      const firstChapter = chapters.find((chapter) => getChapterNumber(chapter) === '1');
      const chapterId = resumeTarget?.chapterId ?? (firstChapter ? getChapterId(firstChapter) : chapterIds[0]);

      navigate(`/manga/${mangaId}/chapter/${chapterId}`, {
        state: chapterIds.includes(chapterId) ? { chapterIds } : undefined,
      });
    } catch (err) {
      setReadingError((err as Error).message || 'Unable to continue this manga. Please try again.');
    } finally {
      setReadingAction(null);
    }
  };

  return (
    <Container className="py-5">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h1 className="mb-1">My Library</h1>
          <p className="text-muted mb-0">Titles you have saved for later.</p>
        </div>
        <Button variant="primary" onClick={() => navigate('/manga/search')}>Browse manga</Button>
      </div>

      {isLoading && (
        <div className="text-center p-5">
          <Spinner animation="border" size="sm" className="me-2" />
          Loading your library...
        </div>
      )}

      {!isLoading && error && <Alert variant="danger">{error}</Alert>}
      {!isLoading && readingError && <Alert variant="danger">{readingError}</Alert>}
      {!isLoading && loadMoreError && <Alert variant="danger">{loadMoreError}</Alert>}

      {!isLoading && !error && library.length === 0 && (
        <Card className="text-center p-5">
          <Card.Body>
            <h4 className="mb-3">Your library is empty</h4>
            <p className="text-muted mb-4">Find a manga you like and use “Add to My Library” to save it here.</p>
            <Button variant="primary" onClick={() => navigate('/manga/search')}>Find manga</Button>
          </Card.Body>
        </Card>
      )}

      {!isLoading && !error && library.length > 0 && (
        <>
          <Row>
            {library.map((manga) => (
              <Col key={manga.mangaId} xs={6} sm={4} md={3} lg={2} className="mb-4">
                <Card className="manga-library-card h-100" role="button" onClick={() => navigate(`/manga/${manga.mangaId}`)}>
                  {manga.coverUrl && !coverErrors.has(manga.mangaId) ? (
                    <Card.Img
                      variant="top"
                      src={manga.coverUrl}
                      alt={manga.title}
                      onError={() => setCoverErrors((previous) => new Set(previous).add(manga.mangaId))}
                      style={{ height: '280px', objectFit: 'cover' }}
                    />
                  ) : (
                    <div className="library-cover-placeholder">No cover</div>
                  )}
                  <Card.Body className="p-2">
                    <Card.Title as="h2" className="h6 mb-1">{manga.title}</Card.Title>
                    <Card.Text className="small text-muted mb-0">Added {new Date(manga.addedAt).toLocaleDateString()}</Card.Text>
                    <div className="d-grid gap-2 mt-3">
                      <Button
                        size="sm"
                        variant="primary"
                        onClick={(event) => handleContinueReading(event, manga.mangaId, getResumeTarget(manga.mangaId, readingProgress))}
                        disabled={readingAction === `${manga.mangaId}:continue`}
                      >
                        {readingAction === `${manga.mangaId}:continue` ? 'Opening...' : 'Continue reading'}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline-primary"
                        onClick={(event) => handleReadFromBeginning(event, manga.mangaId)}
                        disabled={readingAction === `${manga.mangaId}:beginning`}
                      >
                        {readingAction === `${manga.mangaId}:beginning` ? 'Opening...' : 'Read from chapter 1'}
                      </Button>
                    </div>
                  </Card.Body>
                </Card>
              </Col>
            ))}
          </Row>
          {hasMoreLibrary && (
            <div className="text-center mt-2">
              <Button variant="outline-primary" onClick={handleLoadMore} disabled={isLoadingMore}>
                {isLoadingMore ? 'Loading...' : `Load more (${library.length} of ${libraryTotal})`}
              </Button>
            </div>
          )}
        </>
      )}
    </Container>
  );
};

export default LibraryPage;
