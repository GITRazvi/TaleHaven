import React, { useCallback, useEffect, useMemo, useState } from "react";
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
import { useNavigate } from "react-router-dom";
import { readBrowserCache, writeBrowserCache } from "../../app/browser-cache";
import Backend from "../../app/clients/backend/backend";
import {
  AniListCharacterDetails,
  AniListCharacterFavorite,
  AniListLibraryItem,
  AniListLibraryStats,
  AniListReadingStatus,
  MangaDexFavorite,
} from "../../app/clients/backend/types";
import { useAuth } from "../../contexts/AuthContext";
import { presentationThemeStyle, themeBackgroundStyle, customThemeStyle } from "../../app/theme-style";
import CharacterDetailsModal from "../manga/character-details-modal";

const PAGE_SIZE = 100;
const PROFILE_CACHE_MAX_AGE_MS = 10 * 60 * 1000;

const statusOptions: Array<{
  status: AniListReadingStatus;
  title: string;
  color: string;
}> = [
  { status: "CURRENTLY_READING", title: "Reading", color: "#35c759" },
  { status: "READ", title: "Completed", color: "#4c82df" },
  { status: "ON_HOLD", title: "On Hold", color: "#f7c948" },
  { status: "DROPPED", title: "Dropped", color: "#e05252" },
  { status: "PLAN_TO_READ", title: "Plan to Read", color: "#b9c0cc" },
];

type StatusFilter = AniListReadingStatus | "ALL";

type CachedProfile = {
  library: AniListLibraryItem[];
  stats: AniListLibraryStats;
  favoriteAniListIDs: number[];
  mangaDexFavorites: MangaDexFavorite[];
  characterFavorites: AniListCharacterFavorite[];
};

const profileCacheKey = (userID: number | string) =>
  `talehaven:profile:${userID}`;

const profileThemeBackgrounds: Record<string, string> = {
  midnight: "linear-gradient(145deg, #0b1220, #182b4b)",
  forest: "linear-gradient(145deg, #0a1d18, #1a473a)",
  plum: "linear-gradient(145deg, #1b1022, #432652)",
  ember: "linear-gradient(145deg, #201310, #503127)",
  parchment: "linear-gradient(145deg, #e8dfca, #d6bea0)",
};

const emptyStats = (): AniListLibraryStats => ({
  total: 0,
  byStatus: {
    PLAN_TO_READ: 0,
    CURRENTLY_READING: 0,
    ON_HOLD: 0,
    READ: 0,
    DROPPED: 0,
  },
  topGenres: [],
  chaptersRead: 0,
  lastMangaCompleted: null,
});

const titleOf = (item: AniListLibraryItem) =>
  item.manga.title.english ||
  item.manga.title.userPreferred ||
  item.manga.title.romaji ||
  "Untitled manga";
const statusOf = (item: AniListLibraryItem): AniListReadingStatus =>
  item.readingStatus || "PLAN_TO_READ";
const characterNameOf = (character: Pick<AniListCharacterFavorite, "name">) =>
  character.name.full ||
  character.name.userPreferred ||
  character.name.native ||
  "Unnamed character";
const characterRoleLabel = (role: string) =>
  role
    ? role.charAt(0) + role.slice(1).toLowerCase().replace(/_/g, " ")
    : "Character";
const characterDateLabel = (date: AniListCharacterDetails["dateOfBirth"]) =>
  [date.year, date.month, date.day].filter(Boolean).join("-");
const mediaTitleOf = (
  media: AniListCharacterDetails["media"]["edges"][number]["node"],
) =>
  media.title.english ||
  media.title.userPreferred ||
  media.title.romaji ||
  media.title.native ||
  "Untitled work";

