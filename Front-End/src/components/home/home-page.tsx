import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Container, Row, Spinner, Alert } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { PopularManga } from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';

const POPULAR_PAGE_SIZE = 20;

const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();

  const [popularManga, setPopularManga] = useState<PopularManga[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isLoadingMorePopular, setIsLoadingMorePopular] = useState(false);
  const [hasMorePopular, setHasMorePopular] = useState(false);
  const [error, setError] = useState<string>('');
  const [loadMorePopularError, setLoadMorePopularError] = useState('');
  const [coverErrors, setCoverErrors] = useState<Set<string>>(new Set());
  const [showDebug, setShowDebug] = useState<boolean>(false);
  const [libraryMangaIds, setLibraryMangaIds] = useState<Set<string>>(new Set());
  const [libraryActionId, setLibraryActionId] = useState<string | null>(null);
  const [libraryError, setLibraryError] = useState<string>('');

  useEffect(() => {
    const loadPopularManga = async () => {
      setIsLoading(true);
      setError('');

      try {
        console.log('[HomePage] Fetching popular manga...');
        const response = await backend.getPopularManga({ limit: POPULAR_PAGE_SIZE, offset: 0, orderBy: 'followedCount' });
        console.log('[HomePage] Popular manga response:', response);
        console.log('[HomePage] First manga:', response.data[0]);
        const firstPage = response.data || [];
        const total = response.total ?? firstPage.length;
        setPopularManga(firstPage);
        setHasMorePopular(firstPage.length < total);
      } catch (err) {
        console.error('[HomePage] Error loading popular manga:', err);
        setError((err as Error).message || 'Failed to load popular manga.');
      } finally {
        setIsLoading(false);
      }
    };

    loadPopularManga();
  }, [backend]);

  const handleLoadMorePopular = async () => {
    setLoadMorePopularError('');
    setIsLoadingMorePopular(true);
    const offset = popularManga.length;

    try {
      const response = await backend.getPopularManga({
        limit: POPULAR_PAGE_SIZE,
        offset,
        orderBy: 'followedCount',
      });
      const nextPage = response.data || [];
      const total = response.total ?? offset + nextPage.length;

      setPopularManga((currentManga) => {
        const knownMangaIds = new Set(currentManga.map((manga) => manga.id));
        return [...currentManga, ...nextPage.filter((manga) => !knownMangaIds.has(manga.id))];
      });
      setHasMorePopular(offset + nextPage.length < total);
    } catch (err) {
      setLoadMorePopularError((err as Error).message || 'Unable to load more popular manga. Please try again.');
    } finally {
      setIsLoadingMorePopular(false);
    }
  };

  useEffect(() => {
    let isCurrent = true;

    if (!user) {
      setLibraryMangaIds(new Set());
      return () => {
        isCurrent = false;
      };
    }

    const loadLibrary = async () => {
      try {
        const response = await backend.getMyLibrary();
        if (isCurrent) {
          setLibraryMangaIds(new Set(response.data.map((manga) => manga.mangaId)));
        }
      } catch {
        // The cards remain usable even if the saved-state lookup fails.
        if (isCurrent) {
          setLibraryMangaIds(new Set());
        }
      }
    };

    loadLibrary();
    return () => {
      isCurrent = false;
    };
  }, [backend, user]);

  const handleCoverError = (mangaId: string, coverUrl: string) => {
    console.error('[HomePage] Cover image failed to load:', {
      mangaId,
      coverUrl,
      timestamp: new Date().toISOString()
    });
    setCoverErrors((prev) => new Set(prev).add(mangaId));
  };

  const handleLibraryToggle = async (event: React.MouseEvent<HTMLButtonElement>, mangaId: string) => {
    event.stopPropagation();

    if (!user) {
      navigate('/login');
      return;
    }

    const isSaved = libraryMangaIds.has(mangaId);
    setLibraryError('');
    setLibraryActionId(mangaId);
    try {
      if (isSaved) {
        await backend.removeMangaFromMyLibrary(mangaId);
      } else {
        await backend.addMangaToMyLibrary(mangaId);
      }

      setLibraryMangaIds((currentIds) => {
        const nextIds = new Set(currentIds);
        if (isSaved) {
          nextIds.delete(mangaId);
        } else {
          nextIds.add(mangaId);
        }
        return nextIds;
      });
    } catch (err) {
      setLibraryError((err as Error).message || 'Unable to update your library. Please try again.');
    } finally {
      setLibraryActionId(null);
    }
  };

  return (
    <Container className="p-5">
      <div className="text-center mb-5">
        <h1 className="display-4 mb-3">
          Welcome{user ? ` back, ${user.email.split('@')[0]}` : ' to TaleHaven'}!
        </h1>
        <p className="lead text-muted">Your manga reading platform</p>
        {user && (
          <Alert variant="info" className="mt-3 d-inline-block">
            🎉 Your reading progress is now being saved to your account!
          </Alert>
        )}
      </div>

      <div className="mb-4">
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h2>Popular Manga</h2>
          <div className="d-flex gap-2">
            <Button onClick={() => setShowDebug(!showDebug)} variant="outline-secondary" size="sm">
              {showDebug ? 'Hide' : 'Show'} Debug
            </Button>
            <Button onClick={() => navigate('/manga/search')} variant="primary">
              Browse More Manga
            </Button>
          </div>
        </div>

        {showDebug && !isLoading && popularManga.length > 0 && (
          <div className="mb-3 p-3 border rounded" style={{ backgroundColor: '#f8f9fa', fontSize: '0.85rem', maxHeight: '200px', overflow: 'auto' }}>
            <strong>Debug Info:</strong>
            <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all', margin: '0.5rem 0 0 0' }}>
              {JSON.stringify(popularManga.slice(0, 2), null, 2)}
            </pre>
          </div>
        )}

        {isLoading && (
          <div className="text-center p-5">
            <Spinner animation="border" size="sm" className="me-2" />
            Loading popular manga...
          </div>
        )}

        {error && (
          <div className="text-danger mb-3">Error: {error}</div>
        )}

        {loadMorePopularError && (
          <Alert variant="danger">{loadMorePopularError}</Alert>
        )}

        {libraryError && (
          <Alert variant="danger">{libraryError}</Alert>
        )}

        {!isLoading && !error && (!popularManga || popularManga.length === 0) && (
          <Card className="text-center p-5">
            <Card.Body>
              <h4 className="mb-3">No popular manga available</h4>
              <p className="text-muted mb-4">
                Unable to load popular manga at this time. Please try again later.
              </p>
              <Button onClick={() => navigate('/manga/search')} variant="primary">
                Search Manga
              </Button>
            </Card.Body>
          </Card>
        )}

        {!isLoading && popularManga && popularManga.length > 0 && (
          <>
            <Row>
              {popularManga.map((manga) => (
                <Col key={manga.id} xs={6} sm={4} md={3} lg={2} className="mb-4">
                  <Card
                    style={{ height: '100%', cursor: 'pointer' }}
                    onClick={() => navigate(`/manga/${manga.id}`)}
                    className="manga-library-card"
                  >
                    <div className="manga-card-cover">
                      {!coverErrors.has(manga.id) && manga.coverUrl && (
                        <Card.Img
                          variant="top"
                          src={manga.coverUrl}
                          alt={manga.title}
                          crossOrigin="anonymous"
                          referrerPolicy="no-referrer"
                          onError={() => handleCoverError(manga.id, manga.coverUrl)}
                          onLoad={() => console.log('[HomePage] Cover loaded:', manga.title)}
                          style={{
                            height: '280px',
                            objectFit: 'cover',
                            width: '100%'
                          }}
                        />
                      )}
                      {(coverErrors.has(manga.id) || !manga.coverUrl) && (
                        <div
                          style={{
                            height: '280px',
                            backgroundColor: '#f0f0f0',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: '#999'
                          }}
                        >
                          No Cover
                        </div>
                      )}
                      <button
                        type="button"
                        className={`home-library-heart ${libraryMangaIds.has(manga.id) ? 'is-saved' : ''}`}
                        onClick={(event) => handleLibraryToggle(event, manga.id)}
                        disabled={libraryActionId === manga.id}
                        aria-label={libraryMangaIds.has(manga.id) ? `Remove ${manga.title} from My Library` : user ? `Add ${manga.title} to My Library` : `Log in to save ${manga.title}`}
                        aria-pressed={libraryMangaIds.has(manga.id)}
                        title={libraryMangaIds.has(manga.id) ? 'Remove from My Library' : user ? 'Add to My Library' : 'Log in to save'}
                      >
                        {libraryActionId === manga.id ? '…' : libraryMangaIds.has(manga.id) ? '♥' : '♡'}
                      </button>
                    </div>
                    <Card.Body className="p-2">
                      <Card.Text
                        className="small text-center mb-0"
                        style={{
                          fontSize: '0.85rem',
                          lineHeight: '1.2',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical'
                        }}
                        title={manga.title}
                      >
                        {manga.title}
                      </Card.Text>
                      {showDebug && (
                        <div className="mt-2" style={{ fontSize: '0.7rem', color: '#666' }}>
                          <div>ID: {manga.id.substring(0, 8)}...</div>
                          <div>Cover: {manga.coverUrl ? '✓' : '✗'}</div>
                          {manga.coverUrl && <div className="text-truncate" title={manga.coverUrl}>URL: {manga.coverUrl.substring(0, 40)}...</div>}
                        </div>
                      )}
                    </Card.Body>
                  </Card>
                </Col>
              ))}
            </Row>
            {hasMorePopular && (
              <div className="text-center mt-2">
                <Button variant="outline-primary" onClick={handleLoadMorePopular} disabled={isLoadingMorePopular}>
                  {isLoadingMorePopular ? 'Loading...' : 'Load more popular manga'}
                </Button>
              </div>
            )}
          </>
        )}
      </div>

      {!isLoading && popularManga && popularManga.length > 0 && (
        <Row className="justify-content-center mt-5">
          <Col md={8} lg={6}>
            <Card className="h-100 shadow-sm">
              <Card.Body className="d-flex flex-column text-center">
                <Card.Title className="h3 mb-3">📖 Manga Library</Card.Title>
                <Card.Text className="mb-4">
                  Discover and read manga with our infinite scroll reader. Track your reading progress
                  across chapters, download for offline reading, and enjoy seamless navigation through
                  your favorite series.
                </Card.Text>
                <Button onClick={() => navigate('/manga/search')} variant="primary" size="lg" className="mt-auto">
                  Start Reading
                </Button>
              </Card.Body>
            </Card>
          </Col>
        </Row>
      )}
    </Container>
  );
};

export default HomePage;

