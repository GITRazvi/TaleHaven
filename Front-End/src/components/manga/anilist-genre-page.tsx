import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Container, Row, Spinner } from 'react-bootstrap';
import { useNavigate, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { AniListManga } from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';

const PAGE_SIZE = 12;

const titleOf = (manga: AniListManga) => (
  manga.title.english || manga.title.userPreferred || manga.title.romaji || manga.title.native || 'Untitled manga'
);

const AniListGenrePage: React.FC = () => {
  const { genre: routeGenre } = useParams();
  const navigate = useNavigate();
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();
  const genre = (routeGenre || '').trim();
  const [manga, setManga] = useState<AniListManga[]>([]);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [coverErrors, setCoverErrors] = useState<Set<number>>(new Set());
  const [savedAniListIDs, setSavedAniListIDs] = useState<Set<number>>(new Set());
  const [libraryActionID, setLibraryActionID] = useState<number | null>(null);
  const [favoriteAniListIDs, setFavoriteAniListIDs] = useState<Set<number>>(new Set());
  const [favoriteActionID, setFavoriteActionID] = useState<number | null>(null);

  const loadPage = async (nextPage: number, append: boolean) => {
    if (!genre) {
      setLoading(false);
      setError('Invalid genre.');
      return;
    }
    if (append) setLoadingMore(true); else setLoading(true);
    setError('');
    try {
      const response = await backend.getAniListMangaByGenre(genre, nextPage, PAGE_SIZE);
      setManga((current) => append
        ? [...current, ...response.data.filter((item) => !current.some((known) => known.id === item.id))]
        : response.data);
      setPage(response.page || nextPage);
      setHasNextPage(response.hasNextPage);
    } catch (requestError) {
      setError((requestError as Error).message || 'Unable to load manga for this genre. Please try again.');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  useEffect(() => { void loadPage(1, false); /* backend is memoized for this component instance. */ // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [genre, backend]);

  useEffect(() => {
    if (!user) { setSavedAniListIDs(new Set()); setFavoriteAniListIDs(new Set()); return; }
    void backend.getAniListMyLibrary({ limit: 100, offset: 0 })
      .then((response) => setSavedAniListIDs(new Set(response.data.map((item) => item.anilistId))))
      .catch(() => setSavedAniListIDs(new Set()));
    void backend.getAniListFavoriteIDs()
      .then((response) => setFavoriteAniListIDs(new Set(response.data)))
      .catch(() => setFavoriteAniListIDs(new Set()));
  }, [backend, user]);

  const toggleLibrary = async (event: React.MouseEvent<HTMLButtonElement>, selectedManga: AniListManga) => {
    event.stopPropagation();
    if (!user) { navigate('/login'); return; }
    const saved = savedAniListIDs.has(selectedManga.id);
    setLibraryActionID(selectedManga.id);
    try {
      if (saved) await backend.removeAniListMangaFromMyLibrary(selectedManga.id);
      else await backend.addAniListMangaToMyLibrary(selectedManga.id);
      setSavedAniListIDs((current) => {
        const next = new Set(current);
        if (saved) next.delete(selectedManga.id); else next.add(selectedManga.id);
        return next;
      });
    } finally {
      setLibraryActionID(null);
    }
  };

  const toggleFavorite = async (event: React.MouseEvent<HTMLButtonElement>, selectedManga: AniListManga) => {
    event.stopPropagation();
    if (!user) { navigate('/login'); return; }
    const favorite = favoriteAniListIDs.has(selectedManga.id);
    setFavoriteActionID(selectedManga.id);
    try {
      if (favorite) await backend.removeAniListMangaFromFavorites(selectedManga.id);
      else await backend.addAniListMangaToFavorites(selectedManga.id);
      setFavoriteAniListIDs((current) => { const next = new Set(current); if (favorite) next.delete(selectedManga.id); else next.add(selectedManga.id); return next; });
    } finally { setFavoriteActionID(null); }
  };

  return (
    <Container className="py-5">
      <Button variant="outline-secondary" className="mb-4" onClick={() => navigate(-1)}>Back</Button>
      <h1 className="mb-4">Manga in {genre || 'this genre'}</h1>

      {loading && <div className="text-center p-4"><Spinner animation="border" size="sm" className="me-2" />Loading manga...</div>}
      {!loading && error && <Alert variant="danger">{error}</Alert>}
      {!loading && !error && manga.length === 0 && <Alert variant="secondary">No manga is available for {genre} right now.</Alert>}

      {!loading && manga.length > 0 && <>
        <Row>
          {manga.map((item) => {
            const title = titleOf(item);
            const cover = item.coverImage.large || item.coverImage.medium;
            const hasCoverError = coverErrors.has(item.id);
            const isSaved = savedAniListIDs.has(item.id);
            const isFavorite = favoriteAniListIDs.has(item.id);
            return <Col key={item.id} xs={6} sm={4} md={3} lg={2} className="mb-4">
              <Card className="manga-library-card h-100 shadow-sm" role="button" onClick={() => navigate(`/anilist/manga/${item.id}`)}>
                <div className="position-relative">
                  {!hasCoverError && cover ? <Card.Img variant="top" src={cover} alt={title} referrerPolicy="no-referrer" onError={() => setCoverErrors((current) => new Set(current).add(item.id))} style={{ height: '260px', objectFit: 'cover', width: '100%' }} /> : <div className="d-flex align-items-center justify-content-center text-muted" style={{ height: '260px', backgroundColor: item.coverImage.color || '#f0f0f0' }}>No Cover</div>}
                  <button type="button" className="home-library-toggle" onClick={(event) => void toggleLibrary(event, item)} disabled={libraryActionID === item.id} aria-label={isSaved ? `Remove ${title} from My Library` : `Add ${title} to My Library`}>
                    {libraryActionID === item.id ? '…' : isSaved ? '−' : '+'}
                  </button>
                  {isSaved && <button type="button" className={`home-library-heart ${isFavorite ? 'is-saved' : ''}`} onClick={(event) => void toggleFavorite(event, item)} disabled={favoriteActionID === item.id} aria-label={isFavorite ? `Remove ${title} from Favorites` : `Add ${title} to Favorites`}>
                    {favoriteActionID === item.id ? '…' : isFavorite ? '♥' : '♡'}
                  </button>}
                </div>
                <Card.Body className="p-2 d-flex flex-column">
                  <Card.Title className="small text-center mb-1" title={title} style={{ overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>{title}</Card.Title>
                  {item.averageScore > 0 && <Card.Text className="small text-muted text-center mb-0">★ {item.averageScore}%</Card.Text>}
                </Card.Body>
              </Card>
            </Col>;
          })}
        </Row>
        {hasNextPage && <div className="text-center"><Button variant="outline-primary" onClick={() => void loadPage(page + 1, true)} disabled={loadingMore}>{loadingMore ? 'Loading...' : 'Load more'}</Button></div>}
      </>}
    </Container>
  );
};

export default AniListGenrePage;