const LibraryPage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const navigate = useNavigate();
  const { user } = useAuth();
  const [library, setLibrary] = useState<AniListLibraryItem[]>([]);
  const [stats, setStats] = useState<AniListLibraryStats>(emptyStats);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [removeActionID, setRemoveActionID] = useState<number | null>(null);
  const [favoriteAniListIDs, setFavoriteAniListIDs] = useState<Set<number>>(
    new Set(),
  );
  const [favoriteActionID, setFavoriteActionID] = useState<number | null>(null);
  const [mangaDexFavorites, setMangaDexFavorites] = useState<
    MangaDexFavorite[]
  >([]);
  const [mangaDexFavoriteActionID, setMangaDexFavoriteActionID] = useState<
    string | null
  >(null);
  const [characterFavorites, setCharacterFavorites] = useState<
    AniListCharacterFavorite[]
  >([]);
  const [characterFavoriteActionID, setCharacterFavoriteActionID] = useState<
    number | null
  >(null);
  const [selectedCharacterID, setSelectedCharacterID] = useState<number | null>(null);
  const [characterDetails, setCharacterDetails] = useState<AniListCharacterDetails | null>(null);
  const [characterDetailsError, setCharacterDetailsError] = useState("");
  const [characterDetailsLoading, setCharacterDetailsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isUsingFallback, setIsUsingFallback] = useState(false);
  const [profileError, setProfileError] = useState("");
  const cacheKey = user ? profileCacheKey(user.id) : null;

  const displayTheme = user?.displayTheme || "midnight";
  const customThemeColor = user?.customThemeColor || "";
  const profileThemeImageUrl = user?.themeImagePath ? `${backend.getBaseUrl()}${user.themeImagePath}` : "";
  const profileThemeStyle = presentationThemeStyle(customThemeColor, profileThemeImageUrl);

  useEffect(() => {
    const customStyle = customThemeStyle(customThemeColor);
    const paletteBackground = customStyle
      ? `linear-gradient(145deg, ${customStyle["--theme-page"]}, ${customStyle["--theme-surface-strong"]})`
      : profileThemeBackgrounds[displayTheme] || profileThemeBackgrounds.midnight;
	const imageStyle = themeBackgroundStyle(profileThemeImageUrl);
	const background = imageStyle?.backgroundImage || paletteBackground;
    const previousBodyBackground = document.body.style.background;
    const previousBodyBackgroundColor = document.body.style.backgroundColor;
    const previousDocumentBackground = document.documentElement.style.background;
    const previousDocumentBackgroundColor = document.documentElement.style.backgroundColor;
    document.body.style.background = background;
    document.documentElement.style.background = background;
    return () => {
      document.body.style.background = previousBodyBackground;
      document.body.style.backgroundColor = previousBodyBackgroundColor;
      document.documentElement.style.background = previousDocumentBackground;
      document.documentElement.style.backgroundColor = previousDocumentBackgroundColor;
    };
  }, [displayTheme, customThemeColor, profileThemeImageUrl]);

  const load = useCallback(
    async (offset: number, append: boolean) => {
      append ? setLoadingMore(true) : setLoading(true);
      setError("");
      setIsUsingFallback(false);
      try {
        const [
          response,
          favoriteResponse,
          mangaDexFavoriteResponse,
          characterFavoriteResponse,
        ] = await Promise.all([
          backend.getAniListMyLibrary({ limit: PAGE_SIZE, offset }),
          backend.getAniListFavoriteIDs(),
          backend.getMangaDexFavorites(),
          backend.getAniListCharacterFavorites(),
        ]);
        setLibrary((current) =>
          append
            ? [
                ...current,
                ...response.data.filter(
                  (item) =>
                    !current.some(
                      (known) => known.anilistId === item.anilistId,
                    ),
                ),
              ]
            : response.data,
        );
        setStats(response.stats);
        setFavoriteAniListIDs(new Set(favoriteResponse.data));
        setMangaDexFavorites(mangaDexFavoriteResponse.manga);
        setCharacterFavorites(characterFavoriteResponse.data);
        if (!append && cacheKey) {
          writeBrowserCache<CachedProfile>(cacheKey, {
            library: response.data,
            stats: response.stats,
            favoriteAniListIDs: favoriteResponse.data,
            mangaDexFavorites: mangaDexFavoriteResponse.manga,
            characterFavorites: characterFavoriteResponse.data,
          });
        }
      } catch (err) {
        const cached = !append && cacheKey
          ? readBrowserCache<CachedProfile>(cacheKey, PROFILE_CACHE_MAX_AGE_MS)
          : null;
        if (cached) {
          setLibrary(cached.library);
          setStats(cached.stats);
          setFavoriteAniListIDs(new Set(cached.favoriteAniListIDs));
          setMangaDexFavorites(cached.mangaDexFavorites);
          setCharacterFavorites(cached.characterFavorites);
          setIsUsingFallback(true);
        }
        setError((err as Error).message || "Unable to load your profile.");
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [backend, cacheKey],
  );

  useEffect(() => {
    void load(0, false);
  }, [load]);

  useEffect(() => {
    if (!selectedCharacterID) { setCharacterDetails(null); setCharacterDetailsError(""); return; }
    let isCurrent = true;
    setCharacterDetails(null); setCharacterDetailsError(""); setCharacterDetailsLoading(true);
    void backend.getAniListCharacterDetails(selectedCharacterID).then((response) => {
      if (isCurrent) setCharacterDetails(response.data);
    }).catch((err) => { if (isCurrent) setCharacterDetailsError((err as Error).message || "Unable to load character details."); }).finally(() => { if (isCurrent) setCharacterDetailsLoading(false); });
    return () => { isCurrent = false; };
  }, [backend, selectedCharacterID]);

  const toggleFavorite = async (
    event: React.MouseEvent<HTMLButtonElement>,
    item: AniListLibraryItem,
  ) => {
    event.stopPropagation();
    const favorite = favoriteAniListIDs.has(item.anilistId);
    setFavoriteActionID(item.anilistId);
    setProfileError("");
    try {
      if (favorite)
        await backend.removeAniListMangaFromFavorites(item.anilistId);
      else await backend.addAniListMangaToFavorites(item.anilistId);
      setFavoriteAniListIDs((current) => {
        const next = new Set(current);
        if (favorite) next.delete(item.anilistId);
        else next.add(item.anilistId);
        return next;
      });
    } catch (err) {
      setProfileError(
        (err as Error).message ||
          `Unable to update favorites for ${titleOf(item)}.`,
      );
    } finally {
      setFavoriteActionID(null);
    }
  };

  const removeFromLibrary = async (
    event: React.MouseEvent<HTMLButtonElement>,
    item: AniListLibraryItem,
  ) => {
    event.stopPropagation();
    const previousStatus = statusOf(item);
    setRemoveActionID(item.anilistId);
    setProfileError("");
    try {
      await backend.removeAniListMangaFromMyLibrary(item.anilistId);
      setLibrary((current) =>
        current.filter((saved) => saved.anilistId !== item.anilistId),
      );
      setStats((current) => ({
        ...current,
        total: Math.max(0, current.total - 1),
        byStatus: {
          ...current.byStatus,
          [previousStatus]: Math.max(0, current.byStatus[previousStatus] - 1),
        },
      }));
    } catch (err) {
      setProfileError(
        (err as Error).message ||
          `Unable to remove ${titleOf(item)} from your library.`,
      );
    } finally {
      setRemoveActionID(null);
    }
  };

  const removeMangaDexFavorite = async (
    event: React.MouseEvent<HTMLButtonElement>,
    favorite: MangaDexFavorite,
  ) => {
    event.stopPropagation();
    setMangaDexFavoriteActionID(favorite.mangaId);
    setProfileError("");
    try {
      await backend.removeMangaFromFavorites(favorite.mangaId);
      setMangaDexFavorites((current) =>
        current.filter((item) => item.mangaId !== favorite.mangaId),
      );
    } catch (err) {
      setProfileError(
        (err as Error).message ||
          `Unable to remove ${favorite.title} from favorites.`,
      );
    } finally {
      setMangaDexFavoriteActionID(null);
    }
  };

  const removeCharacterFavorite = async (
    event: React.MouseEvent<HTMLButtonElement>,
    favorite: AniListCharacterFavorite,
  ) => {
    event.stopPropagation();
    setCharacterFavoriteActionID(favorite.characterId);
    setProfileError("");
    try {
      await backend.removeAniListCharacterFromFavorites(favorite.characterId);
      setCharacterFavorites((current) =>
        current.filter((item) => item.characterId !== favorite.characterId),
      );
    } catch (err) {
      setProfileError(
        (err as Error).message ||
          `Unable to remove ${characterNameOf(favorite)} from favorite characters.`,
      );
    } finally {
      setCharacterFavoriteActionID(null);
    }
  };

  const displayName = user?.displayName || user?.email.split("@")[0] || "Reader";
  const profileImageUrl = user?.profileImagePath
    ? `${process.env.REACT_APP_API_BASE_URL ?? "http://localhost:3001"}${user.profileImagePath}`
    : "";
  const avatarInitial = displayName.trim().charAt(0).toUpperCase() || "R";
  const filteredLibrary =
    statusFilter === "ALL"
      ? library
      : library.filter((item) => statusOf(item) === statusFilter);
  const favoriteAniListManga = library.filter((item) =>
    favoriteAniListIDs.has(item.anilistId),
  );
  const currentlyReading = library.filter(
    (item) => statusOf(item) === "CURRENTLY_READING",
  );
  const showProfileContent = !error || isUsingFallback;

  return (
    <Container className={`py-5 profile-display-theme profile-theme-${displayTheme}`} style={profileThemeStyle}>
      <Card className="mb-4 profile-identity-card shadow-sm">
        <Card.Body className="p-4 d-flex flex-wrap align-items-center gap-3">
          {profileImageUrl ? (
            <img className="profile-avatar-image" src={profileImageUrl} alt={`${displayName}'s profile`} />
          ) : (
            <div className="profile-avatar-fallback" aria-label="Profile photo placeholder">{avatarInitial}</div>
          )}
          <div className="flex-grow-1">
            <p className="text-uppercase small text-muted mb-1">My Profile</p>
            <h1 className="mb-1">{displayName}</h1>
            <p className="text-muted mb-0">
              Member since {user ? new Date(user.created_at).toLocaleDateString() : "today"} · {stats.total} saved {stats.total === 1 ? "title" : "titles"}
            </p>
          </div>
          <Button variant="outline-primary" onClick={() => navigate("/profile")}>Edit profile</Button>
        </Card.Body>
      </Card>

      <section aria-labelledby="profile-statistics-heading" className="mb-4">
        <Card className="profile-stats-card shadow-sm">
          <Card.Body className="p-3 p-md-4">
            <div className="d-flex justify-content-between align-items-center gap-3 mb-3">
              <h2
                id="profile-statistics-heading"
                className="profile-stats-title h5 mb-0"
              >
                Manga Stats
              </h2>
              <span className="profile-stats-total">
                <span>Total entries</span>
                <strong>{stats.total}</strong>
              </span>
            </div>
            <div
              className="profile-stat-progress"
              role="img"
              aria-label="Manga status distribution"
            >
              {statusOptions.map((option) => {
                const count = stats.byStatus[option.status];
                return (
                  count > 0 && (
                    <span
                      key={option.status}
                      title={`${option.title}: ${count}`}
                      style={{
                        backgroundColor: option.color,
                        flexBasis: `${(count / Math.max(stats.total, 1)) * 100}%`,
                      }}
                    />
                  )
                );
              })}
            </div>
            <div className="profile-stat-grid mt-3">
              <div className="profile-status-list">
                {statusOptions.map((option) => (
                  <div key={option.status} className="profile-status-item">
                    <span
                      className="profile-status-dot"
                      style={{ backgroundColor: option.color }}
                    />
                    <span>{option.title}</span>
                    <strong>{stats.byStatus[option.status]}</strong>
                  </div>
                ))}
              </div>
              <div className="profile-stat-details">
                <div>
                  <span>Chapters read</span>
                  <strong>{stats.chaptersRead}</strong>
                </div>
                <div className="profile-last-manga">
                  <span>Last manga completed</span>
                  {stats.lastMangaCompleted ? (
                    <div className="profile-last-manga-value">
                      {stats.lastMangaCompleted.coverUrl && (
                        <img
                          className="profile-last-manga-cover"
                          src={stats.lastMangaCompleted.coverUrl}
                          alt={`Cover of ${stats.lastMangaCompleted.title}`}
                          referrerPolicy="no-referrer"
                        />
                      )}
                      <strong className="profile-last-manga-title">
                        {stats.lastMangaCompleted.title}
                      </strong>
                    </div>
                  ) : (
                    <strong>—</strong>
                  )}
                </div>
                <div>
                  <span>
                    Top genre{stats.topGenres.length === 1 ? "" : "s"}
                  </span>
                  <strong>
                    {stats.topGenres.length ? stats.topGenres.join(" · ") : "—"}
                  </strong>
                </div>
              </div>
            </div>
          </Card.Body>
        </Card>
      </section>

      {!loading && showProfileContent && currentlyReading.length > 0 && (
        <section className="mb-5" aria-labelledby="currently-reading-heading">
          <div className="d-flex justify-content-between align-items-end gap-3 mb-3">
            <div>
              <h2 id="currently-reading-heading" className="h3 mb-1">Currently Reading</h2>
              <p className="text-muted mb-0">Pick up one of the titles you are actively reading.</p>
            </div>
            <Button variant="link" onClick={() => setStatusFilter("CURRENTLY_READING")}>View all</Button>
          </div>
          <Row className="g-3">
            {currentlyReading.slice(0, 6).map((item) => {
              const title = titleOf(item);
              const cover = item.manga.coverImage.large || item.manga.coverImage.medium;
              return (
                <Col key={`reading-${item.anilistId}`} xs={6} sm={4} md={3} lg={2}>
                  <Card className="profile-reading-card h-100" role="button" aria-label={`Open ${title}`} onClick={() => navigate(`/anilist/manga/${item.anilistId}`)}>
                    {cover ? <Card.Img variant="top" src={cover} alt={title} referrerPolicy="no-referrer" /> : <div className="library-cover-placeholder">No cover</div>}
                    <Card.Body className="p-2"><strong className="profile-reading-title">{title}</strong></Card.Body>
                  </Card>
                </Col>
              );
            })}
          </Row>
        </section>
      )}

      <div className="d-flex flex-wrap justify-content-between align-items-end gap-3 mb-4">
        <div>
          <h2 className="h3 mb-1">Books</h2>
          <p className="text-muted mb-0">
            Organize saved books without changing chapter progress.
          </p>
        </div>
        <div className="d-flex gap-2 align-items-end">
          <Form.Group>
            <Form.Label className="small mb-1" htmlFor="profile-status-filter">
              Filter
            </Form.Label>
            <Form.Select
              id="profile-status-filter"
              aria-label="Filter books by status"
              value={statusFilter}
              onChange={(event) =>
                setStatusFilter(event.target.value as StatusFilter)
              }
            >
              <option value="ALL">All statuses</option>
              {statusOptions.map((option) => (
                <option key={option.status} value={option.status}>
                  {option.title}
                </option>
              ))}
            </Form.Select>
          </Form.Group>
          <Button onClick={() => navigate("/")}>Browse manga</Button>
        </div>
      </div>
      {loading && (
        <div className="text-center p-5">
          <Spinner animation="border" size="sm" className="me-2" />
          Loading your profile...
        </div>
      )}
      {!loading && error && (
        <Alert variant={isUsingFallback ? "warning" : "danger"}>
          {isUsingFallback
            ? "AniList is temporarily unavailable. Showing saved profile results from this browser."
            : error}
          <Button
            className="ms-2"
            variant={isUsingFallback ? "outline-dark" : "outline-light"}
            size="sm"
            onClick={() => void load(0, false)}
          >
            Retry
          </Button>
        </Alert>
      )}
      {!loading && profileError && (
        <Alert variant="danger" dismissible onClose={() => setProfileError("")}>
          {profileError}
        </Alert>
      )}
      {!loading && showProfileContent && filteredLibrary.length === 0 && (
        <Card className="text-center p-4">
          <Card.Body>
            {library.length === 0
              ? "Your profile does not have any saved books yet."
              : "No saved books match this filter."}
          </Card.Body>
        </Card>
      )}
      {!loading && showProfileContent && filteredLibrary.length > 0 && (
        <Row>
          {filteredLibrary.map((item) => {
            const title = titleOf(item);
            const cover =
              item.manga.coverImage.large || item.manga.coverImage.medium;
            const isFavorite = favoriteAniListIDs.has(item.anilistId);
            return (
              <Col
                key={item.anilistId}
                xs={6}
                sm={4}
                md={3}
                lg={2}
                className="mb-4"
              >
                <Card
                  className="manga-library-card profile-manga-tile h-100 shadow-sm"
                  role="button"
                  aria-label={title}
                  onClick={() => navigate(`/anilist/manga/${item.anilistId}`)}
                >
                  <div className="position-relative">
                    {cover ? (
                      <Card.Img
                        variant="top"
                        src={cover}
                        alt={title}
                        referrerPolicy="no-referrer"
                        style={{ height: "280px", objectFit: "cover" }}
                      />
                    ) : (
                      <div className="library-cover-placeholder">No cover</div>
                    )}
                    <button
                      type="button"
                      className="home-library-toggle"
                      onClick={(event) => void removeFromLibrary(event, item)}
                      disabled={removeActionID === item.anilistId}
                      aria-label={`Remove ${title} from My Library`}
                    >
                      {removeActionID === item.anilistId ? "…" : "−"}
                    </button>
                    <button
                      type="button"
                      className={`home-library-heart ${isFavorite ? "is-saved" : ""}`}
                      onClick={(event) => void toggleFavorite(event, item)}
                      disabled={favoriteActionID === item.anilistId}
                      aria-label={
                        isFavorite
                          ? `Remove ${title} from Favorites`
                          : `Add ${title} to Favorites`
                      }
                    >
                      {favoriteActionID === item.anilistId
                        ? "…"
                        : isFavorite
                          ? "♥"
                          : "♡"}
                    </button>
                    <div
                      className="profile-manga-title-overlay"
                      aria-hidden="true"
                    >
                      {title}
                    </div>
                  </div>
                </Card>
              </Col>
            );
          })}
        </Row>
      )}
      {!loading && showProfileContent && library.length < stats.total && (
        <div className="text-center">
          <Button
            variant="outline-primary"
            onClick={() => void load(library.length, true)}
            disabled={loadingMore}
          >
            {loadingMore
              ? "Loading..."
              : `Load more (${library.length} of ${stats.total})`}
          </Button>
        </div>
      )}
      {!loading &&
        showProfileContent &&
        (favoriteAniListManga.length > 0 ||
          mangaDexFavorites.length > 0 ||
          characterFavorites.length > 0) && (
          <section className="mt-5" aria-labelledby="favorites-heading">
            <div className="mb-3">
              <h2 id="favorites-heading" className="h3 mb-1">
                Favorites
              </h2>
              <p className="text-muted mb-0">
                Your saved manga and favorite characters.
              </p>
            </div>
            <Card className="profile-favorites-card shadow-sm">
              <Card.Body className="p-3 p-md-4">
                {(favoriteAniListManga.length > 0 ||
                  mangaDexFavorites.length > 0) && (
                  <section
                    className="profile-favorite-group"
                    aria-labelledby="favorite-manga-heading"
                  >
                    <h3 id="favorite-manga-heading" className="h5 mb-3">
                      Manga
                    </h3>
                    <Row>
                      {favoriteAniListManga.map((item) => {
                        const title = titleOf(item);
                        const cover =
                          item.manga.coverImage.large ||
                          item.manga.coverImage.medium;
                        return (
                          <Col
                            key={`anilist-${item.anilistId}`}
                            xs={6}
                            sm={4}
                            md={3}
                            lg={2}
                            className="mb-4"
                          >
                            <Card
                              className="manga-library-card profile-manga-tile h-100 shadow-sm"
                              role="button"
                              aria-label={title}
                              onClick={() =>
                                navigate(`/anilist/manga/${item.anilistId}`)
                              }
                            >
                              <div className="position-relative">
                                {cover ? (
                                  <Card.Img
                                    variant="top"
                                    src={cover}
                                    alt={title}
                                    referrerPolicy="no-referrer"
                                    style={{
                                      height: "280px",
                                      objectFit: "cover",
                                    }}
                                  />
                                ) : (
                                  <div className="library-cover-placeholder">
                                    No cover
                                  </div>
                                )}
                                <button
                                  type="button"
                                  className="home-library-heart is-saved"
                                  onClick={(event) =>
                                    void toggleFavorite(event, item)
                                  }
                                  disabled={favoriteActionID === item.anilistId}
                                  aria-label={`Remove ${title} from Favorites`}
                                >
                                  {favoriteActionID === item.anilistId
                                    ? "…"
                                    : "♥"}
                                </button>
                                <div
                                  className="profile-manga-title-overlay"
                                  aria-hidden="true"
                                >
                                  {title}
                                </div>
                              </div>
                            </Card>
                          </Col>
                        );
                      })}
                      {mangaDexFavorites.map((favorite) => (
                        <Col
                          key={`mangadex-${favorite.mangaId}`}
                          xs={6}
                          sm={4}
                          md={3}
                          lg={2}
                          className="mb-4"
                        >
                          <Card
                            className="manga-library-card profile-manga-tile h-100 shadow-sm"
                            role="button"
                            aria-label={favorite.title}
                            onClick={() =>
                              navigate(`/manga/${favorite.mangaId}`)
                            }
                          >
                            <div className="position-relative">
                              {favorite.coverUrl ? (
                                <Card.Img
                                  variant="top"
                                  src={favorite.coverUrl}
                                  alt={favorite.title}
                                  referrerPolicy="no-referrer"
                                  style={{
                                    height: "280px",
                                    objectFit: "cover",
                                  }}
                                />
                              ) : (
                                <div className="library-cover-placeholder">
                                  No cover
                                </div>
                              )}
                              <button
                                type="button"
                                className="home-library-heart is-saved"
                                onClick={(event) =>
                                  void removeMangaDexFavorite(event, favorite)
                                }
                                disabled={
                                  mangaDexFavoriteActionID === favorite.mangaId
                                }
                                aria-label={`Remove ${favorite.title} from Favorites`}
                              >
                                {mangaDexFavoriteActionID === favorite.mangaId
                                  ? "…"
                                  : "♥"}
                              </button>
                              <div
                                className="profile-manga-title-overlay"
                                aria-hidden="true"
                              >
                                {favorite.title}
                              </div>
                            </div>
                          </Card>
                        </Col>
                      ))}
                    </Row>
                  </section>
                )}
                {characterFavorites.length > 0 && (
                  <section
                    className="profile-favorite-group"
                    aria-labelledby="favorite-characters-heading"
                  >
                    <h3 id="favorite-characters-heading" className="h5 mb-3">
                      Characters
                    </h3>
                    <Row className="g-3">
                      {characterFavorites.map((favorite) => {
                        const name = characterNameOf(favorite);
                        const image =
                          favorite.image.large || favorite.image.medium;
                        return (
                          <Col key={favorite.characterId} xs="auto">
                            <Card className="profile-character-favorite h-100">
                              <button
                                type="button"
                                className="profile-character-details-trigger"
                                onClick={() => setSelectedCharacterID(favorite.characterId)}
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
                                    className="profile-character-favorite-placeholder"
                                    aria-hidden="true"
                                  >
                                    ?
                                  </div>
                                )}
                                <span className="profile-character-name-overlay">
                                  {name}
                                </span>
                              </button>
                              <button
                                type="button"
                                className="home-library-heart is-saved ms-auto"
                                onClick={(event) =>
                                  void removeCharacterFavorite(event, favorite)
                                }
                                disabled={
                                  characterFavoriteActionID ===
                                  favorite.characterId
                                }
                                aria-label={`Remove ${name} from favorite characters`}
                              >
                                {characterFavoriteActionID ===
                                favorite.characterId
                                  ? "…"
                                  : "♥"}
                              </button>
                            </Card>
                          </Col>
                        );
                      })}
                    </Row>
                  </section>
                )}
              </Card.Body>
            </Card>
          </section>
        )}
      <Modal show={false} onHide={() => setSelectedCharacterID(null)} size="lg" centered>
        <Modal.Header closeButton><Modal.Title>{characterDetails ? characterNameOf(characterDetails) : "Character details"}</Modal.Title></Modal.Header>
        <Modal.Body>
          {characterDetailsLoading && <div className="text-center py-4"><Spinner animation="border" size="sm" className="me-2" />Loading character details...</div>}
          {characterDetailsError && <Alert variant="danger" className="mb-0">{characterDetailsError}</Alert>}
          {characterDetails && <>
            <div className="character-details-summary">
              {(characterDetails.image.large || characterDetails.image.medium) && <img src={characterDetails.image.large || characterDetails.image.medium} alt={characterNameOf(characterDetails)} referrerPolicy="no-referrer" />}
              <div><dl className="character-details-facts mb-0">{characterDetails.gender && <><dt>Gender</dt><dd>{characterDetails.gender}</dd></>}{characterDateLabel(characterDetails.dateOfBirth) && <><dt>Born</dt><dd>{characterDateLabel(characterDetails.dateOfBirth)}</dd></>}{characterDetails.age && <><dt>Age</dt><dd>{characterDetails.age}</dd></>}{characterDetails.bloodType && <><dt>Blood type</dt><dd>{characterDetails.bloodType}</dd></>}</dl>{characterDetails.siteUrl && <a href={characterDetails.siteUrl} target="_blank" rel="noreferrer">View on AniList</a>}</div>
            </div>
            {characterDetails.description && <section className="mt-4"><h3 className="h5">About</h3><p className="mb-0" style={{ whiteSpace: "pre-line" }}>{characterDetails.description}</p></section>}
            <section className="mt-4"><h3 className="h5">Appears in</h3>{characterDetails.media.edges.length === 0 ? <p className="text-muted mb-0">AniList does not list associated works for this character.</p> : <Row className="g-3">{characterDetails.media.edges.map((edge) => { const workTitle = mediaTitleOf(edge.node); const workCover = edge.node.coverImage.large || edge.node.coverImage.medium; return <Col key={edge.node.id} sm={6}><button type="button" className="character-media-item" onClick={() => { if (edge.node.type === "MANGA") { setSelectedCharacterID(null); navigate(`/anilist/manga/${edge.node.id}`); } }} disabled={edge.node.type !== "MANGA"}>{workCover ? <img src={workCover} alt={workTitle} referrerPolicy="no-referrer" /> : <div className="character-media-placeholder" aria-hidden="true">?</div>}<span><strong>{workTitle}</strong><small>{edge.node.type === "MANGA" ? "Manga" : edge.node.type}{edge.characterRole ? ` · ${characterRoleLabel(edge.characterRole)}` : ""}</small></span></button></Col>; })}</Row>}</section>
          </>}
        </Modal.Body>
      </Modal>
      <CharacterDetailsModal
        characterID={selectedCharacterID}
        character={characterDetails}
        loading={characterDetailsLoading}
        error={characterDetailsError}
        onHide={() => setSelectedCharacterID(null)}
        onSelectManga={(anilistID) => {
          setSelectedCharacterID(null);
          navigate(`/anilist/manga/${anilistID}`);
        }}
      />
    </Container>
  );
};

export default LibraryPage;
