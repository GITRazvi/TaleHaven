import React, { useEffect, useMemo, useRef, useState } from 'react';
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

const MangaReaderPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const location = useLocation();
  const { mangaId = '', chapterId = '' } = useParams<{ mangaId: string; chapterId: string }>();
  const containerRef = useRef<HTMLDivElement>(null);
  const pageElementsRef = useRef<HTMLDivElement[]>([]);
  const hasScrolledToResume = useRef<boolean>(false);

  // Get chapter IDs from navigation state (passed from manga details page)
  const chapterIds = useMemo(() => {
    return (location.state as { chapterIds?: string[] })?.chapterIds || [];
  }, [location.state]);

  const [loadedPagesCount, setLoadedPagesCount] = useState<number>(0);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [resumePage, setResumePage] = useState<number>(1);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [loadingMessage, setLoadingMessage] = useState<string>('Loading chapter...');
  const [error, setError] = useState<string>('');
  const [saveStatus, setSaveStatus] = useState<string>('');
  const [prevChapterId, setPrevChapterId] = useState<string | null>(null);
  const [nextChapterId, setNextChapterId] = useState<string | null>(null);


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
      hasScrolledToResume.current = false;
      pageElementsRef.current = [];

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
        const url = `${baseUrl}/chapter/${encodeURIComponent(chapterId)}/pages/stream?quality=data`;

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

          img.src = imageUrl;
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
            placeholder.remove();
            img.style.display = 'block';
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

          // Append to container
          pageContainer.appendChild(placeholder);
          pageContainer.appendChild(img);

          if (containerRef.current) {
            containerRef.current.appendChild(pageContainer);
            pageElementsRef.current.push(pageContainer);
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

            // Load reading progress (requires JWT authentication)
            // API v3.0.0: Checks Redis first, falls back to PostgreSQL
            // 404 = no saved progress (normal for first-time readers)
            if (mangaId) {
              backend.getChapterProgress(chapterId, mangaId)
                .then((progress) => {
                  const initialPage = progress.data?.currentPage ?? 1;  // Renamed from lastPage in API v3.0.0
                  const clampedInitialPage = Math.max(1, Math.min(initialPage, pageElementsRef.current.length));
                  setCurrentPage(clampedInitialPage);
                  setResumePage(clampedInitialPage);
                  console.log('[Reader] Resumed from page:', clampedInitialPage);
                })
                .catch((err) => {
                  // 404 = no saved progress (normal for first-time readers)
                  // 401 = not authenticated (can still read without auth)
                  if (err?.status === 404) {
                    console.info('[Reader] No saved progress found - starting from page 1');
                  } else if (err?.status === 401) {
                    console.info('[Reader] Not authenticated - starting from page 1');
                  } else {
                    console.warn('[Reader] Error loading progress:', err);
                  }
                  setCurrentPage(1);
                  setResumePage(1);
                });
            } else {
              setCurrentPage(1);
              setResumePage(1);
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
  }, [backend, chapterId, mangaId]);

  // Load chapters for navigation
  useEffect(() => {
    if (!chapterId) return;

    // If we have chapter IDs from navigation state, use them directly
    if (chapterIds.length > 0) {
      const currentIndex = chapterIds.indexOf(chapterId);

      if (currentIndex > 0) {
        setPrevChapterId(chapterIds[currentIndex - 1]);
      } else {
        setPrevChapterId(null);
      }

      if (currentIndex >= 0 && currentIndex < chapterIds.length - 1) {
        setNextChapterId(chapterIds[currentIndex + 1]);
      } else {
        setNextChapterId(null);
      }

      return;
    }

    // Fallback: Try to load from backend if no state provided
    if (!mangaId) return;

    const loadChapters = async () => {
      try {
        const chaptersResponse = await backend.getMangaChapters(mangaId, {
          limit: 500,
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
    if (loadedPagesCount === 0 || resumePage < 1 || hasScrolledToResume.current) {
      return;
    }

    const targetElement = pageElementsRef.current[resumePage - 1];
    if (!targetElement) {
      return;
    }

    targetElement.scrollIntoView({ behavior: 'auto', block: 'start' });
    hasScrolledToResume.current = true;
  }, [loadedPagesCount, resumePage]);

  useEffect(() => {
    if (loadedPagesCount === 0) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const visibleEntries = entries.filter((entry) => entry.isIntersecting);
        if (visibleEntries.length === 0) {
          return;
        }

        const mostVisible = visibleEntries.reduce((best, entry) =>
          entry.intersectionRatio > best.intersectionRatio ? entry : best
        );

        const indexAttr = mostVisible.target.getAttribute('data-page-index');
        const index = indexAttr ? Number(indexAttr) : NaN;

        if (!Number.isNaN(index)) {
          setCurrentPage(index + 1);
        }
      },
      {
        root: null,
        threshold: [0.25, 0.5, 0.75],
      }
    );

    pageElementsRef.current.forEach((element) => {
      if (element) {
        observer.observe(element);
      }
    });

    return () => observer.disconnect();
  }, [loadedPagesCount]);

  // Auto-save reading progress to Redis (requires JWT authentication)
  // API v3.0.0: Writes to Redis cache only for fast response (~1-2ms)
  useEffect(() => {
    if (!mangaId || !chapterId || loadedPagesCount === 0) {
      return;
    }

    const timeout = setTimeout(async () => {
      try {
        await backend.saveReadingProgress({
          mangaId,
          chapterId,
          currentPage: currentPage,  // Renamed from lastPage in API v3.0.0
          totalPages: loadedPagesCount,
        });
        setSaveStatus('✓ Progress saved');
      } catch (err: any) {
        // 401 = not authenticated (user can still read without saving)
        if (err?.status === 401) {
          setSaveStatus('💡 Sign in to save progress');
        } else {
          console.warn('[Reader] Failed to save progress:', err);
          setSaveStatus('⚠ Could not save progress');
        }
      }
    }, 300);

    return () => clearTimeout(timeout);
  }, [backend, chapterId, currentPage, mangaId, loadedPagesCount]);

  // Sync progress to PostgreSQL when leaving chapter
  // API v3.0.0: Persists Redis cache to database for permanent storage
  useEffect(() => {
    if (!chapterId) return;

    // Sync on component unmount (user navigates away)
    return () => {
      backend.syncReadingProgress(chapterId).catch((err) => {
        // Silent fail - don't show error to user on page navigation
        if (err?.status !== 401) {
          console.warn('[Reader] Failed to sync progress to database:', err);
        }
      });
    };
  }, [backend, chapterId]);

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

  if (isLoading) {
    return (
      <div className="p-3">
        <Spinner animation="border" size="sm" className="me-2" />
        {loadingMessage}
      </div>
    );
  }

  if (error) {
    return <div className="p-3 text-danger">{error}</div>;
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
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h4 className="mb-0">Chapter Reader</h4>
        <Button variant="outline-secondary" size="sm" onClick={() => navigate(`/manga/${mangaId}`)}>
          Back to Manga
        </Button>
      </div>

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


