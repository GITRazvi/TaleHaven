import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useParams } from 'react-router-dom';
import HorizontalNavbar from './navbar/navbar';
import MangaSearchPage from './manga/manga-search-page';
import MangaDetailsPage from './manga/manga-details-page';
import MangaReaderPage from './manga/manga-reader-page';
import HomePage from './home/home-page';
import LoginPage from './auth/login-page';
import RegisterPage from './auth/register-page';
import ProfilePage from './auth/profile-page';
import ProtectedRoute from './auth/ProtectedRoute';
import LibraryPage from './library/library-page';
import { AuthProvider } from '../contexts/AuthContext';

// Wrapper to force remount when chapter changes
const MangaReaderWrapper: React.FC = () => {
    const { chapterId } = useParams<{ chapterId: string }>();
    return <MangaReaderPage key={chapterId} />;
};

const App: React.FC = () => {
    const [searchQuery, setSearchQuery] = useState<string>('');

    return (
        <Router>
            <AuthProvider>
                <HorizontalNavbar
                    searchQuery={searchQuery}
                    setSearchQuery={setSearchQuery}
                />
                <Routes>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/register" element={<RegisterPage />} />
                    <Route path="/profile" element={
                        <ProtectedRoute>
                            <ProfilePage />
                        </ProtectedRoute>
                    } />
                    <Route path="/library" element={
                        <ProtectedRoute>
                            <LibraryPage />
                        </ProtectedRoute>
                    } />
                    <Route path="/manga/search" element={<MangaSearchPage />} />
                    <Route path="/manga/:mangaId" element={<MangaDetailsPage />} />
                    <Route path="/manga/:mangaId/chapter/:chapterId" element={<MangaReaderWrapper />} />
                </Routes>
            </AuthProvider>
        </Router>
    );
};

export default App;
