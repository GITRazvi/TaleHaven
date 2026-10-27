import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Col,
  Container,
  Form,
  Modal,
  Row,
  Spinner,
} from "react-bootstrap";
import { Link, useNavigate, useParams } from "react-router-dom";
import Backend from "../../app/clients/backend/backend";
import {
  AniListCharacterDetails,
  AniListLibraryItem,
  AniListManga,
  AniListReadingStatus,
} from "../../app/clients/backend/types";
import { useAuth } from "../../contexts/AuthContext";
import { ANI_LIST_SEARCH_TAGS } from "./anilist-search-tags";

const statusOptions: Array<{ status: AniListReadingStatus; title: string }> = [
  { status: "CURRENTLY_READING", title: "Reading" },
  { status: "READ", title: "Completed" },
  { status: "ON_HOLD", title: "On Hold" },
  { status: "DROPPED", title: "Dropped" },
  { status: "PLAN_TO_READ", title: "Plan to Read" },
];

const titleOf = (manga: AniListManga) =>
  manga.title.english ||
  manga.title.userPreferred ||
  manga.title.romaji ||
  manga.title.native ||
  "Untitled manga";
const alternativeTitlesOf = (manga: AniListManga) => {
  const primaryTitle = titleOf(manga).trim().toLocaleLowerCase();
  const seen = new Set<string>();
  return [
    manga.title.english,
    manga.title.romaji,
    manga.title.native,
    manga.title.userPreferred,
    ...(manga.synonyms || []),
  ].filter((title) => {
    const normalizedTitle = title.trim().toLocaleLowerCase();
    if (!normalizedTitle || normalizedTitle === primaryTitle || seen.has(normalizedTitle)) {
      return false;
    }
    seen.add(normalizedTitle);
    return true;
  });
};
const characterNameOf = (
  character: NonNullable<AniListManga["characters"]>["edges"][number]["node"],
) =>
  character.name.full ||
  character.name.userPreferred ||
  character.name.native ||
  "Unnamed character";
const characterRoleLabel = (role: string) =>
  role
    ? role.charAt(0) + role.slice(1).toLowerCase().replace(/_/g, " ")
    : "Character";
const characterRoleRank = (role: string) => {
  switch (role.toUpperCase()) {
    case "MAIN":
      return 0;
    case "SUPPORTING":
      return 1;
    default:
      return 2;
  }
};
const characterDateLabel = (date: AniListCharacterDetails["dateOfBirth"]) =>
  [date.year, date.month, date.day].filter(Boolean).join("-");
const mangaReleaseDateLabel = (date?: AniListManga["startDate"]) => {
  if (!date?.year) return "Unknown";
  const month = date.month ? String(date.month).padStart(2, "0") : "??";
  const day = date.day ? String(date.day).padStart(2, "0") : "??";
  return `${date.year}-${month}-${day}`;
};
const labelOf = (value: string) =>
  value
    .toLocaleLowerCase()
    .replace(/_/g, " ")
    .replace(/\b\w/g, (letter) => letter.toLocaleUpperCase());
const publicationStatusLabel = (status: string) => {
  if (status === "RELEASING") return "Publishing";
  return status ? labelOf(status) : "Unknown";
};
const staffNameOf = (
  staff: NonNullable<AniListManga["staff"]>["edges"][number]["node"],
) => staff.name.full || staff.name.userPreferred || staff.name.native || "Unknown";
const rankingOf = (rankings: AniListManga["rankings"] | undefined, type: string) =>
  (rankings || []).find((ranking) => ranking.type === type && ranking.allTime) ||
  (rankings || []).find((ranking) => ranking.type === type);
const scoreLabel = (score: number) =>
  (score / 10).toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
const mediaTitleOf = (
  media: AniListCharacterDetails["media"]["edges"][number]["node"],
) =>
  media.title.english ||
  media.title.userPreferred ||
  media.title.romaji ||
  media.title.native ||
  "Untitled work";
const relationLabel = (relationType: string) =>
  relationType
    ? relationType.charAt(0) + relationType.slice(1).toLowerCase().replace(/_/g, " ")
    : "Related";
type MangaCharacterEdge = NonNullable<AniListManga["characters"]>["edges"][number];
const externalReadingLinksOf = (manga: AniListManga) => {
  const seen = new Set<string>();
  return (manga.externalLinks || []).filter((link) => {
    if (link.type !== "STREAMING" || !/^https?:\/\//i.test(link.url) || seen.has(link.url)) {
      return false;
    }
    seen.add(link.url);
    return true;
  });
};
const RECOMMENDATION_COUNT = 6;

