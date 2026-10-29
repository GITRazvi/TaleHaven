import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation, useParams } from 'react-router-dom';
import AniListMangaDetailsPage from './manga/anilist-manga-details-page';
import AniListGenrePage from './manga/anilist-genre-page';
import HorizontalNavbar from './navbar/navbar';
import MangaSearchPage from './manga/manga-search-page';
import MangaDetailsPage from './manga/manga-details-page';
import MangaReaderPage from './manga/manga-reader-page';
import HomePage from './home/home-page';
import LoginPage from './auth/login-page';
import RegisterPage from './auth/register-page';
import ProfilePage from './auth/profile-page';
import PublicProfilePage from './auth/public-profile-page';
import ProtectedRoute from './auth/ProtectedRoute';
import AdminRoute from './auth/AdminRoute';
import LibraryPage from './library/library-page';
import AdminDashboardPage from './admin/admin-dashboard-page';
import ChatRoomsPage from './chat/chat-rooms-page';
import ChatRoomPage from './chat/chat-room-page';
import ChatInvitePage from './chat/chat-invite-page';
import { AuthProvider, useAuth } from '../contexts/AuthContext';
import { presentationThemeStyle, themeBackgroundStyle } from '../app/theme-style';

// Wrapper to force remount when chapter changes
const MangaReaderWrapper: React.FC = () => {
    const { chapterId } = useParams<{ chapterId: string }>();
    return <MangaReaderPage key={chapterId} />;
};

const siteBackdropColors = {
    midnight: '#111a2d',
    forest: '#10241f',
    plum: '#24162d',
    ember: '#2a1b18',
    parchment: '#f2ead9',
};

