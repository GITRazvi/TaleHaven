import axios, { AxiosError, AxiosInstance } from "axios";
import { notifySessionExpired } from "../../../contexts/auth-session-events";
import {
  ApiError,
  ApiMessageResponse,
  AniListDiscoveryResponse,
  AniListDiscoverySection,
  AniListCharacterFavoritesResponse,
  AniListCharacterDetailsResponse,
  AniListMangaCharactersResponse,
  AniListMangaDetailsResponse,
  AniListMangaSearchParams,
  AniListReadingSourceResponse,
  AniListLibraryResponse,
  AniListReadingStatus,
  AdminDashboardResponse,
  AuthResponse,
  ChangePasswordRequest,
  ChapterPageResponse,
  ChapterPageCountResponse,
  ChapterPagesResponse,
  ChapterQuality,
  ChapterResponse,
  ChatMessagesResponse,
  ChatRoomResponse,
  ChatRoomsResponse,
  CreateChatInviteResponse,
  CreateChatRoomRequest,
  CreateChatRoomResponse,
  CoverQuality,
	DisplayTheme,
  FavoriteIDsResponse,
  GetBookResponse,
  LoginRequest,
  MangaChapterParams,
  MangaChaptersResponse,
  MangaCoverResponse,
  MangaDetailsResponse,
  MangaDexFavoritesResponse,
  MangaLibraryResponse,
  MangaSearchParams,
  MangaSearchResponse,
  JoinChatRoomResponse,
  PopularMangaParams,
  PopularMangaResponse,
  ReadingProgressAllResponse,
  ReadingProgressByChapterResponse,
  ReadingProgressByMangaResponse,
  ReadingProgressMangaListResponse,
  ReadingSettingsRequest,
  ReadingSettingsResponse,
  ScanlationGroupPreferenceRequest,
  ScanlationGroupPreferenceResponse,
  RegisterRequest,
  SaveReadingProgressBody,
  SearchBooksResponse,
  SyncReadingProgressBody,
  SendChatMessageResponse,
  UserResponse,
  UserLibraryParams,
  UserLibraryResponse,
  UpdateProfileRequest,
	UpdateProfileThemeRequest,
} from "./types";

type QueryValue = string | number | boolean | null | undefined;

export default class Backend {
  private readonly BASEURL = process.env.REACT_APP_API_BASE_URL ?? "http://localhost:3001";
  private readonly client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: this.BASEURL,
      timeout: 60000,  // 60 seconds for chapter downloads
      withCredentials: true,  // Required for HttpOnly server-session cookies.
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
      const requestUrl = axiosError.config?.url ?? "";
      // Login failures and the initial /auth/me restoration check are expected
      // authentication outcomes. Other 401s indicate the active session was
      // revoked or expired and should immediately clear React auth state.
      if (axiosError.response?.status === 401 && requestUrl !== "/auth/login" && requestUrl !== "/auth/register" && requestUrl !== "/auth/me") {
        notifySessionExpired();
      }
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

  private async patch<T>(url: string, body?: unknown): Promise<T> {
    try {
      const response = await this.client.patch<T>(url, body);
      return response.data;
    } catch (error) {
      throw this.normalizeError(error);
    }
  }

