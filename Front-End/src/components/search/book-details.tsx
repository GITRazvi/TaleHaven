import React from 'react';

interface BookDetailsProps {
    book: {
        Title: string;
        Authors: string[];
        Description: string;
        Categories: string[];
        Thumbnail: string;
        Rating: number;
        RatingCount: number;
        PublishedDate: string;
    };

}

const BookDetails: React.FC<BookDetailsProps> = ({ book }) => {
    return (
        <div style={{ padding: '20px' }}>

            <img src={book.Thumbnail} alt={book.Title} style={{ maxWidth: '200px' }} />
            <h1>{book.Title}</h1>
            <h2>Authors: {book.Authors?.join(', ')}</h2>
            <p>{book.Description}</p>
            <p>Categories: {book.Categories?.join(', ')}</p>
            <p>Rating: {book.Rating} ({book.RatingCount} reviews)</p>
            <p>Published Date: {book.PublishedDate}</p>
        </div>
    );
};

export default BookDetails;
