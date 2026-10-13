import React, { useEffect } from 'react';
import { Row, Col } from 'react-bootstrap';
import SearchResult from './search-result';
import Backend from '../../app/clients/backend/backend';
import { BookInfo } from '../../app/clients/backend/types';

interface SearchProps {
    query: string;
    results: BookInfo[];
    setResults: React.Dispatch<React.SetStateAction<BookInfo[]>>;
    onSelectBook: (book: BookInfo) => void;
}

const Search: React.FC<SearchProps> = ({ query, results, setResults, onSelectBook }) => {
    useEffect(() => {
        const fetchResults = async () => {
            if (query && results.length === 0) {
                const response = await new Backend().searchBooks(query);
                setResults(response.booksList ?? []);
            }
        };
        fetchResults();
    }, [query, results, setResults]);

    return (
        <Row className="p-3">
            {results.map((book) => (
                <Col key={book.Id} sm={12} md={6} lg={4} className="mb-4">
                    <SearchResult result={book} onSelectBook={onSelectBook} />
                </Col>
            ))}
        </Row>
    );
};

export default Search;
