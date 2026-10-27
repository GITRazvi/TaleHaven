export type BookInfo = {
  Title: string;
  Authors: string[];
  Id: string;
  Description: string;
  Categories: string[];
  Thumbnail: string;
  Rating: number;
  RatingCount: number;
  PublishedDate: string;
};

export type ApiResult = "ok" | "not_found" | string;

export type PagedResponse<T> = {
  result: ApiResult;
  data: T[];
  total: number;
  limit: number;
  offset: number;
};

export type ApiMessageResponse = {
  result: ApiResult;
  message: string;
};

export type SearchBooksResponse = {
  booksList: BookInfo[];
};

export type GetBookResponse = {
  booksList: BookInfo;
};

export type MangaSearchParams = {
  title: string;
  limit?: number;
  offset?: number;
  includedTags?: string[];
  excludedTags?: string[];
};

export type MangaChapterParams = {
  limit?: number;
  offset?: number;
  translatedLanguage?: string[];
  language?: string;
  scanlationGroup?: string;
};

export type CoverQuality = "256" | "512";
export type ChapterQuality = "data" | "saver";

export type MangaSearchItem = Record<string, unknown>;
export type MangaDetails = Record<string, unknown>;
export type ChapterDetails = Record<string, unknown>;

export type MangaSearchResponse = PagedResponse<MangaSearchItem>;

export type MangaDetailsResponse = {
  result: ApiResult;
  data: MangaDetails;
};

export type ScanlationGroupOption = {
  id: string;
  name: string;
  chapterCount: number;
};

export type MangaChaptersResponse = PagedResponse<ChapterDetails> & {
  availableLanguages?: string[];
  scanlationGroups?: ScanlationGroupOption[];
  language?: string;
  scanlationGroup?: string;
};

export type ReadingSettingsResponse = {
  result: ApiResult;
  preferredTranslationLanguage: string;
    preferredReadingMode: 'vertical' | 'ltr' | 'rtl';
};

export type ScanlationGroupPreferenceResponse = {
  result: ApiResult;
  mangaId: string;
  language: string;
  scanlationGroup: string;
};

export type MangaCoverResponse = {
  result: ApiResult;
  coverUrl: string;
};

export type ChapterResponse = {
  result: ApiResult;
  data: ChapterDetails;
};

export type ChapterPagesResponse = {
  result: ApiResult;
  mangaId: string;
  chapterId: string;
  pages: string[];
  pageCount: number;
};

export type ChapterPageResponse = {
  result: ApiResult;
  mangaId: string;
  chapterId: string;
  pageNumber: number;
  totalPages: number;
  pageUrl: string;
};

export type ChapterPageCountResponse = {
  result: ApiResult;
  chapterId: string;
  totalPages: number;
};

export type SaveReadingProgressBody = {
  mangaId: string;
  chapterId: string;
  currentPage: number;  // Renamed from lastPage in API v3.0.0
  totalPages?: number;  // Optional in API v3.0.0
};

export type SyncReadingProgressBody = {
  chapterId: string;
};

export type ReadingProgressRecord = {
  userId?: number;  // Added in API v3.0.0
  mangaId: string;
  chapterId: string;
  currentPage: number;  // Renamed from lastPage in API v3.0.0
  totalPages?: number;
  lastReadAt?: string;  // Added in API v3.0.0 (renamed from updatedAt)
  updatedAt?: string;  // Deprecated but kept for backward compatibility
};

export type ReadingProgressAllResponse = {
  result: ApiResult;
  data: {
    userId?: number;  // Added in API v3.0.0
    mangas: Record<string, {
      mangaId: string;
      chapters: Record<string, {
        currentPage: number;
        totalPages?: number;
        lastReadAt: string;
      }>;
      updatedAt: string;
    }>;
    updatedAt?: string;
  };
};

export type ReadingProgressByChapterResponse = {
  result: ApiResult;
  data: ReadingProgressRecord;
  message?: string;
};

export type ReadingProgressByMangaResponse = {
  result: ApiResult;
  data: {
    mangaId: string;
    chapters: Record<string, {
      currentPage: number;
      totalPages?: number;
      lastReadAt: string;
    }>;
    updatedAt?: string;
  };
};

export type ReadingProgressMangaListResponse = {
  result: ApiResult;
  data: string[];
  count: number;
};

