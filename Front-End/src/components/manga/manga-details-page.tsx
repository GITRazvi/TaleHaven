import React, { useEffect, useMemo, useState } from 'react';
import { Button, ListGroup, Spinner } from 'react-bootstrap';
import { Link, useNavigate, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';

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

  const [manga, setManga] = useState<MangaView | null>(null);
  const [coverUrl, setCoverUrl] = useState<string>('');
  const [coverError, setCoverError] = useState<boolean>(false);
  const [chapters, setChapters] = useState<MangaChapterItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

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

      try {
        console.log('[MangaDetails] Loading data for mangaId:', mangaId);

        const [mangaResponse, chaptersResponse, coverResponse] = await Promise.all([
          backend.getManga(mangaId),
          backend.getMangaChapters(mangaId, { limit: 100, offset: 0, translatedLanguage: ['en'] }),
          backend.getMangaCover(mangaId, '512').catch((err) => {
            console.warn('[MangaDetails] Cover fetch failed:', err);
            return { result: 'ok', coverUrl: '' };
          }),
        ]);

        console.log('[MangaDetails] Manga response:', mangaResponse);
        console.log('[MangaDetails] Chapters response:', chaptersResponse);
        console.log('[MangaDetails] Cover response:', coverResponse);

        const mappedManga = mapManga(mangaResponse.data);
        const mappedChapters = chaptersResponse.data.map(mapChapter).filter((item): item is MangaChapterItem => item !== null);

        console.log('[MangaDetails] Mapped manga:', mappedManga);
        console.log('[MangaDetails] Mapped chapters:', mappedChapters);

        setManga(mappedManga);
        setChapters(mappedChapters);
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

  return (
    <div className="p-3">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="mb-0">{manga?.title}</h2>
        <Button variant="outline-secondary" size="sm" onClick={() => navigate(-1)}>
          ← Back
        </Button>
      </div>

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
      {chapters.length === 0 ? (
        <div>No chapters available.</div>
      ) : (
        <ListGroup>
          {chapters.map((chapter) => (
            <ListGroup.Item key={chapter.id} className="p-0">
              <Link
                to={`/manga/${mangaId}/chapter/${chapter.id}`}
                state={{ chapterIds: chapters.map(ch => ch.id) }}
                className="text-decoration-none chapter-link p-3"
                style={{ display: 'block' }}
              >
                Chapter {chapter.number} - {chapter.title}
              </Link>
            </ListGroup.Item>
          ))}
        </ListGroup>
      )}
    </div>
  );
};

export default MangaDetailsPage;


