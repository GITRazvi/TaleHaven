import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Badge, Button, Card, Col, Form, Row, Spinner } from 'react-bootstrap';
import { useNavigate, useSearchParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';

type MangaListItem = {
  id: string;
  title: string;
  description: string;
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
  const id = typeof source.id === 'string' ? source.id : '';
  if (!id) {
    return null;
  }

  // Backend already transforms the data, so title/description are at root level
  const title = typeof source.title === 'string' ? source.title :
                readLocalizedValue(source.title) || 'Untitled manga';
  
  const description = typeof source.description === 'string' ? source.description :
                      readLocalizedValue(source.description) || 'No description available.';

  return {
    id,
    title,
    description,
  };
};

const MangaSearchPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const initialTitle = searchParams.get('title') ?? '';
  const [title, setTitle] = useState<string>(initialTitle);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState<boolean>(false);
  const [results, setResults] = useState<MangaListItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');

  const runSearch = useCallback(
    async (searchTitle: string, tags: string[]) => {
      // At least title or tags must be provided
      if (!searchTitle.trim() && tags.length === 0) {
        setResults([]);
        setError('Please enter a manga title or select at least one genre.');
        return;
      }

      setIsLoading(true);
      setError('');

      try {
        const searchParams: {
          title?: string;
          includedTags?: string[];
          limit: number;
          offset: number;
        } = {
          limit: 20,
          offset: 0,
        };

        if (searchTitle.trim()) {
          searchParams.title = searchTitle.trim();
        }

        if (tags.length > 0) {
          searchParams.includedTags = tags;
        }

        const response = await backend.searchManga(searchParams as any);

        const mapped = response.data.map(mapMangaItem).filter((item): item is MangaListItem => item !== null);
        setResults(mapped);
      } catch (err) {
        setError((err as Error).message || 'Failed to search manga.');
      } finally {
        setIsLoading(false);
      }
    },
    [backend]
  );

  useEffect(() => {
    const titleFromUrl = searchParams.get('title') ?? '';
    if (titleFromUrl) {
      setTitle(titleFromUrl);
      runSearch(titleFromUrl, selectedTags);
    }
  }, [runSearch, searchParams, selectedTags]);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    const normalizedTitle = title.trim();

    if (!normalizedTitle && selectedTags.length === 0) {
      setError('Please enter a manga title or select at least one genre.');
      setResults([]);
      return;
    }

    if (normalizedTitle) {
      setSearchParams({ title: normalizedTitle });
    } else {
      setSearchParams({});
    }
    await runSearch(normalizedTitle, selectedTags);
  };

  const toggleTag = (tagId: string) => {
    setSelectedTags((prev) =>
      prev.includes(tagId)
        ? prev.filter((id) => id !== tagId)
        : [...prev, tagId]
    );
  };

  return (
    <div className="p-3">
      <h2>Manga Search</h2>

      <Form className="mb-3" onSubmit={handleSubmit}>
        <div className="d-flex gap-2 mb-2">
          <Form.Control
            type="search"
            placeholder="Search manga by title (optional)"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <Button type="submit" variant="primary" disabled={isLoading}>
            Search
          </Button>
          <Button
            variant="outline-secondary"
            onClick={() => setShowFilters(!showFilters)}
          >
            {showFilters ? 'Hide Filters' : 'Show Filters'}
          </Button>
        </div>

        {showFilters && (
          <div className="border rounded p-3 mb-2">
            <h6 className="mb-2">Filter by Genres:</h6>
            <div className="d-flex flex-wrap gap-2">
              {COMMON_GENRES.map((genre) => (
                <Badge
                  key={genre.id}
                  bg={selectedTags.includes(genre.id) ? 'primary' : 'secondary'}
                  style={{ cursor: 'pointer', fontSize: '0.9rem', padding: '0.5rem 0.75rem' }}
                  onClick={() => toggleTag(genre.id)}
                >
                  {genre.name}
                </Badge>
              ))}
            </div>
            {selectedTags.length > 0 && (
              <div className="mt-2">
                <small className="text-muted">
                  {selectedTags.length} genre{selectedTags.length !== 1 ? 's' : ''} selected
                </small>
                <Button
                  variant="link"
                  size="sm"
                  className="ms-2 p-0"
                  onClick={() => setSelectedTags([])}
                >
                  Clear all
                </Button>
              </div>
            )}
          </div>
        )}
      </Form>

      {isLoading && (
        <div className="mb-3">
          <Spinner animation="border" size="sm" className="me-2" />
          Loading manga...
        </div>
      )}

      {error && <div className="text-danger mb-3">{error}</div>}

      <Row>
        {results.map((manga) => (
          <Col key={manga.id} xs={12} md={6} lg={4} className="mb-3">
            <Card style={{ height: '100%', cursor: 'pointer' }} onClick={() => navigate(`/manga/${manga.id}`)}>
              <Card.Body>
                <Card.Title>{manga.title}</Card.Title>
                <Card.Text style={{ maxHeight: '120px', overflow: 'hidden' }}>{manga.description}</Card.Text>
              </Card.Body>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
};

export default MangaSearchPage;