export type MangaLibraryItem = {
  mangadex_id: string;
  title: string;
  cover_image_path: string;
};

export type MangaLibraryResponse = {
  result: ApiResult;
  count: number;
  data: MangaLibraryItem[];
};

export type UserLibraryManga = {
  mangaId: string;
  title: string;
  coverUrl: string;
  addedAt: string;
};

export type UserLibraryResponse = {
  result: ApiResult;
  count: number;
  total: number;
  limit: number;
  offset: number;
  data: UserLibraryManga[];
};

export type UserLibraryParams = {
  limit?: number;
  offset?: number;
};

export type FavoriteIDsResponse<T extends string | number> = {
  result: ApiResult;
  data: T[];
};

export type MangaDexFavorite = {
  mangaId: string;
  title: string;
  coverUrl: string;
};

export type MangaDexFavoritesResponse = FavoriteIDsResponse<string> & {
  manga: MangaDexFavorite[];
};

export type ApiError = Error & {
  status?: number;
  details?: unknown;
};

// ==========================================
// Authentication Types
// ==========================================

export type User = {
  id: number;
  email: string;
  role: 'user' | 'admin';
  displayName: string;
  profileImagePath: string;
  displayTheme: DisplayTheme;
  customThemeColor: string;
	  themeImagePath: string;
  created_at: string;
};

export type DisplayTheme = 'midnight' | 'forest' | 'plum' | 'ember' | 'parchment';