const AppContent: React.FC = () => {
    const [searchQuery, setSearchQuery] = useState<string>('');
    const { user } = useAuth();
    const userID = user?.id ?? null;
    const location = useLocation();
    const siteTheme = user?.siteTheme || 'midnight';
    const siteCustomColor = user?.siteCustomColor || '';
    const siteThemeImagePaths = user?.siteThemeImagePaths || [];
    const siteThemeImageRotationMode = user?.siteThemeImageRotationMode || 'static';
    const siteThemeImageRotationMinutes = user?.siteThemeImageRotationMinutes || 15;
    const siteThemeImageKey = siteThemeImagePaths.join('|');
    const siteThemeImageCount = siteThemeImagePaths.length;
    const [siteThemeImageIndex, setSiteThemeImageIndex] = useState(0);
    const currentSiteThemeImagePath = siteThemeImagePaths[siteThemeImageIndex] || siteThemeImagePaths[0] || '';
    const currentSiteThemeImageUrl = currentSiteThemeImagePath
        ? `${process.env.REACT_APP_API_BASE_URL ?? 'http://localhost:3001'}${currentSiteThemeImagePath}`
        : '';
    const siteThemeStyle = presentationThemeStyle(siteCustomColor, currentSiteThemeImageUrl);
    const siteThemePageColor = siteThemeStyle?.['--theme-page'];
    const usesScopedTheme = location.pathname === '/library' || location.pathname.startsWith('/chat/rooms/');
    const isReaderPage = /^\/manga\/[^/]+\/chapter\/[^/]+/.test(location.pathname);

    useEffect(() => {
        document.body.classList.toggle('reader-page-active', isReaderPage);
        return () => document.body.classList.remove('reader-page-active');
    }, [isReaderPage]);

    useEffect(() => {
        if (!siteThemeImageCount) {
            setSiteThemeImageIndex(0);
            return;
        }
        if (siteThemeImageRotationMode === 'refresh') {
            setSiteThemeImageIndex(Math.floor(Math.random() * siteThemeImageCount));
            return;
        }
        if (siteThemeImageRotationMode === 'session') {
            const sessionKey = `talehaven-site-theme-image-${userID ?? 'anonymous'}-${siteThemeImageKey}`;
            const storedValue = window.sessionStorage.getItem(sessionKey);
            const storedIndex = storedValue === null ? Number.NaN : Number(storedValue);
            const imageIndex = Number.isInteger(storedIndex) && storedIndex >= 0 && storedIndex < siteThemeImageCount
                ? storedIndex
                : Math.floor(Math.random() * siteThemeImageCount);
            window.sessionStorage.setItem(sessionKey, String(imageIndex));
            setSiteThemeImageIndex(imageIndex);
            return;
        }
        setSiteThemeImageIndex(0);
    }, [siteThemeImageKey, siteThemeImageCount, siteThemeImageRotationMode, userID]);

    useEffect(() => {
        if (siteThemeImageCount < 2 || siteThemeImageRotationMode !== 'interval') return undefined;
        const interval = window.setInterval(() => {
            setSiteThemeImageIndex((current) => (current + 1) % siteThemeImageCount);
        }, siteThemeImageRotationMinutes * 60 * 1000);
        return () => window.clearInterval(interval);
    }, [siteThemeImageKey, siteThemeImageCount, siteThemeImageRotationMode, siteThemeImageRotationMinutes]);

    useEffect(() => {
        if (usesScopedTheme) return undefined;

        const body = document.body;
        const root = document.documentElement;
        const previousBodyBackground = body.style.background;
        const previousBodyBackgroundImage = body.style.backgroundImage;
        const previousBodyBackgroundSize = body.style.backgroundSize;
        const previousBodyBackgroundPosition = body.style.backgroundPosition;
        const previousBodyBackgroundAttachment = body.style.backgroundAttachment;
        const previousRootBackground = root.style.background;
        const background = siteThemePageColor || siteBackdropColors[siteTheme];
        const backgroundImage = themeBackgroundStyle(currentSiteThemeImageUrl)?.backgroundImage;

        body.classList.add('site-theme-active');
        body.style.background = background;
        root.style.background = background;
        if (backgroundImage) {
            body.style.backgroundImage = backgroundImage;
            body.style.backgroundSize = 'cover';
            body.style.backgroundPosition = 'center';
            body.style.backgroundAttachment = 'fixed';
        }

        return () => {
            body.classList.remove('site-theme-active');
            body.style.background = previousBodyBackground;
            body.style.backgroundImage = previousBodyBackgroundImage;
            body.style.backgroundSize = previousBodyBackgroundSize;
            body.style.backgroundPosition = previousBodyBackgroundPosition;
            body.style.backgroundAttachment = previousBodyBackgroundAttachment;
            root.style.background = previousRootBackground;
        };
    }, [siteTheme, siteThemePageColor, currentSiteThemeImageUrl, usesScopedTheme]);

    return (
        <>
            <HorizontalNavbar
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                siteTheme={siteTheme}
                siteThemeStyle={siteThemeStyle}
                hasSiteThemeImage={Boolean(currentSiteThemeImageUrl)}
            />
            <div className={`site-theme-app site-theme-${siteTheme}${currentSiteThemeImageUrl ? ' theme-has-image' : ''}${isReaderPage ? ' manga-reader-route' : ''}`} style={siteThemeStyle}>
                <Routes>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/register" element={<RegisterPage />} />
                    <Route path="/profile" element={
                        <ProtectedRoute>
                            <ProfilePage />
                        </ProtectedRoute>
                    } />
                    <Route path="/profile/:userID" element={
                        <ProtectedRoute>
                            <PublicProfilePage />
                        </ProtectedRoute>
                    } />
                    <Route path="/library" element={
                        <ProtectedRoute>
                            <LibraryPage />
                        </ProtectedRoute>
                    } />
                    <Route path="/admin" element={
                        <AdminRoute>
                            <AdminDashboardPage />
                        </AdminRoute>
                    } />
                    <Route path="/chat" element={
                        <ProtectedRoute>
                            <ChatRoomsPage />
                        </ProtectedRoute>
                    } />
                    <Route path="/chat/rooms/:roomID" element={
                        <ProtectedRoute>
                            <ChatRoomPage />
                        </ProtectedRoute>
                    } />
                    <Route path="/chat/invite/:token" element={
                        <ProtectedRoute>
                            <ChatInvitePage />
                        </ProtectedRoute>
                    } />
                    <Route path="/manga/search" element={<MangaSearchPage />} />
                    <Route path="/anilist/manga/:anilistId" element={<AniListMangaDetailsPage />} />
                    <Route path="/anilist/genre/:genre" element={<AniListGenrePage />} />
                    <Route path="/manga/:mangaId" element={<MangaDetailsPage />} />
                    <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderWrapper />} />
                </Routes>
            </div>
        </>
    );
};

const App: React.FC = () => {
    return (
        <Router>
            <AuthProvider>
                <AppContent />
            </AuthProvider>
        </Router>
    );
};

export default App;
