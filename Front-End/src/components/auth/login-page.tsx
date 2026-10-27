import React, { useState } from 'react';
import { Container, Form, Button, Card, Alert } from 'react-bootstrap';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const LoginPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();

    const from = (location.state as any)?.from?.pathname || '/';
    const sessionExpired = Boolean((location.state as any)?.sessionExpired);
    const notice = (location.state as any)?.notice as string | undefined;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!email || !password) {
            setError('Please fill in all fields');
            return;
        }

        try {
            setError('');
            setLoading(true);
            await login(email, password);
            navigate(from, { replace: true });
        } catch (err: any) {
            const errorMessage = err?.details?.message || err?.message || 'Failed to login. Please check your credentials.';
            setError(errorMessage);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Container className="d-flex align-items-center justify-content-center" style={{ minHeight: '80vh' }}>
            <div style={{ width: '100%', maxWidth: '400px' }}>
                <Card>
                    <Card.Body>
                        <h2 className="text-center mb-4">Login to TaleHaven</h2>
                        {notice && <Alert variant="info">{notice}</Alert>}
                        {sessionExpired && <Alert variant="warning">Your session expired. Please log in again.</Alert>}
                        {error && <Alert variant="danger">{error}</Alert>}
                        <Form onSubmit={handleSubmit}>
                            <Form.Group className="mb-3" controlId="email">
                                <Form.Label>Email</Form.Label>
                                <Form.Control
                                    type="email"
                                    placeholder="Enter email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    required
                                />
                            </Form.Group>

                            <Form.Group className="mb-3" controlId="password">
                                <Form.Label>Password</Form.Label>
                                <Form.Control
                                    type="password"
                                    placeholder="Enter password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                />
                            </Form.Group>

                            <Button
                                className="w-100"
                                type="submit"
                                disabled={loading}
                                variant="primary"
                            >
                                {loading ? 'Logging in...' : 'Login'}
                            </Button>
                        </Form>
                    </Card.Body>
                </Card>
                <div className="text-center mt-3">
                    Don't have an account? <Link to="/register">Register here</Link>
                </div>
            </div>
        </Container>
    );
};

export default LoginPage;