export type RegisterRequest = {
  email: string;
  password: string;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type ChangePasswordRequest = {
  old_password: string;
  new_password: string;
};

export type ReadingSettingsRequest = {
  preferredTranslationLanguage: string;
    preferredReadingMode: 'vertical' | 'ltr' | 'rtl';
};

export type ScanlationGroupPreferenceRequest = {
  language: string;
  scanlationGroup: string;
};

export type AuthResponse = {
  message: string;
  user: User;
};

export type UserResponse = {
  user: User;
};

export type AdminDashboardStats = {
  users: number;
  aniListLibrary: number;
  mangaDexLibrary: number;
  favorites: number;
  trackedChapters: number;
};

export type AdminDashboardResponse = {
  result: ApiResult;
  stats: AdminDashboardStats;
};

export type ChatRoomVisibility = 'public' | 'invite_only';

export type ChatRoom = {
  id: number;
  name: string;
  visibility: ChatRoomVisibility;
  imagePath: string;
  theme: DisplayTheme;
  customThemeColor: string;
	  themeImagePath: string;
  ownerUserId: number;
  ownerDisplayName: string;
  memberCount: number;
  isMember: boolean;
  isOwner: boolean;
  createdAt: string;
};

export type ChatMessage = {
  id: number;
  roomId: number;
  userId: number;
  authorDisplayName: string;
  body: string;
  imagePath: string;
  createdAt: string;
};

export type ChatRoomsResponse = {
  result: ApiResult;
  data: ChatRoom[];
};

export type ChatRoomResponse = {
  result: ApiResult;
  room: ChatRoom;
};

export type ChatMessagesResponse = {
  result: ApiResult;
  data: ChatMessage[];
  limit: number;
};

export type CreateChatRoomRequest = {
  name: string;
  visibility: ChatRoomVisibility;
};

export type CreateChatRoomResponse = {
  result: ApiResult;
  room: ChatRoom;
};

export type JoinChatRoomResponse = ApiMessageResponse & {
  joined: boolean;
  roomId?: number;
};

export type CreateChatInviteResponse = {
  result: ApiResult;
  inviteToken: string;
};

export type SendChatMessageResponse = {
  result: ApiResult;
  message: ChatMessage;
};

export type UpdateProfileRequest = {
  displayName: string;
};

export type UpdateProfileThemeRequest = {
  displayTheme: DisplayTheme;
  customThemeColor?: string;
};

export type AuthErrorResponse = {
  error: string;
  message: string;
};

// ==========================================
// Popular Manga Types
// ==========================================

export type PopularManga = {
  id: string;
  title: string;
  description: string;
  coverUrl: string;
  authors: string[];
  artists: string[];
  status: string;
  year: number;
  tags: string[];
};

export type PopularMangaParams = {
  limit?: number;
  offset?: number;
  orderBy?: 'followedCount' | 'rating';
  includedTags?: string[];
  excludedTags?: string[];
};

export type PopularMangaResponse = {
  result: ApiResult;
  data: PopularManga[];
  total: number;
  limit: number;
  offset: number;
};

export type AniListDiscoverySection = 'trending' | 'popular' | 'top-rated' | 'recent';

export type AniListMangaSearchParams = {
  title?: string;
  genre?: string;
  tag?: string;
  year?: number;
  format?: string;
  status?: string;
  sort?: string;
  page?: number;
  perPage?: number;
};

export type AniListManga = {
  id: number;
  title: {
    romaji: string;
    english: string;
    native: string;
    userPreferred: string;
  };
  synonyms?: string[];
  description: string;
  coverImage: {
    large: string;
    medium: string;
    color: string;
  };
  format: string;
  status: string;
  startDate?: {
    year: number;
    month: number;
    day: number;
  };
  chapters: number;
  volumes: number;
  genres: string[];
  tags?: Array<{ name: string; rank: number; isMediaSpoiler: boolean }>;
  staff?: {
    edges: Array<{
      role: string;
      node: {
        id: number;
        name: { full: string; native: string; userPreferred: string };
      };
    }>;
  };
  averageScore: number;
  popularity: number;
  favourites?: number;
  rankings?: Array<{
    rank: number;
    type: string;
    allTime: boolean;
    context: string;
  }>;
  siteUrl: string;
  isAdult: boolean;
  characters?: {
    pageInfo: { currentPage: number; hasNextPage: boolean; perPage: number; total: number };
    edges: Array<{
      role: string;
      node: {
        id: number;
        name: { full: string; native: string; userPreferred: string };
        image: { large: string; medium: string };
      };
    }>;
  };
  relations?: {
    edges: Array<{
      relationType: string;
      node: {
        id: number;
        type: string;
        title: { romaji: string; english: string; native: string; userPreferred: string };
        coverImage: { large: string; medium: string; color: string };
        format: string;
        status: string;
        siteUrl: string;
      };
    }>;
  };
  externalLinks?: Array<{
    id: number;
    site: string;
    url: string;
    type: string;
    language: string;
  }>;
};

export type AniListDiscoveryResponse = {
  result: ApiResult;
  data: AniListManga[];
  total: number;
  page: number;
  perPage: number;
  hasNextPage: boolean;
};
export type AniListMangaDetailsResponse = { result: ApiResult; data: AniListManga };
export type AniListMangaCharactersResponse = {
  result: ApiResult;
  data: NonNullable<AniListManga["characters"]>;
};
export type AniListReadingSourceResponse = { result: ApiResult; available: boolean; mangaId?: string };

export type AniListCharacterFavorite = {
  characterId: number;
  mangaId: number;
  role: string;
  name: { full: string; native: string; userPreferred: string };
  image: { large: string; medium: string };
};
export type AniListCharacterFavoritesResponse = { result: ApiResult; data: AniListCharacterFavorite[] };
export type AniListCharacterDetails = {
  id: number;
  name: { full: string; native: string; userPreferred: string };
  image: { large: string; medium: string };
  description: string;
  gender: string;
  dateOfBirth: { year: number; month: number; day: number };
  age: string;
  bloodType: string;
  siteUrl: string;
  favourites: number;
  media: {
    pageInfo: { currentPage: number; hasNextPage: boolean; perPage: number; total: number };
    edges: Array<{
      characterRole: string;
      node: { id: number; type: string; title: { romaji: string; english: string; native: string; userPreferred: string }; coverImage: { large: string; medium: string; color: string }; format: string; status: string; siteUrl: string };
    }>;
  };
};
export type AniListCharacterDetailsResponse = { result: ApiResult; data: AniListCharacterDetails };

export type AniListReadingStatus = 'PLAN_TO_READ' | 'CURRENTLY_READING' | 'ON_HOLD' | 'READ' | 'DROPPED';
export type AniListLibraryItem = { anilistId: number; addedAt: string; readingStatus: AniListReadingStatus; personalRating: number | null; manga: AniListManga };
export type AniListLibraryStats = { total: number; byStatus: Record<AniListReadingStatus, number>; topGenres: string[]; chaptersRead: number; lastMangaCompleted: { anilistId: number; title: string; coverUrl: string } | null };
export type AniListLibraryResponse = PagedResponse<AniListLibraryItem> & { stats: AniListLibraryStats };