const AniListMangaDetailsPage: React.FC = () => {
  const { anilistId } = useParams();
  const navigate = useNavigate();
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();
  const [manga, setManga] = useState<AniListManga | null>(null);
  const [genreRecommendations, setGenreRecommendations] = useState<AniListManga[]>([]);
  const [genreRecommendationsLoading, setGenreRecommendationsLoading] = useState(false);
  const [characterEdges, setCharacterEdges] = useState<MangaCharacterEdge[]>([]);
  const [characterPageInfo, setCharacterPageInfo] = useState<NonNullable<AniListManga["characters"]>["pageInfo"] | null>(null);
  const [charactersLoadingMore, setCharactersLoadingMore] = useState(false);
  const [charactersLoadError, setCharactersLoadError] = useState("");
  const [savedItem, setSavedItem] = useState<AniListLibraryItem | null>(null);
  const [error, setError] = useState("");
  const [settingsError, setSettingsError] = useState("");
  const [libraryActionSaving, setLibraryActionSaving] = useState(false);
  const [favoriteMangaActionSaving, setFavoriteMangaActionSaving] =
    useState(false);
  const [favoriteAniListIDs, setFavoriteAniListIDs] = useState<Set<number>>(
    new Set(),
  );
  const [statusSaving, setStatusSaving] = useState(false);
  const [ratingSaving, setRatingSaving] = useState(false);
  const [favoriteCharacterIDs, setFavoriteCharacterIDs] = useState<Set<number>>(
    new Set(),
  );
  const [characterFavoriteActionID, setCharacterFavoriteActionID] = useState<
    number | null
  >(null);
  const [selectedCharacterID, setSelectedCharacterID] = useState<number | null>(
    null,
  );
  const [characterDetails, setCharacterDetails] =
    useState<AniListCharacterDetails | null>(null);
  const [characterDetailsError, setCharacterDetailsError] = useState("");
  const [characterDetailsLoading, setCharacterDetailsLoading] = useState(false);
  const [chapterSourceID, setChapterSourceID] = useState<string | null>(null);
  const [chapterAvailabilityLoading, setChapterAvailabilityLoading] = useState(true);
  const [chapterAvailabilityError, setChapterAvailabilityError] = useState("");

  useEffect(() => {
    const id = Number(anilistId);
    if (!Number.isInteger(id) || id < 1) {
      setError("Invalid manga ID.");
      return;
    }
    setCharacterEdges([]);
    setCharacterPageInfo(null);
    setCharactersLoadError("");
    void backend
      .getAniListMangaDetails(id)
      .then((response) => {
        setManga(response.data);
        setCharacterEdges(response.data.characters?.edges || []);
        setCharacterPageInfo(response.data.characters?.pageInfo || null);
      })
      .catch((err) =>
        setError((err as Error).message || "Unable to load manga details."),
      );
  }, [anilistId, backend]);

  useEffect(() => {
    const genre = manga?.genres[0];
    if (!manga || !genre) {
      setGenreRecommendations([]);
      return;
    }
    let isCurrent = true;
    setGenreRecommendations([]);
    setGenreRecommendationsLoading(true);
    void backend
      .getAniListMangaByGenre(genre, 1, RECOMMENDATION_COUNT + 1)
      .then((response) => {
        if (isCurrent) {
          setGenreRecommendations(
            response.data
              .filter((item) => item.id !== manga.id)
              .slice(0, RECOMMENDATION_COUNT),
          );
        }
      })
      .catch(() => {
        if (isCurrent) setGenreRecommendations([]);
      })
      .finally(() => {
        if (isCurrent) setGenreRecommendationsLoading(false);
      });
    return () => { isCurrent = false; };
  }, [backend, manga]);

  const loadMoreCharacters = async () => {
    if (!manga || !characterPageInfo?.hasNextPage || charactersLoadingMore) return;
    setCharactersLoadingMore(true);
    setCharactersLoadError("");
    try {
      const response = await backend.getAniListMangaCharacters(
        manga.id,
        characterPageInfo.currentPage + 1,
      );
      setCharacterEdges((current) => {
        const seen = new Set(current.map((character) => character.node.id));
        return [...current, ...response.data.edges.filter((character) => !seen.has(character.node.id))];
      });
      setCharacterPageInfo(response.data.pageInfo);
    } catch (requestError) {
      setCharactersLoadError((requestError as Error).message || "Unable to load more characters.");
    } finally {
      setCharactersLoadingMore(false);
    }
  };

  useEffect(() => {
    const id = Number(anilistId);
    if (!Number.isInteger(id) || id < 1) return;
    let isCurrent = true;
    setChapterSourceID(null);
    setChapterAvailabilityError("");
    setChapterAvailabilityLoading(true);
    void backend
      .getAniListReadingSource(id)
      .then((response) => {
        if (isCurrent) setChapterSourceID(response.available ? response.mangaId || null : null);
      })
      .catch((err) => {
        if (isCurrent) setChapterAvailabilityError((err as Error).message || "Unable to check chapter availability.");
      })
      .finally(() => {
        if (isCurrent) setChapterAvailabilityLoading(false);
      });
    return () => { isCurrent = false; };
  }, [anilistId, backend]);

  useEffect(() => {
    const id = Number(anilistId);
    if (!user || !Number.isInteger(id) || id < 1) {
      setSavedItem(null);
      return;
    }
    let isCurrent = true;
    void backend
      .getAniListMyLibrary({ limit: 100, offset: 0 })
      .then((response) => {
        if (isCurrent)
          setSavedItem(
            response.data.find((item) => item.anilistId === id) || null,
          );
      })
      .catch(() => {
        if (isCurrent) setSavedItem(null);
      });
    return () => {
      isCurrent = false;
    };
  }, [anilistId, backend, user]);

  useEffect(() => {
    if (!user) {
      setFavoriteAniListIDs(new Set());
      return;
    }
    let isCurrent = true;
    void backend
      .getAniListFavoriteIDs()
      .then((response) => {
        if (isCurrent) setFavoriteAniListIDs(new Set(response.data));
      })
      .catch(() => {
        if (isCurrent) setFavoriteAniListIDs(new Set());
      });
    return () => {
      isCurrent = false;
    };
  }, [backend, user]);

  useEffect(() => {
    if (!user) {
      setFavoriteCharacterIDs(new Set());
      return;
    }
    let isCurrent = true;
    void backend
      .getAniListCharacterFavorites()
      .then((response) => {
        if (isCurrent)
          setFavoriteCharacterIDs(
            new Set(response.data.map((character) => character.characterId)),
          );
      })
      .catch(() => {
        if (isCurrent) setFavoriteCharacterIDs(new Set());
      });
    return () => {
      isCurrent = false;
    };
  }, [backend, user]);

  useEffect(() => {
    if (!selectedCharacterID) {
      setCharacterDetails(null);
      setCharacterDetailsError("");
      return;
    }
    let isCurrent = true;
    setCharacterDetails(null);
    setCharacterDetailsError("");
    setCharacterDetailsLoading(true);
    void backend
      .getAniListCharacterDetails(selectedCharacterID)
      .then((response) => {
        if (isCurrent) setCharacterDetails(response.data);
      })
      .catch((err) => {
        if (isCurrent)
          setCharacterDetailsError(
            (err as Error).message || "Unable to load character details.",
          );
      })
      .finally(() => {
        if (isCurrent) setCharacterDetailsLoading(false);
      });
    return () => {
      isCurrent = false;
    };
  }, [backend, selectedCharacterID]);

  const updateStatus = async (status: AniListReadingStatus) => {
    if (!savedItem || savedItem.readingStatus === status) return;
    setStatusSaving(true);
    setSettingsError("");
    try {
      await backend.updateAniListMangaReadingStatus(
        savedItem.anilistId,
        status,
      );
      setSavedItem((current) =>
        current ? { ...current, readingStatus: status } : current,
      );
    } catch (err) {
      setSettingsError(
        (err as Error).message || "Unable to update your reading status.",
      );
    } finally {
      setStatusSaving(false);
    }
  };

  const toggleLibrary = async () => {
    if (!manga) return;
    if (!user) {
      navigate("/login");
      return;
    }
    setLibraryActionSaving(true);
    setSettingsError("");
    try {
      if (savedItem) {
        await backend.removeAniListMangaFromMyLibrary(manga.id);
        setSavedItem(null);
        setFavoriteAniListIDs((current) => {
          const next = new Set(current);
          next.delete(manga.id);
          return next;
        });
      } else {
        await backend.addAniListMangaToMyLibrary(manga.id);
        setSavedItem({
          anilistId: manga.id,
          addedAt: "",
          readingStatus: "PLAN_TO_READ",
          personalRating: null,
          manga,
        });
      }
    } catch (err) {
      setSettingsError(
        (err as Error).message || "Unable to update your library.",
      );
    } finally {
      setLibraryActionSaving(false);
    }
  };

  const toggleMangaFavorite = async () => {
    if (!manga || !savedItem) return;
    const isFavorite = favoriteAniListIDs.has(manga.id);
    setFavoriteMangaActionSaving(true);
    setSettingsError("");
    try {
      if (isFavorite) await backend.removeAniListMangaFromFavorites(manga.id);
      else await backend.addAniListMangaToFavorites(manga.id);
      setFavoriteAniListIDs((current) => {
        const next = new Set(current);
        if (isFavorite) next.delete(manga.id);
        else next.add(manga.id);
        return next;
      });
    } catch (err) {
      setSettingsError(
        (err as Error).message || "Unable to update favorites.",
      );
    } finally {
      setFavoriteMangaActionSaving(false);
    }
  };

  const updateRating = async (personalRating: number | null) => {
    if (!savedItem || savedItem.personalRating === personalRating) return;
    setRatingSaving(true);
    setSettingsError("");
    try {
      await backend.updateAniListMangaPersonalRating(
        savedItem.anilistId,
        personalRating,
      );
      setSavedItem((current) =>
        current ? { ...current, personalRating } : current,
      );
    } catch (err) {
      setSettingsError(
        (err as Error).message || "Unable to update your personal rating.",
      );
    } finally {
      setRatingSaving(false);
    }
  };

  const toggleCharacterFavorite = async (
    event: React.MouseEvent<HTMLButtonElement>,
    character: NonNullable<AniListManga["characters"]>["edges"][number],
  ) => {
    event.stopPropagation();
    if (!manga) return;
    const favorite = favoriteCharacterIDs.has(character.node.id);
    setCharacterFavoriteActionID(character.node.id);
    try {
      if (favorite)
        await backend.removeAniListCharacterFromFavorites(character.node.id);
      else
        await backend.addAniListCharacterToFavorites(
          character.node.id,
          manga.id,
          character.role,
        );
      setFavoriteCharacterIDs((current) => {
        const next = new Set(current);
        if (favorite) next.delete(character.node.id);
        else next.add(character.node.id);
        return next;
      });
    } finally {
      setCharacterFavoriteActionID(null);
    }
  };

  if (error)
    return (
      <Container className="py-5">
        <Alert variant="danger">{error}</Alert>
        <Button onClick={() => navigate("/")}>Back to discovery</Button>
      </Container>
    );
  if (!manga)
    return (
      <Container className="py-5 text-center">
        <Spinner animation="border" />{" "}
        <span className="ms-2">Loading manga details...</span>
      </Container>
    );
  const title = titleOf(manga);
  const alternativeTitles = alternativeTitlesOf(manga);
  const releaseDate = mangaReleaseDateLabel(manga.startDate);
  const cover = manga.coverImage.large || manga.coverImage.medium;
  const relatedManga = (manga.relations?.edges || []).filter(
    (relation) => relation.node.type === "MANGA",
  );
  const externalReadingLinks = externalReadingLinksOf(manga);
  const authors = (manga.staff?.edges || []).filter((staff) => staffNameOf(staff.node) !== "Unknown");
  const ratedRanking = rankingOf(manga.rankings, "RATED");
  const popularityRanking = rankingOf(manga.rankings, "POPULAR");
  const hasStatistics = Boolean(
    manga.averageScore > 0 ||
      ratedRanking?.rank ||
      popularityRanking?.rank ||
      manga.popularity > 0 ||
      manga.favourites,
  );
  const shownGenreNames = new Set<string>();
  const genresAndSearchTags = [
    ...manga.genres.map((name) => ({
      name,
      to: `/anilist/genre/${encodeURIComponent(name)}`,
    })),
    ...(manga.tags || [])
      .filter(
        (tag) =>
          tag.name &&
          !tag.isMediaSpoiler &&
          ANI_LIST_SEARCH_TAGS.has(tag.name),
      )
      .sort((left, right) => right.rank - left.rank)
      .map((tag) => ({
        name: tag.name,
        to: `/manga/search?tag=${encodeURIComponent(tag.name)}`,
      })),
  ].filter(({ name }) => {
    const normalizedName = name.toLocaleLowerCase();
    if (shownGenreNames.has(normalizedName)) return false;
    shownGenreNames.add(normalizedName);
    return true;
  });
  const characters = characterEdges
    .map((character, index) => ({ character, index }))
    .sort(
      (left, right) =>
        characterRoleRank(left.character.role) -
          characterRoleRank(right.character.role) || left.index - right.index,
    )
    .map(({ character }) => character);
  const isFavorite = favoriteAniListIDs.has(manga.id);
  return (
    <Container className="py-5">
      <Button
        variant="outline-secondary"
        className="mb-4"
        onClick={() => navigate(-1)}
      >
        Back
      </Button>
      <Row>
        <Col md={4} lg={3}>
          {cover ? (
            <img
              src={cover}
              alt={title}
              referrerPolicy="no-referrer"
              className="img-fluid rounded shadow-sm"
            />
          ) : (
            <Card className="p-5 text-center">No cover</Card>
          )}
          <section className="mt-3" aria-label="Manga information">
            <h1 className="h3 mb-3">{title}</h1>
            {alternativeTitles.length > 0 && (
              <p className="text-muted mb-3">
                <strong>Alternative titles:</strong> {alternativeTitles.join(" · ")}
              </p>
            )}
            <p className="mb-3">
              <strong>Release date:</strong> {releaseDate}
            </p>
            <dl className="mb-3">
              <dt>Type</dt>
              <dd>{manga.format ? labelOf(manga.format) : "Unknown"}</dd>
              <dt>Volumes</dt>
              <dd>{manga.volumes > 0 ? manga.volumes : "Unknown"}</dd>
              <dt>Chapters</dt>
              <dd>{manga.chapters > 0 ? manga.chapters : "Unknown"}</dd>
              <dt>Status</dt>
              <dd>{publicationStatusLabel(manga.status)}</dd>
              {authors.length > 0 && (
                <>
                  <dt>Authors</dt>
                  <dd>
                    {authors.map((staff) => (
                      <div key={staff.node.id}>
                        {staffNameOf(staff.node)}{staff.role ? ` (${staff.role})` : ""}
                      </div>
                    ))}
                  </dd>
                </>
              )}
            </dl>
            {genresAndSearchTags.length > 0 && (
              <div className="mb-3" aria-label="Genres">
                <strong className="d-block mb-2">Genres</strong>
                {genresAndSearchTags.map(({ name, to }) => (
                  <Link
                    key={name}
                    to={to}
                    className="badge bg-secondary text-decoration-none me-2 mb-2"
                  >
                    {name}
                  </Link>
                ))}
              </div>
            )}
            {hasStatistics && (
              <section className="mb-3" aria-labelledby="manga-statistics-heading">
                <h2 id="manga-statistics-heading" className="h6 mb-2">Statistics</h2>
                <dl className="mb-0">
                  {manga.averageScore > 0 && <><dt>Score</dt><dd>{scoreLabel(manga.averageScore)}/10</dd></>}
                  {ratedRanking?.rank && <><dt>Ranked</dt><dd>#{ratedRanking.rank.toLocaleString()}</dd></>}
                  {popularityRanking?.rank && <><dt>Popularity</dt><dd>#{popularityRanking.rank.toLocaleString()}</dd></>}
                  {manga.popularity > 0 && <><dt>Members</dt><dd>{manga.popularity.toLocaleString()}</dd></>}
                  {manga.favourites ? <><dt>Favorites</dt><dd>{manga.favourites.toLocaleString()}</dd></> : null}
                </dl>
              </section>
            )}
          </section>
        </Col>
        <Col md={8} lg={9}>
          <div className="d-flex flex-wrap align-items-center gap-2 mb-2">
            <div className="me-auto" />
            <Button
              size="sm"
              variant={savedItem ? "outline-danger" : "outline-primary"}
              onClick={() => void toggleLibrary()}
              disabled={libraryActionSaving}
              aria-label={
                savedItem
                  ? `Remove ${title} from My Library`
                  : `Add ${title} to My Library`
              }
            >
              {libraryActionSaving ? "…" : savedItem ? "− Library" : "+ Library"}
            </Button>
            {savedItem && (
              <Button
                size="sm"
                variant={isFavorite ? "danger" : "outline-danger"}
                onClick={() => void toggleMangaFavorite()}
                disabled={favoriteMangaActionSaving}
                aria-label={
                  isFavorite
                    ? `Remove ${title} from Favorites`
                    : `Add ${title} to Favorites`
                }
              >
                {favoriteMangaActionSaving ? "…" : isFavorite ? "♥" : "♡"}
              </Button>
            )}
          </div>
          <p className="text-muted">
            {manga.format} {manga.status ? `• ${manga.status}` : ""}
          </p>
          {manga.averageScore > 0 && <p>★ {manga.averageScore}%</p>}
          {settingsError && (
            <Alert variant="danger" className="py-2">
              {settingsError}
            </Alert>
          )}
          {(chapterAvailabilityLoading || chapterSourceID || chapterAvailabilityError) && (
            <Card className="mb-4" aria-live="polite">
              <Card.Body className="py-3 d-flex flex-wrap align-items-center gap-2">
                {chapterAvailabilityLoading && <><Spinner animation="border" size="sm" /> Checking chapter availability...</>}
                {chapterSourceID && <><span className="text-success">Chapters available on MangaDex</span><Button size="sm" onClick={() => navigate(`/manga/${chapterSourceID}`)}>Start reading</Button></>}
                {chapterAvailabilityError && <small className="text-danger">{chapterAvailabilityError}</small>}
              </Card.Body>
            </Card>
          )}
          {externalReadingLinks.length > 0 && (
            <section className="mb-4" aria-labelledby="read-on-heading">
              <h2 id="read-on-heading" className="h5 mb-2">Read on</h2>
              <div className="d-flex flex-wrap gap-2">
                {externalReadingLinks.map((link) => (
                  <a
                    key={link.url}
                    href={link.url}
                    target="_blank"
                    rel="noreferrer"
                    className="btn btn-outline-primary btn-sm"
                    aria-label={`Read ${title} on ${link.site}${link.language ? ` (${link.language})` : ""}`}
                  >
                    {link.site}{link.language ? ` · ${link.language}` : ""}
                  </a>
                ))}
              </div>
            </section>
          )}
          {savedItem && (
            <Card className="anilist-library-metadata-panel mb-4">
              <Card.Body className="p-3">
                <h2 className="h5 mb-3">Your library settings</h2>
                <Row className="g-3">
                  <Col sm={6}>
                    <Form.Label htmlFor="anilist-reading-status">
                      Reading status
                    </Form.Label>
                    <Form.Select
                      id="anilist-reading-status"
                      aria-label="Reading status"
                      value={savedItem.readingStatus || "PLAN_TO_READ"}
                      disabled={statusSaving}
                      onChange={(event) =>
                        void updateStatus(
                          event.target.value as AniListReadingStatus,
                        )
                      }
                    >
                      {statusOptions.map((option) => (
                        <option key={option.status} value={option.status}>
                          {option.title}
                        </option>
                      ))}
                    </Form.Select>
                  </Col>
                  <Col sm={6}>
                    <Form.Label htmlFor="anilist-personal-rating">
                      Your rating
                    </Form.Label>
                    <Form.Select
                      id="anilist-personal-rating"
                      aria-label="Your rating"
                      value={savedItem.personalRating ?? ""}
                      disabled={ratingSaving}
                      onChange={(event) =>
                        void updateRating(
                          event.target.value === ""
                            ? null
                            : Number(event.target.value),
                        )
                      }
                    >
                      <option value="">Not rated</option>
                      {Array.from({ length: 10 }, (_, index) => (
                        <option key={index + 1} value={index + 1}>
                          {index + 1}/10
                        </option>
                      ))}
                    </Form.Select>
                  </Col>
                </Row>
              </Card.Body>
            </Card>
          )}
          <p style={{ whiteSpace: "pre-line" }}>
            {manga.description || "No description is available."}
          </p>
          {relatedManga.length > 0 && (
            <section className="mt-4" aria-labelledby="related-manga-heading">
              <h2 id="related-manga-heading" className="h3 mb-3">Related manga</h2>
              <Row className="g-3">
                {relatedManga.map((relation) => {
                  const relatedTitle =
                    relation.node.title.english ||
                    relation.node.title.userPreferred ||
                    relation.node.title.romaji ||
                    relation.node.title.native ||
                    "Untitled manga";
                  const relatedCover =
                    relation.node.coverImage.large || relation.node.coverImage.medium;
                  return (
                    <Col key={`${relation.relationType}-${relation.node.id}`} md={6} lg={4}>
                      <Link
                        to={`/anilist/manga/${relation.node.id}`}
                        className="text-decoration-none"
                        aria-label={`Open ${relatedTitle}`}
                      >
                        <Card className="h-100">
                          {relatedCover && (
                            <Card.Img
                              variant="top"
                              src={relatedCover}
                              alt={relatedTitle}
                              referrerPolicy="no-referrer"
                            />
                          )}
                          <Card.Body>
                            <Card.Title className="h6">{relatedTitle}</Card.Title>
                            <Card.Text className="text-muted mb-0">
                              {relationLabel(relation.relationType)}
                            </Card.Text>
                          </Card.Body>
                        </Card>
                      </Link>
                    </Col>
                  );
                })}
              </Row>
            </section>
          )}
          {characters.length > 0 && (
            <section
              className="anilist-characters-section mt-4"
              aria-labelledby="characters-heading"
            >
              <h2 id="characters-heading" className="h3 mb-3">
                Characters
              </h2>
              <Row className="g-3">
                {characters.map((character) => {
                  const name = characterNameOf(character.node);
                  const image =
                    character.node.image.large || character.node.image.medium;
                  const isFavorite = favoriteCharacterIDs.has(
                    character.node.id,
                  );
                  return (
                    <Col key={character.node.id} md={6}>
                      <div className="anilist-character-item">
                        <button
                          type="button"
                          className="anilist-character-details-trigger"
                          onClick={() =>
                            setSelectedCharacterID(character.node.id)
                          }
                          aria-label={`View ${name} details`}
                        >
                          {image ? (
                            <img
                              src={image}
                              alt={name}
                              referrerPolicy="no-referrer"
                            />
                          ) : (
                            <div
                              className="anilist-character-placeholder"
                              aria-hidden="true"
                            >
                              ?
                            </div>
                          )}
                          <div>
                            <h3 className="h6 mb-1">{name}</h3>
                            <p className="mb-0">
                              {characterRoleLabel(character.role)}
                            </p>
                          </div>
                        </button>
                        {user && (
                          <button
                            type="button"
                            className={`anilist-character-favorite ${isFavorite ? "is-saved" : ""}`}
                            onClick={(event) =>
                              void toggleCharacterFavorite(event, character)
                            }
                            disabled={
                              characterFavoriteActionID === character.node.id
                            }
                            aria-label={
                              isFavorite
                                ? `Remove ${name} from favorite characters`
                                : `Add ${name} to favorite characters`
                            }
                          >
                            {characterFavoriteActionID === character.node.id
                              ? "…"
                              : isFavorite
                                ? "♥"
                                : "♡"}
                          </button>
                        )}
                      </div>
                    </Col>
                  );
                })}
              </Row>
              {charactersLoadError && <Alert variant="danger" className="mt-3 mb-0">{charactersLoadError}</Alert>}
              {characterPageInfo?.hasNextPage && (
                <Button
                  variant="outline-primary"
                  className="mt-3"
                  onClick={() => void loadMoreCharacters()}
                  disabled={charactersLoadingMore}
                >
                  {charactersLoadingMore ? "Loading characters..." : "Load more characters"}
                </Button>
              )}
            </section>
          )}
          {(genreRecommendationsLoading || genreRecommendations.length > 0) && (
            <section className="mt-4" aria-labelledby="genre-recommendations-heading">
              <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 id="genre-recommendations-heading" className="h3 mb-0">
                  Top-rated in {manga.genres[0]}
                </h2>
                <Link to={`/anilist/genre/${encodeURIComponent(manga.genres[0])}`} className="small">
                  View all
                </Link>
              </div>
              {genreRecommendationsLoading ? (
                <div className="text-center py-3"><Spinner animation="border" size="sm" className="me-2" />Loading recommendations...</div>
              ) : (
                <Row className="g-3">
                  {genreRecommendations.map((recommendation) => {
                    const recommendationTitle = titleOf(recommendation);
                    const recommendationCover = recommendation.coverImage.large || recommendation.coverImage.medium;
                    return (
                      <Col key={recommendation.id} xs={6} sm={4} md={3} lg={2}>
                        <Link
                          to={`/anilist/manga/${recommendation.id}`}
                          className="text-decoration-none"
                          aria-label={`Open ${recommendationTitle}`}
                        >
                          <Card className="h-100 manga-library-card">
                            {recommendationCover ? (
                              <Card.Img variant="top" src={recommendationCover} alt={recommendationTitle} referrerPolicy="no-referrer" />
                            ) : (
                              <div className="d-flex align-items-center justify-content-center text-muted" style={{ minHeight: "180px", backgroundColor: recommendation.coverImage.color || "#f0f0f0" }}>No cover</div>
                            )}
                            <Card.Body className="p-2">
                              <Card.Title className="small mb-1" title={recommendationTitle}>{recommendationTitle}</Card.Title>
                              {recommendation.averageScore > 0 && <Card.Text className="small text-muted mb-0">★ {recommendation.averageScore}%</Card.Text>}
                            </Card.Body>
                          </Card>
                        </Link>
                      </Col>
                    );
                  })}
                </Row>
              )}
            </section>
          )}
        </Col>
      </Row>
      <Modal
        show={selectedCharacterID !== null}
        onHide={() => setSelectedCharacterID(null)}
        size="lg"
        centered
      >
        <Modal.Header closeButton>
          <Modal.Title>
            {characterDetails
              ? characterNameOf(characterDetails)
              : "Character details"}
          </Modal.Title>
        </Modal.Header>
        <Modal.Body>
          {characterDetailsLoading && (
            <div className="text-center py-4">
              <Spinner animation="border" size="sm" className="me-2" />
              Loading character details...
            </div>
          )}
          {characterDetailsError && (
            <Alert variant="danger" className="mb-0">
              {characterDetailsError}
            </Alert>
          )}
          {characterDetails && (
            <>
              <div className="character-details-summary">
                {(characterDetails.image.large ||
                  characterDetails.image.medium) && (
                  <img
                    src={
                      characterDetails.image.large ||
                      characterDetails.image.medium
                    }
                    alt={characterNameOf(characterDetails)}
                    referrerPolicy="no-referrer"
                  />
                )}
                <div>
                  <dl className="character-details-facts mb-0">
                    {characterDetails.gender && (
                      <>
                        <dt>Gender</dt>
                        <dd>{characterDetails.gender}</dd>
                      </>
                    )}
                    {characterDateLabel(characterDetails.dateOfBirth) && (
                      <>
                        <dt>Born</dt>
                        <dd>
                          {characterDateLabel(characterDetails.dateOfBirth)}
                        </dd>
                      </>
                    )}
                    {characterDetails.age && (
                      <>
                        <dt>Age</dt>
                        <dd>{characterDetails.age}</dd>
                      </>
                    )}
                    {characterDetails.bloodType && (
                      <>
                        <dt>Blood type</dt>
                        <dd>{characterDetails.bloodType}</dd>
                      </>
                    )}
                    {characterDetails.favourites > 0 && (
                      <>
                        <dt>AniList favorites</dt>
                        <dd>{characterDetails.favourites.toLocaleString()}</dd>
                      </>
                    )}
                  </dl>
                  {characterDetails.siteUrl && (
                    <a
                      href={characterDetails.siteUrl}
                      target="_blank"
                      rel="noreferrer"
                    >
                      View on AniList
                    </a>
                  )}
                </div>
              </div>
              {characterDetails.description && (
                <section className="mt-4">
                  <h3 className="h5">About</h3>
                  <p className="mb-0" style={{ whiteSpace: "pre-line" }}>
                    {characterDetails.description}
                  </p>
                </section>
              )}
              <section className="mt-4">
                <h3 className="h5">Appears in</h3>
                {characterDetails.media.edges.length === 0 ? (
                  <p className="text-muted mb-0">
                    AniList does not list associated works for this character.
                  </p>
                ) : (
                  <Row className="g-3">
                    {characterDetails.media.edges.map((edge) => {
                      const workTitle = mediaTitleOf(edge.node);
                      const workCover =
                        edge.node.coverImage.large ||
                        edge.node.coverImage.medium;
                      return (
                        <Col key={edge.node.id} sm={6}>
                          <button
                            type="button"
                            className="character-media-item"
                            onClick={() => {
                              if (edge.node.type === "MANGA") {
                                setSelectedCharacterID(null);
                                navigate(`/anilist/manga/${edge.node.id}`);
                              }
                            }}
                            disabled={edge.node.type !== "MANGA"}
                          >
                            {workCover ? (
                              <img
                                src={workCover}
                                alt={workTitle}
                                referrerPolicy="no-referrer"
                              />
                            ) : (
                              <div
                                className="character-media-placeholder"
                                aria-hidden="true"
                              >
                                ?
                              </div>
                            )}
                            <span>
                              <strong>{workTitle}</strong>
                              <small>
                                {edge.node.type === "MANGA"
                                  ? "Manga"
                                  : edge.node.type}
                                {edge.characterRole
                                  ? ` · ${characterRoleLabel(edge.characterRole)}`
                                  : ""}
                              </small>
                            </span>
                          </button>
                        </Col>
                      );
                    })}
                  </Row>
                )}
              </section>
            </>
          )}
        </Modal.Body>
      </Modal>
    </Container>
  );
};

export default AniListMangaDetailsPage;
