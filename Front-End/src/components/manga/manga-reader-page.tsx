import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button, Spinner } from 'react-bootstrap';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';

const normalizePagePath = (path: string): string => {
  // Convert Windows backslashes to forward slashes
  let normalized = path.replace(/\\/g, '/');

  // Remove leading /images/ or images/ prefix
  if (normalized.startsWith('/images/')) {
    normalized = normalized.replace(/^\/images\//, '');
  } else if (normalized.startsWith('images/')) {
    normalized = normalized.replace(/^images\//, '');
  }

  return normalized;
};

const readLocalizedText = (value: unknown): string => {
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

const readResponseField = (value: unknown, field: string): unknown => {
  const data = value as Record<string, unknown> | undefined;
  const attributes = data?.attributes as Record<string, unknown> | undefined;
  return data?.[field] ?? attributes?.[field];
};

type PendingProgress = {
  chapterId: string;
  isFinalPage: boolean;
  mangaId: string;
  page: number;
  session: number;
  totalPages: number;
};

const MangaReaderPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const location = useLocation();
  const { mangaId = '', chapterId = '' } = useParams<{ mangaId: string; chapterId: string }>();
  const containerRef = useRef<HTMLDivElement>(null);
  const pageElementsRef = useRef<HTMLDivElement[]>([]);
  const hasScrolledToResume = useRef<boolean>(false);
  const currentPageRef = useRef<number>(1);
  const chapterPageCountRef = useRef<number>(0);
  const progressResolvedRef = useRef<boolean>(false);
  const lastSavedPageRef = useRef<number | null>(null);
  const completionSyncedRef = useRef<boolean>(false);
  const pendingProgressRef = useRef<PendingProgress | null>(null);
  const isPersistingProgressRef = useRef<boolean>(false);
  const readerSessionRef = useRef<number>(0);

  // Get chapter IDs and an optional explicit starting page from navigation state.
  // An explicit start page is used by “Read from page 1” and takes precedence
  // over saved reading progress.
  const chapterIds = useMemo(() => {
    return (location.state as { chapterIds?: string[] })?.chapterIds || [];
  }, [location.state]);
  const startAtPage = useMemo(() => {
    const requestedPage = (location.state as { startAtPage?: unknown } | null)?.startAtPage;
    return typeof requestedPage === 'number' && Number.isInteger(requestedPage) && requestedPage >= 1
      ? requestedPage
      : null;
  }, [location.state]);

  const [loadedPagesCount, setLoadedPagesCount] = useState<number>(0);
  const [renderedPagesCount, setRenderedPagesCount] = useState<number>(0);
  const [resumePage, setResumePage] = useState<number>(0);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [loadingMessage, setLoadingMessage] = useState<string>('Loading chapter...');
  const [error, setError] = useState<string>('');
  const [saveStatus, setSaveStatus] = useState<string>('');
  const [prevChapterId, setPrevChapterId] = useState<string | null>(null);
  const [nextChapterId, setNextChapterId] = useState<string | null>(null);
  const [mangaTitle, setMangaTitle] = useState<string>('');
  const [chapterNumber, setChapterNumber] = useState<string>('');

  const preserveViewport = (applyChange: () => void) => {
    const viewportCenter = window.innerHeight / 2;
    const anchor = pageElementsRef.current
      .filter((element): element is HTMLDivElement => Boolean(element) && element.dataset.pageLoaded === 'true')
      .map((element) => ({
        element,
        bounds: element.getBoundingClientRect(),
      }))
      .filter(({ bounds }) => bounds.bottom > 0 && bounds.top < window.innerHeight)
      .sort((first, second) => {
        const firstDistance = Math.abs((first.bounds.top + first.bounds.bottom) / 2 - viewportCenter);
        const secondDistance = Math.abs((second.bounds.top + second.bounds.bottom) / 2 - viewportCenter);
        return firstDistance - secondDistance;
      })[0];

    if (!anchor) {
      applyChange();
      return;
    }

    const previousTop = anchor.element.getBoundingClientRect().top;
    applyChange();
    window.requestAnimationFrame(() => {
      if (!document.contains(anchor.element)) {
        return;
      }

      const layoutShift = anchor.element.getBoundingClientRect().top - previousTop;
      if (Math.abs(layoutShift) > 0.5) {
        window.scrollBy(0, layoutShift);
      }
    });
  };

  const flushProgressQueue = useCallback(async () => {
    if (isPersistingProgressRef.current) {
      return;
    }

    isPersistingProgressRef.current = true;
    try {
      while (pendingProgressRef.current) {
        const progress = pendingProgressRef.current;
        pendingProgressRef.current = null;

        try {
          await backend.saveReadingProgress({
            mangaId: progress.mangaId,
            chapterId: progress.chapterId,
            currentPage: progress.page,
            totalPages: progress.totalPages,
          });
          await backend.syncReadingProgress(progress.chapterId);

          if (progress.session === readerSessionRef.current) {
            setSaveStatus(progress.isFinalPage ? '✓ Chapter marked as read' : '✓ Progress saved');
          }
        } catch (err: any) {
          if (progress.session !== readerSessionRef.current) {
            continue;
          }

          if (progress.isFinalPage) {
            completionSyncedRef.current = false;
          }
          if (err?.status === 401) {
            setSaveStatus('💡 Sign in to save progress');
          } else {
            console.warn('[Reader] Failed to save progress:', err);
            setSaveStatus('⚠ Could not save progress');
          }
        }
      }
    } finally {
      isPersistingProgressRef.current = false;
    }
  }, [backend]);

  const persistPage = useCallback((page: number) => {
    if (
      !mangaId ||
      !chapterId ||
      chapterPageCountRef.current < 1 ||
      lastSavedPageRef.current === page
    ) {
      return;
    }

    const isFinalPage = page === chapterPageCountRef.current && !completionSyncedRef.current;
    if (isFinalPage) {
      completionSyncedRef.current = true;
    }

    lastSavedPageRef.current = page;
    pendingProgressRef.current = {
      mangaId,
      chapterId,
      page,
      totalPages: chapterPageCountRef.current,
      isFinalPage,
      session: readerSessionRef.current,
    };
    void flushProgressQueue();
  }, [chapterId, flushProgressQueue, mangaId]);

  useEffect(() => {
    let isCurrent = true;

    const loadReaderHeader = async () => {
      if (!mangaId || !chapterId) {
        return;
      }

      try {
        const [mangaResponse, chapterResponse] = await Promise.all([
          backend.getManga(mangaId),
          backend.getChapter(chapterId),
        ]);

        if (!isCurrent) {
          return;
        }

        const title = readLocalizedText(readResponseField(mangaResponse.data, 'title'));
        const number = readResponseField(chapterResponse.data, 'chapter');
        setMangaTitle(title || 'Untitled manga');
        setChapterNumber(number === undefined || number === null || number === '' ? '?' : String(number));
      } catch {
        if (isCurrent) {
          setMangaTitle('Manga reader');
          setChapterNumber('?');
        }
      }
    };

    loadReaderHeader();
    return () => {
      isCurrent = false;
    };
  }, [backend, mangaId, chapterId]);


  useEffect(() => {
    const loadPages = async () => {
      console.log('[Reader] loadPages called with chapterId:', chapterId, 'mangaId:', mangaId);

      if (!chapterId) {
        console.error('[Reader] Missing chapter id');
        setError('Missing chapter id.');
        setIsLoading(false);
        return;
      }

      console.log('[Reader] Starting page load process...');
      setIsLoading(true);
      setError('');
      setLoadedPagesCount(0);
      setRenderedPagesCount(0);
      chapterPageCountRef.current = 0;
      currentPageRef.current = 1;
      progressResolvedRef.current = false;
      lastSavedPageRef.current = null;
      completionSyncedRef.current = false;
      setResumePage(0);
      hasScrolledToResume.current = false;
      pageElementsRef.current = [];
      readerSessionRef.current += 1;

      // Clear container
      if (containerRef.current) {
        containerRef.current.innerHTML = '';
        console.log('[Reader] Container cleared');
      }

      setLoadingMessage('Fetching chapter pages from server...');

      try {
        setLoadingMessage('Downloading chapter images (this may take a moment)...');

        // @ts-ignore
        const baseUrl = backend.getBaseUrl();
        let requestedStartPage = startAtPage ?? 1;

        if (startAtPage === null && mangaId) {
          setLoadingMessage('Finding your last read page...');
          try {
            const progress = await backend.getChapterProgress(chapterId, mangaId);
            const savedPage = progress.data?.currentPage;
            if (typeof savedPage === 'number' && savedPage >= 1) {
              requestedStartPage = savedPage;
            }
          } catch (err: any) {
            if (err?.status !== 404 && err?.status !== 401) {
              console.warn('[Reader] Error loading progress:', err);
            }
          }
        }

        const url = `${baseUrl}/chapter/${encodeURIComponent(chapterId)}/pages/stream?quality=data&startPage=${requestedStartPage}`;

        console.log('[Reader] Opening SSE connection:', url);
        console.log('[Reader] Base URL:', baseUrl);
        console.log('[Reader] Chapter ID:', chapterId);

        let eventSource: EventSource;
        try {
          eventSource = new EventSource(url, { withCredentials: true });
          console.log('[Reader] EventSource created successfully, readyState:', eventSource.readyState);
        } catch (constructorError) {
          console.error('[Reader] Failed to create EventSource:', constructorError);
          throw new Error(`Failed to create SSE connection: ${constructorError}`);
        }

        // Add open handler to confirm connection
        eventSource.addEventListener('open', () => {
          console.log('[Reader] SSE connection opened successfully');
        });

        // Add a timeout to check if connection is stuck
        const connectionTimeout = setTimeout(() => {
          console.warn('[Reader] SSE connection taking longer than expected. ReadyState:', eventSource.readyState);
          if (eventSource.readyState === EventSource.CONNECTING) {
            console.warn('[Reader] Still connecting after 5 seconds - possible backend issue');
          }
        }, 5000);

        // Start event - chapter metadata
        eventSource.addEventListener('start', (e) => {
          clearTimeout(connectionTimeout);
          const data = JSON.parse((e as MessageEvent).data);
          console.log('[Reader] SSE start event:', data);
          const totalPages = Number(data.totalPages);
          if (Number.isInteger(totalPages) && totalPages > 0) {
            chapterPageCountRef.current = totalPages;
          }

          const resolvedStartPage = Number(data.startPage);
          const initialPage = Number.isInteger(resolvedStartPage) && resolvedStartPage > 0
            ? resolvedStartPage
            : requestedStartPage;
          const setInitialPage = (page: number) => {
            const clampedPage = Number.isInteger(totalPages) && totalPages > 0
              ? Math.max(1, Math.min(page, totalPages))
              : Math.max(1, page);
            currentPageRef.current = clampedPage;
            setResumePage(clampedPage);
            progressResolvedRef.current = true;
          };

          setInitialPage(initialPage);

          setIsLoading(false); // Stop showing loading spinner, start showing container
          setLoadingMessage(`Receiving chapter data (${data.totalPages || 0} pages)...`);
        });

        // Page event - individual page ready
        eventSource.addEventListener('page', (e) => {
          const data = JSON.parse((e as MessageEvent).data);
          const pageNumber = data.pageNumber;
          const pageUrl = data.pageUrl;

          console.log('[Reader] Page received:', pageNumber, 'URL:', pageUrl);

          // Create page container
          const pageContainer = document.createElement('div');
          pageContainer.style.position = 'relative';
          pageContainer.style.width = '100%';
          pageContainer.style.maxWidth = '980px';
          pageContainer.style.marginBottom = '1rem';
          pageContainer.setAttribute('data-page-index', String(pageNumber - 1));

          // Create loading placeholder
          const placeholder = document.createElement('div');
          placeholder.style.position = 'absolute';
          placeholder.style.top = '0';
          placeholder.style.left = '0';
          placeholder.style.right = '0';
          placeholder.style.height = '600px';
          placeholder.style.backgroundColor = '#f0f0f0';
          placeholder.style.display = 'flex';
          placeholder.style.alignItems = 'center';
          placeholder.style.justifyContent = 'center';
          placeholder.style.color = '#666';
          placeholder.innerHTML = `
            <div>
              <span class="spinner-border spinner-border-sm me-2" role="status"></span>
              Loading page ${pageNumber}...
            </div>
          `;

          // Create image
          const img = document.createElement('img');
          const normalizedPath = normalizePagePath(pageUrl);
          const imageUrl = backend.getImageUrl(normalizedPath);

          img.alt = `Chapter page ${pageNumber}`;
          img.style.width = '100%';
          img.style.maxWidth = '980px';
          img.style.display = 'none';
          img.loading = 'eager'; // Load all images immediately since they're streaming progressively
          img.setAttribute('data-page-index', String(pageNumber - 1));

          // Timeout check: if image takes too long (15 seconds), show warning
          const timeoutId = setTimeout(() => {
            if (img.style.display === 'none') {
              placeholder.innerHTML = `
                <div style="color: #ff9800;">
                  <span class="spinner-border spinner-border-sm me-2"></span>
                  Still downloading page ${pageNumber}...<br/>
                  <small>This may take a moment</small>
                </div>
              `;
            }
          }, 15000);

          // Image load handler
          img.onload = () => {
            clearTimeout(timeoutId); // Cancel timeout
            preserveViewport(() => {
              placeholder.remove();
              img.style.display = 'block';
              pageContainer.dataset.pageLoaded = 'true';
            });
            setRenderedPagesCount((previous) => previous + 1);
            console.log('[Reader] Image loaded successfully:', pageNumber, imageUrl);
          };

          // Image error handler
          img.onerror = (event) => {
            clearTimeout(timeoutId); // Cancel timeout
            console.error(`[Reader] Image ${pageNumber} failed to load:`, imageUrl, event);
            placeholder.innerHTML = `
              <div style="color: #dc3545;">
                Failed to load page ${pageNumber}<br/>
                <small>${imageUrl}</small><br/>
                <button onclick="location.reload()" class="btn btn-sm btn-outline-danger mt-2">Retry</button>
              </div>
            `;
          };

          // Attach load handlers before assigning the source. A cached image can
          // otherwise finish before onload is registered and never update progress.
          img.src = imageUrl;

          // Append to container
          pageContainer.appendChild(placeholder);
          pageContainer.appendChild(img);

          if (containerRef.current) {
            const followingPage = Array.from(containerRef.current.children).find((element) => {
              const pageIndex = Number(element.getAttribute('data-page-index'));
              return Number.isInteger(pageIndex) && pageIndex > pageNumber - 1;
            });
            preserveViewport(() => {
              containerRef.current?.insertBefore(pageContainer, followingPage ?? null);
              pageElementsRef.current[pageNumber - 1] = pageContainer;
            });
          }

          setLoadedPagesCount((prev) => prev + 1);
        });

        // Page error event
        eventSource.addEventListener('page_error', (e) => {
          console.error('[Reader] Page error:', e);
          const data = JSON.parse((e as MessageEvent).data);
          console.error('[Reader] Page error details:', data);
        });

          // Complete event
          eventSource.addEventListener('complete', (e) => {
            clearTimeout(connectionTimeout);
            const data = JSON.parse((e as MessageEvent).data);
            console.log('[Reader] SSE complete event:', data);
            eventSource.close();

            // Check if we actually received any pages
            if (data.totalPages === 0 || data.successCount === 0) {
              console.error('[Reader] No pages received from backend!', data);
              setError(`No pages available for this chapter. The backend may have failed to download the chapter from MangaDex. Total pages: ${data.totalPages}, Success: ${data.successCount}`);
              setIsLoading(false);
              return;
            }

            console.log('[Reader] Successfully loaded', data.successCount, 'of', data.totalPages, 'pages');
            const totalPages = Number(data.totalPages);
            if (Number.isInteger(totalPages) && totalPages > 0) {
              chapterPageCountRef.current = totalPages;
            }

          });

        // Error event - connection error
        eventSource.onerror = (err) => {
          clearTimeout(connectionTimeout);
          console.error('[Reader] SSE connection error:', err);
          console.error('[Reader] EventSource readyState:', eventSource.readyState);
          console.error('[Reader] EventSource url:', eventSource.url);

          // Provide more specific error message
          let errorMsg = 'Connection lost. Please try again.';
          if (eventSource.readyState === EventSource.CLOSED) {
            errorMsg = 'Connection closed by server. The chapter may not exist or the server is unavailable.';
          } else if (eventSource.readyState === EventSource.CONNECTING) {
            errorMsg = 'Attempting to reconnect...';
          }

          setError(errorMsg);
          setIsLoading(false);
          eventSource.close();
        };

        // Cleanup on unmount
        return () => {
          clearTimeout(connectionTimeout);
          console.log('[Reader] Cleaning up SSE connection');
          eventSource.close();
        };
      } catch (err) {
        console.error('[Reader] Error setting up SSE:', err);
        setError((err as Error).message || 'Failed to load chapter pages.');
        setIsLoading(false);
      }
    };

    const cleanup = loadPages();
    return () => {
      if (cleanup instanceof Promise) {
        cleanup.then((fn) => fn && fn());
      }
    };
  }, [backend, chapterId, mangaId, startAtPage]);

  // Load chapters for navigation
  useEffect(() => {
    if (!chapterId) return;

    // If we have chapter IDs from navigation state, use them directly
    if (chapterIds.length > 0) {
      const currentIndex = chapterIds.indexOf(chapterId);

      if (currentIndex >= 0) {
        if (currentIndex > 0) {
          setPrevChapterId(chapterIds[currentIndex - 1]);
        } else {
          setPrevChapterId(null);
        }

        if (currentIndex < chapterIds.length - 1) {
          setNextChapterId(chapterIds[currentIndex + 1]);
        } else {
          setNextChapterId(null);
        }

        return;
      }
    }

    // Fallback: Try to load from backend if no state provided
    if (!mangaId) return;

    const loadChapters = async () => {
      try {
        const chaptersResponse = await backend.getMangaChapters(mangaId, {
          limit: 100,
          offset: 0,
          translatedLanguage: ['en']
        });

        const loadedChapterIds = chaptersResponse.data.map((ch: any) => ch.id);
        const currentIndex = loadedChapterIds.indexOf(chapterId);

        if (currentIndex > 0) {
          setPrevChapterId(loadedChapterIds[currentIndex - 1]);
        } else {
          setPrevChapterId(null);
        }

        if (currentIndex >= 0 && currentIndex < loadedChapterIds.length - 1) {
          setNextChapterId(loadedChapterIds[currentIndex + 1]);
        } else {
          setNextChapterId(null);
        }
      } catch (err) {
        console.warn('[Reader] Could not load chapter list from backend. Navigation buttons disabled.');
        // Keep buttons disabled if we can't load chapters
        setPrevChapterId(null);
        setNextChapterId(null);
      }
    };

    loadChapters();
  }, [backend, mangaId, chapterId, chapterIds]);

  useEffect(() => {
    if (loadedPagesCount === 0 || resumePage < 1 || hasScrolledToResume.current || !progressResolvedRef.current) {
      return;
    }

    const targetElement = pageElementsRef.current[resumePage - 1];
    if (!targetElement) {
      return;
    }

    targetElement.scrollIntoView({ behavior: 'auto', block: 'start' });
    currentPageRef.current = resumePage;
    hasScrolledToResume.current = true;
    persistPage(resumePage);
  }, [loadedPagesCount, persistPage, resumePage]);

  useEffect(() => {
    if (renderedPagesCount === 0) {
      return;
    }

    let animationFrame: number | null = null;
    const trackVisiblePage = () => {
      animationFrame = null;
      const viewportCenter = window.innerHeight / 2;
      const visiblePages = pageElementsRef.current
        .filter((element): element is HTMLDivElement => Boolean(element) && element.dataset.pageLoaded === 'true')
        .map((element) => ({
          element,
          page: Number(element.getAttribute('data-page-index')) + 1,
          bounds: element.getBoundingClientRect(),
        }))
        .filter(({ page, bounds }) => Number.isInteger(page) && bounds.bottom > 0 && bounds.top < window.innerHeight);

      if (visiblePages.length === 0) {
        return;
      }

      const currentPage = visiblePages.reduce((closest, candidate) => {
        const closestDistance = Math.abs((closest.bounds.top + closest.bounds.bottom) / 2 - viewportCenter);
        const candidateDistance = Math.abs((candidate.bounds.top + candidate.bounds.bottom) / 2 - viewportCenter);
        return candidateDistance < closestDistance ? candidate : closest;
      });

      currentPageRef.current = currentPage.page;
      if (hasScrolledToResume.current) {
        persistPage(currentPage.page);
      }
    };

    const requestTracking = () => {
      if (animationFrame === null) {
        animationFrame = window.requestAnimationFrame(trackVisiblePage);
      }
    };

    requestTracking();
    window.addEventListener('scroll', requestTracking, { passive: true });
    window.addEventListener('resize', requestTracking);

    return () => {
      if (animationFrame !== null) {
        window.cancelAnimationFrame(animationFrame);
      }
      window.removeEventListener('scroll', requestTracking);
      window.removeEventListener('resize', requestTracking);
    };
  }, [persistPage, renderedPagesCount]);

  // Queue the exact active chapter position before leaving it. The queue
  // serializes writes so it cannot be overwritten by an earlier page request.
  useEffect(() => {
    if (!chapterId) return;

    return () => {
      const currentChapterPage = currentPageRef.current;
      const totalChapterPages = chapterPageCountRef.current;

      if (!mangaId || totalChapterPages < 1 || !hasScrolledToResume.current) {
        return;
      }

      lastSavedPageRef.current = null;
      persistPage(currentChapterPage);
    };
  }, [chapterId, mangaId, persistPage]);

  // Sync progress to PostgreSQL on tab close/refresh
  // API v3.0.0: Ensures progress is saved to database before leaving
  useEffect(() => {
    const handleBeforeUnload = () => {
      if (chapterId) {
        // Use sendBeacon for reliable delivery during page unload
        const url = `${backend.getBaseUrl()}/reading-progress/sync`;
        const data = JSON.stringify({ chapterId });

        // Try sendBeacon first (most reliable)
        if (navigator.sendBeacon) {
          const blob = new Blob([data], { type: 'application/json' });
          navigator.sendBeacon(url, blob);
        } else {
          // Fallback: synchronous request (less reliable but better than nothing)
          backend.syncReadingProgress(chapterId).catch(() => {});
        }
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [backend, chapterId]);

  const ReaderHeader = () => (
    <div className="d-flex justify-content-between align-items-center mb-3">
      <h4 className="mb-0">
        {mangaTitle || 'Loading manga...'}
        {chapterNumber && ` — Chapter ${chapterNumber}`}
      </h4>
      <Button variant="outline-secondary" size="sm" onClick={() => navigate(`/manga/${mangaId}`)}>
        Back to Manga
      </Button>
    </div>
  );

  if (isLoading) {
    return (
      <div className="p-3">
        <ReaderHeader />
        <Spinner animation="border" size="sm" className="me-2" />
        {loadingMessage}
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-3">
        <ReaderHeader />
        <div className="text-danger">{error}</div>
      </div>
    );
  }


  const ChapterNavigation = () => (
    <div className="d-flex justify-content-between gap-2 mb-3">
      <Button
        variant="outline-secondary"
        disabled={!prevChapterId}
        onClick={() => {
          if (prevChapterId) {
            navigate(`/manga/${mangaId}/chapter/${prevChapterId}`, {
              state: { chapterIds }
            });
          }
        }}
      >
        ← Previous Chapter
      </Button>
      <Button
        variant="outline-secondary"
        disabled={!nextChapterId}
        onClick={() => {
          if (nextChapterId) {
            navigate(`/manga/${mangaId}/chapter/${nextChapterId}`, {
              state: { chapterIds }
            });
          }
        }}
      >
        Next Chapter →
      </Button>
    </div>
  );

  return (
    <div className="p-3">
      <ReaderHeader />

      <ChapterNavigation />

      {saveStatus && <div className="text-muted mb-2">{saveStatus}</div>}

      {/* Container for pages - pages are appended via DOM manipulation */}
      <div
        ref={containerRef}
        className="d-flex flex-column align-items-center"
        style={{ gap: '1rem' }}
      />

      {loadedPagesCount === 0 && !isLoading && (
        <div>No pages found for this chapter.</div>
      )}

      {loadedPagesCount > 0 && <ChapterNavigation />}
    </div>
  );
};

export default MangaReaderPage;


