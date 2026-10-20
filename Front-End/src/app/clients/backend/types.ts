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

export type MangaChaptersResponse = PagedResponse<ChapterDetails>;

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
  created_at: string;
};

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

export type AuthResponse = {
  message: string;
  user: User;
};

export type UserResponse = {
  user: User;
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

