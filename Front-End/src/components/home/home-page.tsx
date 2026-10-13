import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Container, Row, Spinner, Alert } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { PopularManga } from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';

const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();

  const [popularManga, setPopularManga] = useState<PopularManga[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [coverErrors, setCoverErrors] = useState<Set<string>>(new Set());
  const [showDebug, setShowDebug] = useState<boolean>(false);

  useEffect(() => {
    const loadPopularManga = async () => {
      setIsLoading(true);
      setError('');

      try {
        console.log('[HomePage] Fetching popular manga...');
        const response = await backend.getPopularManga({ limit: 20, orderBy: 'followedCount' });
        console.log('[HomePage] Popular manga response:', response);
        console.log('[HomePage] First manga:', response.data[0]);
        setPopularManga(response.data || []);
      } catch (err) {
        console.error('[HomePage] Error loading popular manga:', err);
        setError((err as Error).message || 'Failed to load popular manga.');
      } finally {
        setIsLoading(false);
      }
    };

    loadPopularManga();
  }, [backend]);

  const handleCoverError = (mangaId: string, coverUrl: string) => {
    console.error('[HomePage] Cover image failed to load:', {
      mangaId,
      coverUrl,
      timestamp: new Date().toISOString()
    });
    setCoverErrors((prev) => new Set(prev).add(mangaId));
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
          <Row>
            {popularManga.map((manga) => (
              <Col key={manga.id} xs={6} sm={4} md={3} lg={2} className="mb-4">
                <Card
                  style={{ height: '100%', cursor: 'pointer' }}
                  onClick={() => navigate(`/manga/${manga.id}`)}
                  className="manga-library-card"
                >
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

