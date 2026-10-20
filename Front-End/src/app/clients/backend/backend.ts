import axios, { AxiosError, AxiosInstance } from "axios";
import {
  ApiError,
  ApiMessageResponse,
  AuthResponse,
  ChangePasswordRequest,
  ChapterPagesResponse,
  ChapterQuality,
  ChapterResponse,
  CoverQuality,
  GetBookResponse,
  LoginRequest,
  MangaChapterParams,
  MangaChaptersResponse,
  MangaCoverResponse,
  MangaDetailsResponse,
  MangaLibraryResponse,
  MangaSearchParams,
  MangaSearchResponse,
  PopularMangaParams,
  PopularMangaResponse,
  ReadingProgressAllResponse,
  ReadingProgressByChapterResponse,
  ReadingProgressByMangaResponse,
  ReadingProgressMangaListResponse,
  RegisterRequest,
  SaveReadingProgressBody,
  SearchBooksResponse,
  SyncReadingProgressBody,
  UserResponse,
  UserLibraryParams,
  UserLibraryResponse,
} from "./types";

type QueryValue = string | number | boolean | null | undefined;

export default class Backend {
  private readonly BASEURL = process.env.REACT_APP_API_BASE_URL ?? "http://localhost:3001";
  private readonly client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: this.BASEURL,
      timeout: 60000,  // 60 seconds for chapter downloads
      withCredentials: true,  // Required for JWT cookie authentication (talehaven_auth_token)
    });
  }

  private buildQuery(query: Record<string, QueryValue | QueryValue[]>): string {
    const params = new URLSearchParams();

    Object.entries(query).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach((item) => {
          if (item !== undefined && item !== null && item !== "") {
            params.append(key, String(item));
          }
        });
        return;
      }

      if (value !== undefined && value !== null && value !== "") {
        params.append(key, String(value));
      }
    });

    const queryString = params.toString();
    return queryString ? `?${queryString}` : "";
  }

  private normalizeError(error: unknown): ApiError {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      const normalized = new Error(axiosError.message) as ApiError;
      normalized.status = axiosError.response?.status;
      normalized.details = axiosError.response?.data;
      return normalized;
    }

    return (error as ApiError) ?? new Error("Unknown API error");
  }

  private async get<T>(url: string): Promise<T> {
    try {
      const response = await this.client.get<T>(url);
      return response.data;
    } catch (error) {
      throw this.normalizeError(error);
    }
  }

  private async post<T>(url: string, body?: unknown): Promise<T> {
    try {
      const response = await this.client.post<T>(url, body);
      return response.data;
    } catch (error) {
      throw this.normalizeError(error);
    }
  }

  private async delete<T>(url: string): Promise<T> {
    try {
      const response = await this.client.delete<T>(url);
      return response.data;
    } catch (error) {
      throw this.normalizeError(error);
    }
  }

  private async put<T>(url: string, body?: unknown): Promise<T> {
    try {
      const response = await this.client.put<T>(url, body);
      return response.data;
    } catch (error) {
      throw this.normalizeError(error);
    }
  }

  // ==========================================
  // Authentication Methods
  // ==========================================

  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.post<AuthResponse>("/auth/register", data);
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.post<AuthResponse>("/auth/login", data);
  }

  async logout(): Promise<ApiMessageResponse> {
    return this.post<ApiMessageResponse>("/auth/logout");
  }

  async getCurrentUser(): Promise<UserResponse> {
    return this.get<UserResponse>("/auth/me");
  }

  async changePassword(data: ChangePasswordRequest): Promise<ApiMessageResponse> {
    return this.put<ApiMessageResponse>("/auth/password", data);
  }

  async deleteAccount(): Promise<ApiMessageResponse> {
    return this.delete<ApiMessageResponse>("/auth/account");
  }

  // ==========================================
  // Book and Manga Methods
  // ==========================================

  async testRoute(): Promise<{ test: string }> {
    return this.get<{ test: string }>("/test");
  }

  async searchBooks(query: string): Promise<SearchBooksResponse> {
    return this.get<SearchBooksResponse>(`/search/books${this.buildQuery({ query })}`);
  }

  async getBook(bookId: string): Promise<GetBookResponse> {
    return this.get<GetBookResponse>(`/search/book/${encodeURIComponent(bookId)}`);
  }

  async searchManga(params: MangaSearchParams): Promise<MangaSearchResponse> {
    const query = this.buildQuery({
      title: params.title,
      limit: params.limit,
      offset: params.offset,
      "includedTags[]": params.includedTags,
      "excludedTags[]": params.excludedTags,
    });

    return this.get<MangaSearchResponse>(`/manga/search${query}`);
  }

  async getPopularManga(params: PopularMangaParams = {}): Promise<PopularMangaResponse> {
    const query = this.buildQuery({
      limit: params.limit,
      offset: params.offset,
      orderBy: params.orderBy,
      "includedTags[]": params.includedTags,
      "excludedTags[]": params.excludedTags,
    });

    return this.get<PopularMangaResponse>(`/manga/popular${query}`);
  }

  async getManga(mangaId: string): Promise<MangaDetailsResponse> {
    return this.get<MangaDetailsResponse>(`/manga/${encodeURIComponent(mangaId)}`);
  }

  async getMangaChapters(mangaId: string, params: MangaChapterParams = {}): Promise<MangaChaptersResponse> {
    const query = this.buildQuery({
      limit: params.limit,
      offset: params.offset,
      "translatedLanguage[]": params.translatedLanguage,
    });

    return this.get<MangaChaptersResponse>(`/manga/${encodeURIComponent(mangaId)}/chapters${query}`);
  }

  async getMangaCover(mangaId: string, quality?: CoverQuality): Promise<MangaCoverResponse> {
    return this.get<MangaCoverResponse>(
      `/manga/${encodeURIComponent(mangaId)}/cover${this.buildQuery({ quality })}`
    );
  }

  async downloadMangaCover(mangaId: string, quality?: CoverQuality): Promise<ApiMessageResponse & { mangaId: string }> {
    return this.post<ApiMessageResponse & { mangaId: string }>(
      `/manga/${encodeURIComponent(mangaId)}/download-cover${this.buildQuery({ quality })}`
    );
  }

  async getChapter(chapterId: string): Promise<ChapterResponse> {
    return this.get<ChapterResponse>(`/chapter/${encodeURIComponent(chapterId)}`);
  }

  async getChapterPages(chapterId: string, quality?: ChapterQuality): Promise<ChapterPagesResponse> {
    return this.get<ChapterPagesResponse>(
      `/chapter/${encodeURIComponent(chapterId)}/pages${this.buildQuery({ quality })}`
    );
  }

  async downloadChapter(chapterId: string, quality?: ChapterQuality): Promise<ApiMessageResponse & { chapterId: string }> {
    return this.post<ApiMessageResponse & { chapterId: string }>(
      `/chapter/${encodeURIComponent(chapterId)}/download${this.buildQuery({ quality })}`
    );
  }

  async saveReadingProgress(body: SaveReadingProgressBody): Promise<ApiMessageResponse> {
    // Writes to Redis only (fast, ~1-2ms) - API v3.0.0
    // Use syncReadingProgress() to persist to PostgreSQL
    return this.post<ApiMessageResponse>("/reading-progress", body);
  }

  async syncReadingProgress(chapterId: string): Promise<ApiMessageResponse> {
    // NEW in API v3.0.0: Syncs Redis cache to PostgreSQL for permanent storage
    // Call when user leaves chapter or closes tab
    return this.post<ApiMessageResponse>("/reading-progress/sync", { chapterId });
  }

  async getReadingProgress(): Promise<ReadingProgressAllResponse> {
    return this.get<ReadingProgressAllResponse>("/reading-progress");
  }

  async deleteReadingProgress(): Promise<ApiMessageResponse> {
    return this.delete<ApiMessageResponse>("/reading-progress");
  }

  async getChapterProgress(chapterId: string, mangaId: string): Promise<ReadingProgressByChapterResponse> {
    return this.get<ReadingProgressByChapterResponse>(
      `/reading-progress/chapter/${encodeURIComponent(chapterId)}${this.buildQuery({ mangaId })}`
    );
  }

  async deleteChapterProgress(chapterId: string, mangaId: string): Promise<ApiMessageResponse> {
    return this.delete<ApiMessageResponse>(
      `/reading-progress/chapter/${encodeURIComponent(chapterId)}${this.buildQuery({ mangaId })}`
    );
  }

  async getMangaProgress(mangaId: string): Promise<ReadingProgressByMangaResponse> {
    return this.get<ReadingProgressByMangaResponse>(`/reading-progress/manga/${encodeURIComponent(mangaId)}`);
  }

  async deleteMangaProgress(mangaId: string): Promise<ApiMessageResponse> {
    return this.delete<ApiMessageResponse>(`/reading-progress/manga/${encodeURIComponent(mangaId)}`);
  }

  async getReadingProgressMangaList(): Promise<ReadingProgressMangaListResponse> {
    return this.get<ReadingProgressMangaListResponse>("/reading-progress/manga-list");
  }

  async getMangaLibrary(): Promise<MangaLibraryResponse> {
    return this.get<MangaLibraryResponse>("/manga/library");
  }

  async getMyLibrary(params: UserLibraryParams = {}): Promise<UserLibraryResponse> {
    return this.get<UserLibraryResponse>(`/library${this.buildQuery(params)}`);
  }

  async addMangaToMyLibrary(mangaId: string): Promise<ApiMessageResponse & { added: boolean }> {
    return this.post<ApiMessageResponse & { added: boolean }>(`/library/${encodeURIComponent(mangaId)}`);
  }

  async removeMangaFromMyLibrary(mangaId: string): Promise<ApiMessageResponse & { removed: boolean }> {
    return this.delete<ApiMessageResponse & { removed: boolean }>(`/library/${encodeURIComponent(mangaId)}`);
  }

  getImageUrl(path: string): string {
    return `${this.BASEURL}/images/${path}`;
  }

  getBaseUrl(): string {
    return this.BASEURL;
  }
}