  private async postForm<T>(url: string, body: FormData): Promise<T> {
    try {
      const response = await this.client.post<T>(url, body);
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

  async getAdminDashboard(): Promise<AdminDashboardResponse> {
    return this.get<AdminDashboardResponse>("/admin/dashboard");
  }

  async getChatRooms(): Promise<ChatRoomsResponse> {
    return this.get<ChatRoomsResponse>("/chat/rooms");
  }

  async createChatRoom(data: CreateChatRoomRequest): Promise<CreateChatRoomResponse> {
    return this.post<CreateChatRoomResponse>("/chat/rooms", data);
  }

  async joinPublicChatRoom(roomId: number): Promise<JoinChatRoomResponse> {
    return this.post<JoinChatRoomResponse>(`/chat/rooms/${roomId}/join`);
  }

  async updateChatRoomVisibility(roomId: number, visibility: 'public' | 'invite_only'): Promise<ChatRoomResponse> {
    return this.patch<ChatRoomResponse>(`/chat/rooms/${roomId}/visibility`, { visibility });
  }

  async updateChatRoomTheme(roomId: number, theme: DisplayTheme, customThemeColor = ''): Promise<ChatRoomResponse> {
    return this.patch<ChatRoomResponse>(`/chat/rooms/${roomId}/theme`, { theme, customThemeColor });
  }

  async uploadChatRoomImage(roomId: number, image: File): Promise<ChatRoomResponse> {
    const formData = new FormData();
    formData.append("image", image);
    return this.postForm<ChatRoomResponse>(`/chat/rooms/${roomId}/image`, formData);
  }

  async uploadChatRoomThemeImage(roomId: number, image: File): Promise<ChatRoomResponse> {
    const formData = new FormData();
    formData.append("image", image);
    return this.postForm<ChatRoomResponse>(`/chat/rooms/${roomId}/theme-image`, formData);
  }

  async removeChatRoomThemeImage(roomId: number): Promise<ChatRoomResponse> {
    return this.delete<ChatRoomResponse>(`/chat/rooms/${roomId}/theme-image`);
  }

  async getChatRoom(roomId: number): Promise<ChatRoomResponse> {
    return this.get<ChatRoomResponse>(`/chat/rooms/${roomId}`);
  }

  async getChatMessages(roomId: number, limit = 100): Promise<ChatMessagesResponse> {
    return this.get<ChatMessagesResponse>(`/chat/rooms/${roomId}/messages${this.buildQuery({ limit })}`);
  }

  async sendChatMessage(roomId: number, body: string): Promise<SendChatMessageResponse> {
    return this.post<SendChatMessageResponse>(`/chat/rooms/${roomId}/messages`, { body });
  }

  async sendChatMessageWithImage(roomId: number, body: string, image: File): Promise<SendChatMessageResponse> {
    const formData = new FormData();
    formData.append("body", body);
    formData.append("image", image);
    return this.postForm<SendChatMessageResponse>(`/chat/rooms/${roomId}/messages/image`, formData);
  }

  async createChatInvite(roomId: number): Promise<CreateChatInviteResponse> {
    return this.post<CreateChatInviteResponse>(`/chat/rooms/${roomId}/invites`);
  }

  async joinChatInvite(token: string): Promise<JoinChatRoomResponse> {
    return this.post<JoinChatRoomResponse>(`/chat/invites/${encodeURIComponent(token)}/join`);
  }

  async changePassword(data: ChangePasswordRequest): Promise<ApiMessageResponse> {
    return this.put<ApiMessageResponse>("/auth/password", data);
  }

  async deleteAccount(): Promise<ApiMessageResponse> {
    return this.delete<ApiMessageResponse>("/auth/account");
  }

  async getProfile(): Promise<UserResponse> {
    return this.get<UserResponse>("/profile");
  }

  async updateProfile(data: UpdateProfileRequest): Promise<UserResponse> {
    return this.patch<UserResponse>("/profile", data);
  }

  async updateProfileTheme(data: UpdateProfileThemeRequest): Promise<UserResponse> {
    return this.patch<UserResponse>("/profile/theme", data);
  }

  async uploadProfilePhoto(photo: File): Promise<UserResponse> {
    const formData = new FormData();
    formData.append("photo", photo);
    return this.postForm<UserResponse>("/profile/photo", formData);
  }

  async uploadProfileThemeImage(image: File): Promise<UserResponse> {
    const formData = new FormData();
    formData.append("image", image);
    return this.postForm<UserResponse>("/profile/theme-image", formData);
  }

  async removeProfileThemeImage(): Promise<UserResponse> {
    return this.delete<UserResponse>("/profile/theme-image");
  }

  async getReadingSettings(): Promise<ReadingSettingsResponse> {
    return this.get<ReadingSettingsResponse>("/settings/reading");
  }

  async updateReadingSettings(data: ReadingSettingsRequest): Promise<ReadingSettingsResponse> {
    return this.patch<ReadingSettingsResponse>("/settings/reading", data);
  }

  async getPreferredScanlationGroup(mangaId: string, language: string): Promise<ScanlationGroupPreferenceResponse> {
    return this.get<ScanlationGroupPreferenceResponse>(`/settings/reading/manga/${encodeURIComponent(mangaId)}/scanlation-group${this.buildQuery({ language })}`);
  }

  async updatePreferredScanlationGroup(mangaId: string, data: ScanlationGroupPreferenceRequest): Promise<ScanlationGroupPreferenceResponse> {
    return this.patch<ScanlationGroupPreferenceResponse>(`/settings/reading/manga/${encodeURIComponent(mangaId)}/scanlation-group`, data);
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

  async getAniListDiscoverySection(
    section: AniListDiscoverySection,
    page = 1,
    perPage = 12,
  ): Promise<AniListDiscoveryResponse> {
    return this.get<AniListDiscoveryResponse>(
      `/anilist/manga/${encodeURIComponent(section)}${this.buildQuery({ page, perPage })}`
    );
  }

  async searchAniListManga(params: AniListMangaSearchParams): Promise<AniListDiscoveryResponse> {
    return this.get<AniListDiscoveryResponse>(`/anilist/manga/search${this.buildQuery({
      title: params.title,
      genre: params.genre,
      tag: params.tag,
      year: params.year,
      format: params.format,
      status: params.status,
      sort: params.sort,
      page: params.page ?? 1,
      perPage: params.perPage ?? 20,
    })}`);
  }

  async getAniListReadingSource(anilistId: number): Promise<AniListReadingSourceResponse> {
    return this.get<AniListReadingSourceResponse>(`/anilist/manga/reading-source/${anilistId}`);
  }

  async getAniListMangaByGenre(
    genre: string,
    page = 1,
    perPage = 12,
  ): Promise<AniListDiscoveryResponse> {
    return this.get<AniListDiscoveryResponse>(
      `/anilist/manga/genre/${encodeURIComponent(genre)}${this.buildQuery({ page, perPage })}`
    );
  }

  async getAniListMyLibrary(params: UserLibraryParams = {}): Promise<AniListLibraryResponse> {
    return this.get<AniListLibraryResponse>(`/library/anilist${this.buildQuery(params)}`);
  }

  async getAniListMangaDetails(anilistId: number): Promise<AniListMangaDetailsResponse> {
    return this.get<AniListMangaDetailsResponse>(`/anilist/manga/details/${anilistId}`);
  }

  async getAniListMangaCharacters(anilistId: number, page: number, perPage = 10): Promise<AniListMangaCharactersResponse> {
    const query = new URLSearchParams({ page: String(page), perPage: String(perPage) });
    return this.get<AniListMangaCharactersResponse>(`/anilist/manga/characters/${anilistId}?${query.toString()}`);
  }

  async addAniListMangaToMyLibrary(anilistId: number): Promise<ApiMessageResponse & { added: boolean }> {
    return this.post<ApiMessageResponse & { added: boolean }>(`/library/anilist/${anilistId}`);
  }

  async removeAniListMangaFromMyLibrary(anilistId: number): Promise<ApiMessageResponse & { removed: boolean }> {
    return this.delete<ApiMessageResponse & { removed: boolean }>(`/library/anilist/${anilistId}`);
  }

  async getAniListFavoriteIDs(): Promise<FavoriteIDsResponse<number>> {
    return this.get<FavoriteIDsResponse<number>>("/favorites/anilist");
  }

  async addAniListMangaToFavorites(anilistId: number): Promise<ApiMessageResponse & { added: boolean }> {
    return this.post<ApiMessageResponse & { added: boolean }>(`/favorites/anilist/${anilistId}`);
  }

  async removeAniListMangaFromFavorites(anilistId: number): Promise<ApiMessageResponse & { removed: boolean }> {
    return this.delete<ApiMessageResponse & { removed: boolean }>(`/favorites/anilist/${anilistId}`);
  }

  async getAniListCharacterFavorites(): Promise<AniListCharacterFavoritesResponse> {
    return this.get<AniListCharacterFavoritesResponse>('/favorites/anilist/characters');
  }

  async getAniListCharacterDetails(characterId: number): Promise<AniListCharacterDetailsResponse> {
    return this.get<AniListCharacterDetailsResponse>(`/anilist/character/details/${characterId}`);
  }

  async addAniListCharacterToFavorites(characterId: number, mangaId: number, role: string): Promise<ApiMessageResponse & { added: boolean }> {
    return this.post<ApiMessageResponse & { added: boolean }>(`/favorites/anilist/characters/${characterId}`, { mangaId, role });
  }

  async removeAniListCharacterFromFavorites(characterId: number): Promise<ApiMessageResponse & { removed: boolean }> {
    return this.delete<ApiMessageResponse & { removed: boolean }>(`/favorites/anilist/characters/${characterId}`);
  }

  async updateAniListMangaReadingStatus(anilistId: number, status: AniListReadingStatus): Promise<ApiMessageResponse & { readingStatus: AniListReadingStatus }> {
    return this.patch<ApiMessageResponse & { readingStatus: AniListReadingStatus }>(`/library/anilist/${anilistId}/status`, { status });
  }

  async updateAniListMangaPersonalRating(anilistId: number, personalRating: number | null): Promise<ApiMessageResponse & { personalRating: number | null }> {
    return this.patch<ApiMessageResponse & { personalRating: number | null }>(`/library/anilist/${anilistId}/rating`, { personalRating });
  }

  async getManga(mangaId: string): Promise<MangaDetailsResponse> {
    return this.get<MangaDetailsResponse>(`/manga/${encodeURIComponent(mangaId)}`);
  }

  async getMangaChapters(mangaId: string, params: MangaChapterParams = {}): Promise<MangaChaptersResponse> {
    const query = this.buildQuery({
      limit: params.limit,
      offset: params.offset,
      "translatedLanguage[]": params.translatedLanguage,
      language: params.language,
      scanlationGroup: params.scanlationGroup,
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

  async getChapterPage(chapterId: string, page: number, quality?: ChapterQuality): Promise<ChapterPageResponse> {
    return this.get<ChapterPageResponse>(
      `/chapter/${encodeURIComponent(chapterId)}/page${this.buildQuery({ page, quality })}`
    );
  }

  async getChapterPageCount(chapterId: string, quality?: ChapterQuality): Promise<ChapterPageCountResponse> {
    return this.get<ChapterPageCountResponse>(
      `/chapter/${encodeURIComponent(chapterId)}/page-count${this.buildQuery({ quality })}`
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

  async getMangaDexFavoriteIDs(): Promise<FavoriteIDsResponse<string>> {
    return this.get<FavoriteIDsResponse<string>>("/favorites/mangadex");
  }

  async getMangaDexFavorites(): Promise<MangaDexFavoritesResponse> {
    return this.get<MangaDexFavoritesResponse>("/favorites/mangadex");
  }

  async addMangaToFavorites(mangaId: string): Promise<ApiMessageResponse & { added: boolean }> {
    return this.post<ApiMessageResponse & { added: boolean }>(`/favorites/mangadex/${encodeURIComponent(mangaId)}`);
  }

  async removeMangaFromFavorites(mangaId: string): Promise<ApiMessageResponse & { removed: boolean }> {
    return this.delete<ApiMessageResponse & { removed: boolean }>(`/favorites/mangadex/${encodeURIComponent(mangaId)}`);
  }

  getImageUrl(path: string): string {
    return `${this.BASEURL}/images/${path}`;
  }

  getBaseUrl(): string {
    return this.BASEURL;
  }
}
