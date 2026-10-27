import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Card, Col, Form, Modal, Row, Spinner } from 'react-bootstrap';
import { useNavigate, useSearchParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { AniListMangaSearchParams } from '../../app/clients/backend/types';
import { ANI_LIST_SEARCH_TAGS } from './anilist-search-tags';

type MangaListItem = {
  id: number;
  title: string;
  description: string;
  coverUrl: string;
};

// Common manga genres/tags
const COMMON_GENRES = [
  { id: '391b0423-d847-456f-aff0-8b0cfc03066b', name: 'Action' },
  { id: 'f8f62932-27da-4fe4-8ee1-6779a8c5edba', name: 'Adventure' },
  { id: '4d32cc48-9f00-4cca-9b5a-a839f0764984', name: 'Comedy' },
  { id: 'b9af3a63-f058-46de-a9a0-e0c13906197a', name: 'Drama' },
  { id: 'cdc58593-87dd-415e-bbc0-2ec27bf404cc', name: 'Fantasy' },
  { id: 'ace04997-f6bd-436e-b261-779182193d3d', name: 'Historical' },
  { id: '87cc87cd-a395-47af-b27a-93258283bbc6', name: 'Horror' },
  { id: 'e5301a23-ebd9-49dd-a0cb-2add944c7fe9', name: 'Isekai' },
  { id: 'a3c67850-4684-404e-9b7f-c69850ee5da6', name: 'Magic' },
  { id: '799c202e-7daa-44eb-9cf7-8a3c0441531e', name: 'Martial Arts' },
  { id: 'ac72833b-c4e9-4878-b9db-6c8a4a99444a', name: 'Mecha' },
  { id: 'dd1f77c5-dea9-4e2b-97ae-224af09caf99', name: 'Medical' },
  { id: '92d6d951-ca5e-429c-ac78-451071cbf064', name: 'Music' },
  { id: 'df33b754-73a3-4c54-80e6-1a74a8f1f60a', name: 'Mystery' },
  { id: 'ee968100-4191-4968-93d3-f82d72be7e46', name: 'Philosophical' },
  { id: '69964a64-2f90-4d33-beeb-f3ed2875eb4c', name: 'Psychological' },
  { id: '423e2eae-a7a2-4a8b-ac03-a8351462d71d', name: 'Romance' },
  { id: '256c8bd9-4904-4360-bf4f-508a76d67183', name: 'Sci-Fi' },
  { id: 'e64993d2-0d03-4b02-a3f3-87b5c5f86e3f', name: 'Slice of Life' },
  { id: '69f59556-258d-4cc0-8e17-4454f3c9c4d9', name: 'Sports' },
  { id: '07251805-a27e-4d59-b488-f0bfbec15168', name: 'Superhero' },
  { id: '31932a7e-5b8e-49a6-9f12-2afa39dc544c', name: 'Supernatural' },
  { id: 'f42fbf9e-188a-447b-9fdc-f19dc1e4d685', name: 'Thriller' },
  { id: '9ab53f92-3eed-4e9b-903a-917c86035ee3', name: 'Tragedy' },
  { id: 'acc803a4-c95a-4c22-86fc-eb6b582d82a2', name: 'Wuxia' },
];

const ANI_LIST_CATEGORIES: Array<{ name: string; kind: 'genre' | 'tag' }> = COMMON_GENRES.map(({ name }) => ({
  name,
  kind: ANI_LIST_SEARCH_TAGS.has(name) ? 'tag' : 'genre',
}));
const FORMAT_OPTIONS = [['MANGA', 'Manga'], ['ONE_SHOT', 'One-shot'], ['NOVEL', 'Novel']];
const STATUS_OPTIONS = [['RELEASING', 'Releasing'], ['FINISHED', 'Finished'], ['HIATUS', 'Hiatus'], ['NOT_YET_RELEASED', 'Not yet released'], ['CANCELLED', 'Cancelled']];
const SORT_OPTIONS = [['POPULARITY_DESC', 'Most popular'], ['SCORE_DESC', 'Highest rated'], ['START_DATE_DESC', 'Newest release']];

const filtersFromParams = (params: URLSearchParams): AniListMangaSearchParams => {
  const year = Number(params.get('year') ?? '');
  return {
    title: params.get('title') ?? '',
    genre: params.get('genre') ?? '',
    tag: params.get('tag') ?? '',
    year: Number.isInteger(year) && year > 0 ? year : undefined,
    format: params.get('format') ?? '',
    status: params.get('status') ?? '',
    sort: params.get('sort') ?? '',
    page: 1,
    perPage: 20,
  };
};

const hasSearchCriteria = (filters: AniListMangaSearchParams) => Boolean(
  filters.title || filters.genre || filters.tag || filters.year || filters.format || filters.status || filters.sort,
);

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

const mapMangaItem = (item: unknown): MangaListItem | null => {
  if (!item || typeof item !== 'object') {
    return null;
  }

  const source = item as Record<string, unknown>;
  const id = typeof source.id === 'number' ? source.id : 0;
  if (id < 1) {
    return null;
  }

  // Backend already transforms the data, so title/description are at root level
  const title = readLocalizedValue((source.title as Record<string, unknown>)?.english) ||
                readLocalizedValue((source.title as Record<string, unknown>)?.userPreferred) ||
                readLocalizedValue((source.title as Record<string, unknown>)?.romaji) || 'Untitled manga';
  
  const description = typeof source.description === 'string' ? source.description :
                      readLocalizedValue(source.description) || 'No description available.';
  const coverImage = source.coverImage as Record<string, unknown> | undefined;
  const coverUrl = typeof coverImage?.large === 'string' ? coverImage.large : typeof coverImage?.medium === 'string' ? coverImage.medium : '';

  return {
    id,
    title,
    description,
    coverUrl,
  };
};

const MangaSearchPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const initialFilters = filtersFromParams(searchParams);
  const [title, setTitle] = useState(initialFilters.title ?? '');
  const [genre, setGenre] = useState(initialFilters.genre ?? '');
  const [tag, setTag] = useState(initialFilters.tag ?? '');
  const [year, setYear] = useState(initialFilters.year?.toString() ?? '');
  const [format, setFormat] = useState(initialFilters.format ?? '');
  const [status, setStatus] = useState(initialFilters.status ?? '');
  const [sort, setSort] = useState(initialFilters.sort ?? '');
  const [showFilters, setShowFilters] = useState(false);
  const [results, setResults] = useState<MangaListItem[]>([]);
  const [coverErrors, setCoverErrors] = useState<Set<number>>(new Set());
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [readingMangaID, setReadingMangaID] = useState<number | null>(null);
  const [error, setError] = useState<string>('');

  const runSearch = useCallback(
    async (filters: AniListMangaSearchParams) => {
      if (!hasSearchCriteria(filters)) {
        setResults([]);
        setError('Enter a title or choose at least one filter.');
        return;
      }

      setIsLoading(true);
      setError('');

      try {
        const response = await backend.searchAniListManga(filters);

        const mapped = response.data.map(mapMangaItem).filter((item): item is MangaListItem => item !== null);
        setResults(mapped);
        setCoverErrors(new Set());
      } catch (err) {
        setError((err as Error).message || 'Failed to search manga.');
      } finally {
        setIsLoading(false);
      }
    },
    [backend]
  );

  useEffect(() => {
    const filters = filtersFromParams(searchParams);
    setTitle(filters.title ?? '');
    setGenre(filters.genre ?? '');
    setTag(filters.tag ?? '');
    setYear(filters.year?.toString() ?? '');
    setFormat(filters.format ?? '');
    setStatus(filters.status ?? '');
    setSort(filters.sort ?? '');
    if (hasSearchCriteria(filters)) void runSearch(filters);
    else {
      setResults([]);
      setError('');
    }
  }, [runSearch, searchParams]);

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    const filters: AniListMangaSearchParams = {
      title: title.trim(), genre, tag, year: year ? Number(year) : undefined, format, status, sort, page: 1, perPage: 20,
    };
    if (!hasSearchCriteria(filters)) {
      setError('Enter a title or choose at least one filter.');
      setResults([]);
      return;
    }
    const nextParams: Record<string, string> = {};
    if (filters.title) nextParams.title = filters.title;
    if (filters.genre) nextParams.genre = filters.genre;
    if (filters.tag) nextParams.tag = filters.tag;
    if (filters.year) nextParams.year = String(filters.year);
    if (filters.format) nextParams.format = filters.format;
    if (filters.status) nextParams.status = filters.status;
    if (filters.sort) nextParams.sort = filters.sort;
    if (new URLSearchParams(nextParams).toString() === searchParams.toString()) void runSearch(filters);
    else setSearchParams(nextParams);
  };

  const clearFilters = () => {
    setGenre('');
    setTag('');
    setYear('');
    setFormat('');
    setStatus('');
    setSort('');
    const trimmedTitle = title.trim();
    setSearchParams(trimmedTitle ? { title: trimmedTitle } : {});
  };

  const startReading = async (event: React.MouseEvent<HTMLButtonElement>, manga: MangaListItem) => {
    event.stopPropagation();
    setReadingMangaID(manga.id);
    setError('');
    try {
      const source = await backend.getAniListReadingSource(manga.id);
      if (source.available && source.mangaId) {
        navigate(`/manga/${source.mangaId}`);
      } else {
        setError(`We do not have chapters for ${manga.title} yet.`);
      }
    } catch (err) {
      setError((err as Error).message || 'Unable to find a chapter source right now.');
    } finally {
      setReadingMangaID(null);
    }
  };

  return (
    <div className="p-3">
      <h2>Manga Search</h2>

      <Form className="mb-3" onSubmit={handleSubmit}>
        <div className="d-flex flex-wrap gap-2">
          <Form.Control id="manga-search-title" type="search" aria-label="Title" placeholder="Search AniList manga by title" value={title} onChange={(event) => setTitle(event.target.value)} style={{ maxWidth: '640px' }} />
          <Button type="button" variant="outline-secondary" onClick={() => setShowFilters(true)}>Filters</Button>
          <Button type="submit" variant="primary" disabled={isLoading}>Search</Button>
        </div>
      </Form>
      <Modal show={showFilters} onHide={() => setShowFilters(false)} centered>
        <Modal.Header closeButton><Modal.Title>Search filters</Modal.Title></Modal.Header>
        <Modal.Body>
          <Row className="g-3">
            <Col xs={12} sm={6}>
              <Form.Label htmlFor="manga-search-genre">Genre</Form.Label>
              <Form.Select id="manga-search-genre" value={genre || tag} onChange={(event) => {
                const selected = ANI_LIST_CATEGORIES.find((option) => option.name === event.target.value);
                setGenre(selected?.kind === 'genre' ? selected.name : '');
                setTag(selected?.kind === 'tag' ? selected.name : '');
              }}>
                <option value="">Any genre</option>
                {ANI_LIST_CATEGORIES.map((option) => <option key={option.name} value={option.name}>{option.name}</option>)}
              </Form.Select>
            </Col>
            <Col xs={12} sm={6}>
              <Form.Label htmlFor="manga-search-year">Release year</Form.Label>
              <Form.Control id="manga-search-year" type="number" min="1900" max="2100" placeholder="Any year" value={year} onChange={(event) => setYear(event.target.value)} />
            </Col>
            <Col xs={12} sm={6}>
              <Form.Label htmlFor="manga-search-format">Format</Form.Label>
              <Form.Select id="manga-search-format" value={format} onChange={(event) => setFormat(event.target.value)}>
                <option value="">Any format</option>
                {FORMAT_OPTIONS.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
              </Form.Select>
            </Col>
            <Col xs={12} sm={6}>
              <Form.Label htmlFor="manga-search-status">Publication status</Form.Label>
              <Form.Select id="manga-search-status" value={status} onChange={(event) => setStatus(event.target.value)}>
                <option value="">Any status</option>
                {STATUS_OPTIONS.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
              </Form.Select>
            </Col>
            <Col xs={12}>
              <Form.Label htmlFor="manga-search-sort">Sort by</Form.Label>
              <Form.Select id="manga-search-sort" value={sort} onChange={(event) => setSort(event.target.value)}>
                <option value="">Best match / popular</option>
                {SORT_OPTIONS.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
              </Form.Select>
            </Col>
          </Row>
        </Modal.Body>
        <Modal.Footer>
          <Button type="button" variant="outline-secondary" onClick={clearFilters}>Clear filters</Button>
          <Button type="button" variant="primary" onClick={() => setShowFilters(false)}>Done</Button>
        </Modal.Footer>
      </Modal>

      {isLoading && (
        <div className="mb-3">
          <Spinner animation="border" size="sm" className="me-2" />
          Loading manga...
        </div>
      )}

      {error && <div className="text-danger mb-3">{error}</div>}

      {!isLoading && !error && hasSearchCriteria(filtersFromParams(searchParams)) && results.length === 0 && (
        <p className="text-muted">No manga matched those filters.</p>
      )}

      <Row>
        {results.map((manga) => (
          <Col key={manga.id} xs={12} md={6} lg={4} className="mb-3">
            <Card className="flex-row h-100 overflow-hidden" style={{ cursor: 'pointer', minHeight: '180px' }} onClick={() => navigate(`/anilist/manga/${manga.id}`)}>
              {manga.coverUrl && !coverErrors.has(manga.id) ? (
                <Card.Img
                  src={manga.coverUrl}
                  alt={`Cover for ${manga.title}`}
                  referrerPolicy="no-referrer"
                  onError={() => setCoverErrors((previous) => new Set(previous).add(manga.id))}
                  style={{ width: '120px', height: '180px', objectFit: 'cover', flexShrink: 0 }}
                />
              ) : (
                <div
                  className="d-flex align-items-center justify-content-center text-muted bg-light text-center"
                  style={{ width: '120px', height: '180px', flexShrink: 0 }}
                >
                  No cover available
                </div>
              )}
              <Card.Body className="d-flex flex-column overflow-hidden">
                <Card.Title
                  style={{
                    display: '-webkit-box',
                    WebkitBoxOrient: 'vertical',
                    WebkitLineClamp: 2,
                    overflow: 'hidden',
                  }}
                >
                  {manga.title}
                </Card.Title>
                <Card.Text
                  className="mb-0"
                  style={{
                    display: '-webkit-box',
                    WebkitBoxOrient: 'vertical',
                    WebkitLineClamp: 5,
                    overflow: 'hidden',
                  }}
                >
                  {manga.description}
                </Card.Text>
                <Button
                  variant="outline-primary"
                  size="sm"
                  className="mt-auto align-self-start"
                  onClick={(event) => void startReading(event, manga)}
                  disabled={readingMangaID === manga.id}
                >
                  {readingMangaID === manga.id ? 'Finding chapters...' : 'Start reading'}
                </Button>
              </Card.Body>
            </Card>
          </Col>
        ))}
      </Row>

    </div>
  );
};

export default MangaSearchPage;

