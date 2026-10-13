import React from 'react';
import { BookInfo } from '../../app/clients/backend/types';

interface BookDetailsProps {
    book: BookInfo;

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
