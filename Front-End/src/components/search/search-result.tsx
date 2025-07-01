import React from 'react';
import { useNavigate } from 'react-router-dom';

type BookInfo = {
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

interface SearchResultProps {
    result: BookInfo;
    onSelectBook: (book: BookInfo) => void;
}

const SearchResult: React.FC<SearchResultProps> = ({ result, onSelectBook }) => {
    const navigate = useNavigate();

    const handleBookClick = () => {
        onSelectBook(result); // Set the selected book in state
        navigate(`/book/${result.Id}`); // Navigate to the book details page
    };

    return (
        <div
            style={{
                border: '1px solid #ccc',
                padding: '10px',
                margin: '10px',
                borderRadius: '5px',
                cursor: 'pointer',
            }}
            onClick={handleBookClick}
        >
            <div style={{ display: 'flex', alignItems: 'center' }}>
                {result.Thumbnail && (
                    <img
                        src={result.Thumbnail}
                        alt={`${result.Title} Thumbnail`}
                        style={{ width: '50px', height: '75px', marginRight: '15px' }}
                    />
                )}
                <div>
                    <h5 style={{ margin: '0 0 5px' }}>{result.Title}</h5>
                    <p style={{ margin: '0', color: '#555' }}>
                        {result.Authors?.join(', ') || 'Unknown Author'}
                    </p>
                </div>
            </div>
        </div>
    );
};

export default SearchResult;
