import React, { useEffect, useMemo, useState } from 'react';
import { Button, ListGroup, Spinner } from 'react-bootstrap';
import { Link, useNavigate, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';

type MangaChapterItem = {
  id: string;
  number: string;
  title: string;
};

type MangaView = {
  id: string;
  title: string;
  description: string;
};

const MANUAL_COMPLETION_PAGE = 2147483647;

const readLocalizedValue = (value: unknown): string => {
  if (typeof value === 'string') {
    return value;
  }

  if (value && typeof value === 'object') {
    const localized = value as Record<string, unknown>;
    const preferred = localized.en ?? Object.values(localized)[0];
    return typeof preferred === 'string' ? preferred : '';
  }

  return '';
};

const mapManga = (input: unknown): MangaView => {
  console.log('[mapManga] Input:', input);
  
  const data = (input ?? {}) as Record<string, unknown>;

  // Backend already transforms the data, so title/description are at root level
  const title = typeof data.title === 'string' ? data.title :
                readLocalizedValue(data.title) || 'Untitled manga';

  const description = typeof data.description === 'string' ? data.description :
                      readLocalizedValue(data.description) || 'No description available.';

  const result = {
    id: typeof data.id === 'string' ? data.id : '',
    title,
    description,
  };
  
  console.log('[mapManga] Result:', result);
  return result;
};

const mapChapter = (input: unknown): MangaChapterItem | null => {
  if (!input || typeof input !== 'object') {
    return null;
  }

  const data = input as Record<string, unknown>;
  const id = typeof data.id === 'string' ? data.id : '';

  if (!id) {
    return null;
  }

  // Backend may return flat data or nested attributes
  const chapter = data.chapter ?? (data.attributes as Record<string, unknown> | undefined)?.chapter;
  const title = data.title ?? (data.attributes as Record<string, unknown> | undefined)?.title;

  return {
    id,
    number: String(chapter ?? '?'),
    title: (typeof title === 'string' && title) || 'Untitled chapter',
  };
};

const MangaDetailsPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const { mangaId = '' } = useParams<{ mangaId: string }>();
  const { user } = useAuth();

  const [manga, setManga] = useState<MangaView | null>(null);
  const [coverUrl, setCoverUrl] = useState<string>('');
  const [coverError, setCoverError] = useState<boolean>(false);
  const [chapters, setChapters] = useState<MangaChapterItem[]>([]);
  const [completedChapterIds, setCompletedChapterIds] = useState<Set<string>>(new Set());
  const [chapterProgressPages, setChapterProgressPages] = useState<Record<string, number>>({});
  const [resumeChapterId, setResumeChapterId] = useState<string | null>(null);
  const [selectedChapterIds, setSelectedChapterIds] = useState<Set<string>>(new Set());
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [bulkProgressAction, setBulkProgressAction] = useState<'read' | 'unread' | null>(null);
  const [bulkProgressError, setBulkProgressError] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [libraryError, setLibraryError] = useState<string>('');
  const [isInLibrary, setIsInLibrary] = useState<boolean>(false);
  const [isLibraryLoading, setIsLibraryLoading] = useState<boolean>(false);

  const handleLibraryToggle = async () => {
    if (!user) {
      navigate('/login');
      return;
    }

    setLibraryError('');
    setIsLibraryLoading(true);
    try {
      if (isInLibrary) {
        await backend.removeMangaFromMyLibrary(mangaId);
        setIsInLibrary(false);
      } else {
        await backend.addMangaToMyLibrary(mangaId);
        setIsInLibrary(true);
      }
    } catch (err) {
      setLibraryError((err as Error).message || 'Unable to update your library.');
    } finally {
      setIsLibraryLoading(false);
    }
  };

  useEffect(() => {
    let isCurrent = true;

    if (!user || !mangaId) {
      setIsInLibrary(false);
      return () => {
        isCurrent = false;
      };
    }

    const loadLibraryMembership = async () => {
      setIsLibraryLoading(true);
      try {
        const response = await backend.getMyLibrary();
        if (isCurrent) {
          setIsInLibrary(response.data.some((libraryManga) => libraryManga.mangaId === mangaId));
        }
      } catch {
        // Keep the heart hollow when membership cannot be determined. The
        // toggle itself will still display a useful API error if it fails.
        if (isCurrent) {
          setIsInLibrary(false);
        }
      } finally {
        if (isCurrent) {
          setIsLibraryLoading(false);
        }
      }
    };

    loadLibraryMembership();
    return () => {
      isCurrent = false;
    };
  }, [backend, mangaId, user]);

  useEffect(() => {
    const loadData = async () => {
      if (!mangaId) {
        setError('Missing manga id.');
        setIsLoading(false);
        return;
      }

      setIsLoading(true);
      setError('');
      setCoverError(false);
      setSelectedChapterIds(new Set());
      setIsSelectionMode(false);
      setBulkProgressError('');

      try {
        console.log('[MangaDetails] Loading data for mangaId:', mangaId);

        const [mangaResponse, firstChaptersResponse, coverResponse, progressResponse] = await Promise.all([
          backend.getManga(mangaId),
          backend.getMangaChapters(mangaId, { limit: 100, offset: 0, translatedLanguage: ['en'] }),
          backend.getMangaCover(mangaId, '512').catch((err) => {
            console.warn('[MangaDetails] Cover fetch failed:', err);
            return { result: 'ok', coverUrl: '' };
          }),
          backend.getMangaProgress(mangaId).catch(() => null),
        ]);

        console.log('[MangaDetails] Manga response:', mangaResponse);
        const chapterResponses = [firstChaptersResponse];
        let offset = firstChaptersResponse.data.length;

        while (offset < firstChaptersResponse.total && firstChaptersResponse.data.length > 0) {
          const nextResponse = await backend.getMangaChapters(mangaId, {
            limit: 100,
            offset,
            translatedLanguage: ['en'],
          });
          chapterResponses.push(nextResponse);

          if (nextResponse.data.length === 0) {
            break;
          }

          offset += nextResponse.data.length;
        }

        const chapterData = chapterResponses.flatMap((response) => response.data);
        console.log('[MangaDetails] Chapters response:', chapterData);
        console.log('[MangaDetails] Cover response:', coverResponse);

        const mappedManga = mapManga(mangaResponse.data);
        const mappedChapters = chapterData.map(mapChapter).filter((item): item is MangaChapterItem => item !== null);
        const completedIds = new Set(
          Object.entries(progressResponse?.data?.chapters ?? {})
            .filter(([, progress]) => Boolean(progress.totalPages) && progress.currentPage >= (progress.totalPages ?? 0))
            .map(([chapterId]) => chapterId)
        );
        const savedPages = Object.entries(progressResponse?.data?.chapters ?? {}).reduce<Record<string, number>>(
          (progressByChapter, [chapterId, progress]) => {
            progressByChapter[chapterId] = progress.currentPage;
            return progressByChapter;
          },
          {}
        );
        const latestProgress = Object.entries(progressResponse?.data?.chapters ?? {}).reduce<{
          chapterId: string;
          lastReadAt: string;
        } | null>((latest, [chapterId, progress]) => {
          if (!latest || new Date(progress.lastReadAt).getTime() > new Date(latest.lastReadAt).getTime()) {
            return { chapterId, lastReadAt: progress.lastReadAt };
          }
          return latest;
        }, null);

        console.log('[MangaDetails] Mapped manga:', mappedManga);
        console.log('[MangaDetails] Mapped chapters:', mappedChapters);

        setManga(mappedManga);
        setChapters(mappedChapters);
        setCompletedChapterIds(completedIds);
        setChapterProgressPages(savedPages);
        setResumeChapterId(latestProgress?.chapterId ?? null);
        setCoverUrl(coverResponse.coverUrl || '');
      } catch (err) {
        console.error('[MangaDetails] Error loading data:', err);
        setError((err as Error).message || 'Failed to load manga details.');
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [backend, mangaId]);

  if (isLoading) {
    return (
      <div className="p-3">
        <Spinner animation="border" size="sm" className="me-2" />
        Loading manga details...
      </div>
    );
  }

  if (error) {
    return <div className="p-3 text-danger">Error: {error}</div>;
  }

  if (!manga) {
    return <div className="p-3">No manga data found.</div>;
  }

  const chapterIds = chapters.map((chapter) => chapter.id);
  const handleReadFromChapterOne = () => {
    const firstChapter = chapters.find((chapter) => chapter.number === '1') ?? chapters[0];
    if (!firstChapter) {
      return;
    }

    navigate(`/manga/${mangaId}/chapter/${firstChapter.id}`, {
      state: { chapterIds, startAtPage: 1 },
    });
  };
  const handleContinueReading = () => {
    const firstChapter = chapters.find((chapter) => chapter.number === '1') ?? chapters[0];
    const chapterId = resumeChapterId ?? firstChapter?.id;
    if (!chapterId) {
      return;
    }

    navigate(`/manga/${mangaId}/chapter/${chapterId}`, {
      state: chapterIds.includes(chapterId) ? { chapterIds } : undefined,
    });
  };
  const toggleChapterSelection = (chapterId: string) => {
    setSelectedChapterIds((previous) => {
      const next = new Set(previous);
      if (next.has(chapterId)) {
        next.delete(chapterId);
      } else {
        next.add(chapterId);
      }
      return next;
    });
  };
  const updateSelectedChapterProgress = async (action: 'read' | 'unread') => {
    if (!user) {
      navigate('/login');
      return;
    }

    const chapterIdsToUpdate = Array.from(selectedChapterIds);
    if (chapterIdsToUpdate.length === 0) {
      return;
    }

    setBulkProgressError('');
    setBulkProgressAction(action);
    const updatedChapterIds: string[] = [];
    const failedChapterIds: string[] = [];

    for (const chapterId of chapterIdsToUpdate) {
      try {
        if (action === 'read') {
          await backend.saveReadingProgress({
            mangaId,
            chapterId,
            currentPage: MANUAL_COMPLETION_PAGE,
            totalPages: MANUAL_COMPLETION_PAGE,
          });
          await backend.syncReadingProgress(chapterId);
        } else {
          await backend.deleteChapterProgress(chapterId, mangaId);
        }
        updatedChapterIds.push(chapterId);
      } catch (err) {
        console.warn(`[MangaDetails] Unable to mark chapter ${chapterId} as ${action}:`, err);
        failedChapterIds.push(chapterId);
      }
    }

    if (updatedChapterIds.length > 0) {
      setCompletedChapterIds((previous) => {
        const next = new Set(previous);
        updatedChapterIds.forEach((chapterId) => action === 'read' ? next.add(chapterId) : next.delete(chapterId));
        return next;
      });
      setChapterProgressPages((previous) => {
        const next = { ...previous };
        updatedChapterIds.forEach((chapterId) => {
          if (action === 'read') {
            next[chapterId] = MANUAL_COMPLETION_PAGE;
          } else {
            delete next[chapterId];
          }
        });
        return next;
      });
      setSelectedChapterIds((previous) => {
        const next = new Set(previous);
        updatedChapterIds.forEach((chapterId) => next.delete(chapterId));
        return next;
      });
      if (action === 'unread' && resumeChapterId && updatedChapterIds.includes(resumeChapterId)) {
        setResumeChapterId(null);
      }
    }

    if (failedChapterIds.length > 0) {
      setBulkProgressError(`Unable to update ${failedChapterIds.length} selected chapter${failedChapterIds.length === 1 ? '' : 's'}. Please try again.`);
    }
    setBulkProgressAction(null);
  };

  return (
    <div className="p-3">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="mb-0">{manga?.title}</h2>
        <div className="d-flex gap-2">
          <button
            type="button"
            className={`library-heart ${isInLibrary ? 'is-saved' : ''}`}
            onClick={handleLibraryToggle}
            disabled={isLibraryLoading}
            aria-label={isInLibrary ? 'Remove from My Library' : user ? 'Add to My Library' : 'Log in to save to My Library'}
            aria-pressed={isInLibrary}
            title={isInLibrary ? 'Remove from My Library' : user ? 'Add to My Library' : 'Log in to save'}
          >
            {isLibraryLoading ? '…' : isInLibrary ? '♥' : '♡'}
          </button>
          <Button variant="outline-secondary" size="sm" onClick={() => navigate(-1)}>
            ← Back
          </Button>
        </div>
      </div>

      {libraryError && <div className="text-danger mb-3">{libraryError}</div>}
      {bulkProgressError && <div className="text-danger mb-3">{bulkProgressError}</div>}

      {coverUrl && !coverError && (
        <img
          src={coverUrl}
          alt={manga?.title}
          style={{ maxWidth: '220px' }}
          className="mb-3"
          onError={() => {
            console.warn('[MangaDetails] Cover image failed to load:', coverUrl);
            setCoverError(true);
          }}
        />
      )}

      <p>{manga?.description}</p>

      <h4>Chapters</h4>
      <div className="d-flex gap-2 mb-3">
        <Button variant="primary" onClick={handleContinueReading} disabled={chapters.length === 0}>
          Continue reading
        </Button>
        <Button variant="outline-primary" onClick={handleReadFromChapterOne} disabled={chapters.length === 0}>
          Read from chapter 1
        </Button>
      </div>
      {chapters.length === 0 ? (
        <div>No chapters available.</div>
      ) : (
        <div>
          <div className="d-flex flex-wrap align-items-center gap-2 mb-2">
            <Button
              size="sm"
              variant="outline-secondary"
              onClick={() => {
                setIsSelectionMode((previous) => !previous);
                setSelectedChapterIds(new Set());
              }}
              disabled={bulkProgressAction !== null}
            >
              {isSelectionMode ? 'Done selecting' : 'Select chapters'}
            </Button>
            {isSelectionMode && (
              <>
                <span className="text-muted small">{selectedChapterIds.size} selected</span>
                <Button size="sm" variant="outline-secondary" onClick={() => setSelectedChapterIds(new Set(chapterIds))} disabled={bulkProgressAction !== null}>
                  Select all
                </Button>
                <Button size="sm" variant="outline-secondary" onClick={() => setSelectedChapterIds(new Set())} disabled={selectedChapterIds.size === 0 || bulkProgressAction !== null}>
                  Clear selection
                </Button>
                <Button size="sm" variant="success" onClick={() => updateSelectedChapterProgress('read')} disabled={selectedChapterIds.size === 0 || bulkProgressAction !== null}>
                  {bulkProgressAction === 'read' ? 'Marking read...' : 'Mark as read'}
                </Button>
                <Button size="sm" variant="outline-danger" onClick={() => updateSelectedChapterProgress('unread')} disabled={selectedChapterIds.size === 0 || bulkProgressAction !== null}>
                  {bulkProgressAction === 'unread' ? 'Marking unread...' : 'Mark as unread'}
                </Button>
              </>
            )}
          </div>
          <div className="chapter-list-container" aria-label="Chapter list">
          <ListGroup>
            {chapters.map((chapter) => {
              const isCompleted = completedChapterIds.has(chapter.id);
              const savedPage = chapterProgressPages[chapter.id];
              const showSavedPage = typeof savedPage === 'number' && savedPage > 0 && savedPage < MANUAL_COMPLETION_PAGE;
              return (
                <ListGroup.Item key={chapter.id} className="p-0">
                  {isSelectionMode ? (
                    <button
                      type="button"
                      className={`chapter-link p-3 w-100 text-start d-flex justify-content-between align-items-center gap-3 ${isCompleted ? 'chapter-read' : ''} ${selectedChapterIds.has(chapter.id) ? 'chapter-selected' : ''}`}
                      onClick={() => toggleChapterSelection(chapter.id)}
                      disabled={bulkProgressAction !== null}
                      aria-pressed={selectedChapterIds.has(chapter.id)}
                      aria-label={`Select chapter ${chapter.number}: ${chapter.title}${isCompleted ? ', read' : ''}`}
                    >
                      <span>Chapter {chapter.number} - {chapter.title}</span>
                      {showSavedPage && <span className="chapter-progress">Last read: Page {savedPage}</span>}
                    </button>
                  ) : (
                    <Link
                      to={`/manga/${mangaId}/chapter/${chapter.id}`}
                      state={{ chapterIds }}
                      className={`text-decoration-none chapter-link p-3 d-flex justify-content-between align-items-center gap-3 ${isCompleted ? 'chapter-read' : ''}`}
                      aria-label={`Chapter ${chapter.number}: ${chapter.title}${isCompleted ? ', read' : ''}`}
                    >
                      <span>Chapter {chapter.number} - {chapter.title}</span>
                      {showSavedPage && <span className="chapter-progress">Last read: Page {savedPage}</span>}
                    </Link>
                  )}
                </ListGroup.Item>
              );
            })}
          </ListGroup>
        </div>
        </div>
      )}
    </div>
  );
};

export default MangaDetailsPage;


