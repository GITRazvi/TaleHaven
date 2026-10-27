import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Container, Spinner } from 'react-bootstrap';

/**
 * Protected Route Component - server-side cookie sessions
 *
 * Checks the safe user state restored from an HttpOnly session cookie.
 * Redirects to login if token is missing, invalid, or expired
 */

interface ProtectedRouteProps {
    children: React.ReactElement;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
    const { user, loading, sessionExpired } = useAuth();
    const location = useLocation();

    if (loading) {
        return (
            <Container className="d-flex justify-content-center align-items-center" style={{ minHeight: '80vh' }}>
                <Spinner animation="border" role="status">
                    <span className="visually-hidden">Loading...</span>
                </Spinner>
            </Container>
        );
    }

    if (!user) {
        return <Navigate to="/login" state={{ from: location, sessionExpired }} replace />;
    }

    return children;
};

export default ProtectedRoute;


