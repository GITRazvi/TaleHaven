import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Form, Spinner } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { AniListManga, AniListMangaSearchParams } from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';
import { parseChatbotRequest } from './chatbot-request';

type ChoiceGroup = 'start' | 'genre' | 'format' | 'status';

type RecommendationChoice = {
  label: string;
  answer: string;
  filters: Omit<AniListMangaSearchParams, 'page' | 'perPage'>;
};

const genres = ['Action', 'Adventure', 'Comedy', 'Fantasy', 'Horror', 'Romance', 'Sci-Fi'];
const surpriseGenres = ['Action', 'Adventure', 'Comedy', 'Drama', 'Fantasy', 'Mystery', 'Romance', 'Sci-Fi'];

const titleFor = (manga: AniListManga) => (
  manga.title.english || manga.title.userPreferred || manga.title.romaji || manga.title.native || 'Untitled manga'
);

const titleKey = (title: string) => title.toLowerCase().replace(/[^a-z0-9]+/g, ' ').trim();

const isAuthorRole = (role: string) => /author|story|creator|writer/i.test(role);

const GuidedRecommendationChat: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const { user } = useAuth();
  const [choiceGroup, setChoiceGroup] = useState<ChoiceGroup>('start');
  const [answer, setAnswer] = useState('');
  const [filters, setFilters] = useState<RecommendationChoice['filters'] | null>(null);
  const [recommendations, setRecommendations] = useState<AniListManga[]>([]);
  const [savedIDs, setSavedIDs] = useState<Set<number>>(new Set());
  const [libraryReady, setLibraryReady] = useState(!user);
  const [page, setPage] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [typedRequest, setTypedRequest] = useState('');
  const [excludedRecommendationIDs, setExcludedRecommendationIDs] = useState<Set<number>>(new Set());

  useEffect(() => {
    if (!user) {
      setSavedIDs(new Set());
      setLibraryReady(true);
      return;
    }
    let cancelled = false;
    setLibraryReady(false);
    void backend.getAniListMyLibrary({ limit: 100, offset: 0 })
      .then((response) => {
        if (!cancelled) {
          setSavedIDs(new Set(response.data.map((item) => item.anilistId)));
          setLibraryReady(true);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setSavedIDs(new Set());
          setLibraryReady(true);
        }
      });
    return () => { cancelled = true; };
  }, [backend, user]);

  const loadRecommendations = async (
    nextFilters: RecommendationChoice['filters'],
    nextPage: number,
    append: boolean,
    excludedIDs: Set<number> = excludedRecommendationIDs,
  ) => {
    setLoading(true);
    setError('');
    try {
      const response = await backend.searchAniListManga({ ...nextFilters, page: nextPage, perPage: 12, sort: 'SCORE_DESC' });
      const received = response.data.filter((manga) => !savedIDs.has(manga.id) && !excludedIDs.has(manga.id));
      setRecommendations((current) => append
        ? [...current, ...received.filter((manga) => !current.some((known) => known.id === manga.id))]
        : received);
      setPage(response.page || nextPage);
      setHasNextPage(response.hasNextPage);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Unable to find recommendations right now.');
    } finally {
      setLoading(false);
    }
  };

  const selectChoice = (choice: RecommendationChoice) => {
    setAnswer(choice.answer);
    setFilters(choice.filters);
    setRecommendations([]);
    setExcludedRecommendationIDs(new Set());
    void loadRecommendations(choice.filters, 1, false);
  };

  const startOver = () => {
    setChoiceGroup('start');
    setAnswer('');
    setFilters(null);
    setRecommendations([]);
    setPage(0);
    setHasNextPage(false);
    setError('');
    setTypedRequest('');
    setExcludedRecommendationIDs(new Set());
  };

  const resolveReferencedManga = async (title: string) => {
    const response = await backend.searchAniListManga({ title, page: 1, perPage: 5, sort: 'SEARCH_MATCH' });
    const exactMatch = response.data.find((manga) => titleKey(titleFor(manga)) === titleKey(title));
    const manga = exactMatch || response.data[0];
    if (!manga) throw new Error(`I couldn't find a manga named “${title}”.`);
    return manga;
  };

  const submitTypedRequest = async (event: React.FormEvent) => {
    event.preventDefault();
    const parsed = parseChatbotRequest(typedRequest);
    setError('');
    if (parsed.error) {
      setError(parsed.error);
      return;
    }

    setAnswer(typedRequest.trim());
    setRecommendations([]);
    setPage(0);
    setHasNextPage(false);
    setLoading(true);
    try {
      if (parsed.kind === 'filters' && parsed.filters) {
        setFilters(parsed.filters);
        const excluded = new Set<number>();
        setExcludedRecommendationIDs(excluded);
        await loadRecommendations(parsed.filters, 1, false, excluded);
        return;
      }

      const referencedManga = await resolveReferencedManga(parsed.title!);
      const details = await backend.getAniListMangaDetails(referencedManga.id);
      const excluded = new Set([referencedManga.id]);
      let nextFilters: RecommendationChoice['filters'];

      if (parsed.kind === 'same-author') {
        const author = details.data.staff?.edges.find((edge) => isAuthorRole(edge.role)) || details.data.staff?.edges[0];
        if (!author) throw new Error(`AniList does not list an author for “${titleFor(referencedManga)}”.`);
        nextFilters = { staffId: author.node.id };
      } else {
        const genres = (details.data.genres.length ? details.data.genres : referencedManga.genres).slice(0, 2);
        const tag = details.data.tags
          ?.filter((item) => !item.isMediaSpoiler)
          .sort((first, second) => second.rank - first.rank)[0]?.name;
        if (!genres.length && !tag) throw new Error(`AniList does not provide enough metadata to find titles similar to “${titleFor(referencedManga)}”.`);
        nextFilters = { ...(genres.length ? { genres } : {}), ...(tag ? { tag } : {}) };
      }

      setFilters(nextFilters);
      setExcludedRecommendationIDs(excluded);
      await loadRecommendations(nextFilters, 1, false, excluded);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Unable to prepare recommendations right now.');
    } finally {
      setLoading(false);
    }
  };

  const choices = (): RecommendationChoice[] => {
    if (choiceGroup === 'genre') return genres.map((genre) => ({ label: genre, answer: `${genre} sounds great. Here are highly rated picks.`, filters: { genre } }));
    if (choiceGroup === 'format') return [
      { label: 'One-shot', answer: 'Here are highly rated one-shots.', filters: { format: 'ONE_SHOT' } },
      { label: 'Manga series', answer: 'Here are highly rated manga series.', filters: { format: 'MANGA' } },
    ];
    if (choiceGroup === 'status') return [
      { label: 'Completed', answer: 'Here are completed manga you can finish at your own pace.', filters: { status: 'FINISHED' } },
      { label: 'Publishing now', answer: 'Here are highly rated manga that are still publishing.', filters: { status: 'RELEASING' } },
    ];
    return [];
  };

  const surprise = () => {
    const genre = surpriseGenres[Math.floor(Math.random() * surpriseGenres.length)];
    selectChoice({ label: 'Surprise me', answer: `Surprise! Let’s try something in ${genre}.`, filters: { genre } });
  };

  return <Card className="guided-recommendation-chat shadow-sm mb-5" aria-label="Manga recommendation assistant">
    <Card.Body>
      <div className="d-flex justify-content-between align-items-start gap-3 mb-3">
        <div><h2 className="h4 mb-1">Find your next manga</h2><p className="text-muted small mb-0">A quick guided recommendation chat powered by AniList.</p></div>
        {(answer || recommendations.length > 0) && <Button size="sm" variant="outline-secondary" onClick={startOver}>Start over</Button>}
      </div>

      <Form onSubmit={submitTypedRequest} className="mb-3">
        <Form.Label htmlFor="chatbot-request" className="small mb-1">Ask Chatbot</Form.Label>
        <div className="d-flex gap-2">
          <Form.Control
            id="chatbot-request"
            value={typedRequest}
            onChange={(event) => setTypedRequest(event.target.value)}
            placeholder="Action romance manga, or similar to Berserk"
            maxLength={200}
            aria-describedby="chatbot-request-help"
          />
          <Button type="submit" disabled={loading || (Boolean(user) && !libraryReady)}>Ask</Button>
        </div>
        <Form.Text id="chatbot-request-help" className="text-muted">Try “action romance completed manga”, “similar to Berserk”, or “same author as Berserk”.</Form.Text>
      </Form>

      <div className="recommendation-bubble recommendation-bubble-assistant">What would you like help finding?</div>
      {!answer && choiceGroup === 'start' && <div className="recommendation-choice-row" aria-label="Recommendation questions">
        <Button size="sm" variant="outline-primary" onClick={() => setChoiceGroup('genre')} disabled={Boolean(user) && !libraryReady}>A genre</Button>
        <Button size="sm" variant="outline-primary" onClick={() => setChoiceGroup('format')} disabled={Boolean(user) && !libraryReady}>Series type</Button>
        <Button size="sm" variant="outline-primary" onClick={() => setChoiceGroup('status')} disabled={Boolean(user) && !libraryReady}>Publication status</Button>
        <Button size="sm" variant="primary" onClick={surprise} disabled={Boolean(user) && !libraryReady}>Surprise me</Button>
      </div>}

      {!answer && choiceGroup !== 'start' && <>
        <div className="recommendation-bubble recommendation-bubble-assistant">{choiceGroup === 'genre' ? 'Which genre are you in the mood for?' : choiceGroup === 'format' ? 'Would you prefer a one-shot or a series?' : 'Would you like a completed manga or one still publishing?'}</div>
        <div className="recommendation-choice-row">{choices().map((choice) => <Button key={choice.label} size="sm" variant="outline-primary" onClick={() => selectChoice(choice)} disabled={Boolean(user) && !libraryReady}>{choice.label}</Button>)}<Button size="sm" variant="link" onClick={() => setChoiceGroup('start')}>Back</Button></div>
      </>}

      {answer && <div className="recommendation-bubble recommendation-bubble-reader">{answer}</div>}
      {loading && <p className="text-muted small mb-0"><Spinner animation="border" size="sm" className="me-2" />Finding recommendations…</p>}
      {error && <Alert variant="warning" className="mt-3 mb-0">{error}</Alert>}
      {!loading && answer && !error && recommendations.length === 0 && <p className="text-muted small mb-0">No new matches were found. Try another option.</p>}
      {recommendations.length > 0 && <div className="recommendation-results mt-3">{recommendations.map((manga) => {
        const title = titleFor(manga);
        const cover = manga.coverImage.medium || manga.coverImage.large;
        return <button key={manga.id} type="button" className="recommendation-result" onClick={() => navigate(`/anilist/manga/${manga.id}`)}>{cover ? <img src={cover} alt={`Cover for ${title}`} referrerPolicy="no-referrer" /> : <span className="recommendation-cover-fallback">No cover</span>}<span title={title}>{title}</span></button>;
      })}</div>}
      {filters && hasNextPage && <Button className="mt-3" size="sm" variant="outline-primary" onClick={() => void loadRecommendations(filters, page + 1, true)} disabled={loading}>{loading ? 'Loading…' : 'Show more'}</Button>}
      {user && !libraryReady && <p className="text-muted small mb-0 mt-3"><Spinner animation="border" size="sm" className="me-2" />Checking your library before recommending new titles…</p>}
      {user && answer && <p className="text-muted small mb-0 mt-3">Titles already in your library are left out when possible.</p>}
    </Card.Body>
  </Card>;
};

export default GuidedRecommendationChat;
