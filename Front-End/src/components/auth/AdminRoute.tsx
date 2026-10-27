import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Alert, Container, Spinner } from 'react-bootstrap';

interface AdminRouteProps {
    children: React.ReactElement;
}

// AdminRoute only controls frontend navigation. The server independently
// requires the admin role for every /admin API route.
const AdminRoute: React.FC<AdminRouteProps> = ({ children }) => {
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

    if (user.role !== 'admin') {
        return (
            <Container className="py-5" style={{ maxWidth: '44rem' }}>
                <Alert variant="warning" role="alert">
                    Administrator access is required to open this page.
                </Alert>
            </Container>
        );
    }

    return children;
};

export default AdminRoute;
