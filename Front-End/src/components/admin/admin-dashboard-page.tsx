import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Col, Container, Row, Spinner } from 'react-bootstrap';
import Backend from '../../app/clients/backend/backend';
import { AdminDashboardStats } from '../../app/clients/backend/types';

const dashboardCards: Array<{ key: keyof AdminDashboardStats; label: string; detail: string }> = [
    { key: 'users', label: 'Accounts', detail: 'Registered readers' },
    { key: 'aniListLibrary', label: 'AniList library entries', detail: 'Saved metadata titles' },
    { key: 'mangaDexLibrary', label: 'MangaDex library entries', detail: 'Saved readable titles' },
    { key: 'favorites', label: 'Favorites', detail: 'Manga and character favorites' },
    { key: 'trackedChapters', label: 'Tracked chapters', detail: 'Saved chapter-progress records' },
];

const AdminDashboardPage: React.FC = () => {
    const backend = useMemo(() => new Backend(), []);
    const [stats, setStats] = useState<AdminDashboardStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const loadDashboard = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await backend.getAdminDashboard();
            setStats(response.stats);
        } catch (requestError) {
            const message = requestError instanceof Error ? requestError.message : 'Unable to load dashboard statistics.';
            setError(message);
        } finally {
            setLoading(false);
        }
    }, [backend]);

    useEffect(() => {
        void loadDashboard();
    }, [loadDashboard]);

    return (
        <Container className="py-4 py-md-5 admin-dashboard">
            <div className="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-4">
                <div>
                    <p className="text-uppercase small text-muted fw-semibold mb-1">Administration</p>
                    <h1 className="h2 mb-2">TaleHaven dashboard</h1>
                    <p className="text-muted mb-0">A private overview of current site activity. Administrative actions will be added separately.</p>
                </div>
                <Button variant="outline-primary" onClick={() => void loadDashboard()} disabled={loading}>
                    {loading ? 'Refreshing…' : 'Refresh'}
                </Button>
            </div>

            {error && (
                <Alert variant="danger" className="d-flex flex-wrap align-items-center justify-content-between gap-2">
                    <span>Dashboard statistics could not be loaded. {error}</span>
                    <Button variant="outline-danger" size="sm" onClick={() => void loadDashboard()}>Try again</Button>
                </Alert>
            )}

            {loading && !stats ? (
                <div className="d-flex align-items-center gap-2 py-5" role="status">
                    <Spinner animation="border" size="sm" />
                    <span>Loading dashboard statistics…</span>
                </div>
            ) : stats ? (
                <Row className="g-3">
                    {dashboardCards.map((card) => (
                        <Col key={card.key} sm={6} lg={4} xl>
                            <Card className="admin-stat-card h-100 shadow-sm">
                                <Card.Body>
                                    <p className="text-muted small text-uppercase fw-semibold mb-2">{card.label}</p>
                                    <div className="admin-stat-value">{stats[card.key].toLocaleString()}</div>
                                    <p className="text-muted small mb-0 mt-2">{card.detail}</p>
                                </Card.Body>
                            </Card>
                        </Col>
                    ))}
                </Row>
            ) : null}

            <Card className="mt-4 admin-dashboard-next shadow-sm">
                <Card.Body>
                    <h2 className="h5">Next administrative tools</h2>
                    <p className="mb-0 text-muted">This foundation is intentionally read-only. The next additions can be moderation queues, report review, chat-room management, and audited administrator actions.</p>
                </Card.Body>
            </Card>
        </Container>
    );
};

export default AdminDashboardPage;
