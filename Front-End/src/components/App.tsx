import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HorizontalNavbar from './navbar/navbar';
import Search from './search/search';
import BookDetails from './search/book-details';

const App: React.FC = () => {
    const [searchQuery, setSearchQuery] = useState<string>('');
    const [searchResults, setSearchResults] = useState<any[]>([]);
    const [selectedBook, setSelectedBook] = useState<any | null>(null);

    return (
        <Router>
            <HorizontalNavbar
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                setSearchResults={setSearchResults}
            />
            <Routes>
                <Route
                    path="/search"
                    element={
                        <Search
                            query={searchQuery}
                            results={searchResults}
                            setResults={setSearchResults}
                            onSelectBook={(book) => setSelectedBook(book)}
                        />
                    }
                />
                <Route
                    path="/book/:id"
                    element={
                        selectedBook && (
                            <BookDetails
                                book={selectedBook}
                            />
                        )
                    }
                />
            </Routes>
        </Router>
    );
};

export default App;
