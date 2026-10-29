import React, { useEffect, useMemo, useState } from 'react';
import { Navbar, Nav, Container, Form, FormControl, Button, NavDropdown, Offcanvas } from 'react-bootstrap';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import GuidedRecommendationChat from '../home/guided-recommendation-chat';
import Backend from '../../app/clients/backend/backend';
import { NotificationItem } from '../../app/clients/backend/types';

interface NavbarProps {
    searchQuery: string;
    setSearchQuery: React.Dispatch<React.SetStateAction<string>>;
    siteTheme: string;
    siteThemeStyle?: React.CSSProperties;
    hasSiteThemeImage: boolean;
}

const HorizontalNavbar: React.FC<NavbarProps> = ({
    searchQuery,
    setSearchQuery,
    siteTheme,
    siteThemeStyle,
    hasSiteThemeImage,
}) => {
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    const backend = useMemo(() => new Backend(), []);
    const [darkMode, setDarkMode] = useState<boolean>(() => {
        const saved = localStorage.getItem('darkMode');
        return saved ? JSON.parse(saved) : true; // Default to dark mode
    });
    const [showRecommendations, setShowRecommendations] = useState(false);
    const [showNotifications, setShowNotifications] = useState(false);
    const [notifications, setNotifications] = useState<NotificationItem[]>([]);
    const [unreadNotifications, setUnreadNotifications] = useState(0);
    const [notificationsLoading, setNotificationsLoading] = useState(false);

    useEffect(() => {
        localStorage.setItem('darkMode', JSON.stringify(darkMode));
        if (darkMode) {
            document.body.classList.add('dark-mode');
            document.body.style.backgroundColor = '#1a1a1a';
            document.body.style.color = '#e0e0e0';
        } else {
            document.body.classList.remove('dark-mode');
            document.body.style.backgroundColor = '#ffffff';
            document.body.style.color = '#000000';
        }
    }, [darkMode]);

    useEffect(() => {
        if (!user) { setNotifications([]); setUnreadNotifications(0); return undefined; }
        const loadUnread = () => {
            const request = backend.getNotificationUnreadCount?.();
            if (request) void request.then((response) => setUnreadNotifications(response.unread)).catch(() => undefined);
        };
        void loadUnread();
        const timer = window.setInterval(loadUnread, 30000);
        return () => window.clearInterval(timer);
    }, [backend, user]);

    useEffect(() => {
        if (!showNotifications || !user) return;
        setNotificationsLoading(true);
        const request = backend.getNotifications?.();
        if (!request) { setNotificationsLoading(false); return; }
        void request.then((response) => { setNotifications(response.data); setUnreadNotifications(response.unread); }).catch(() => undefined).finally(() => setNotificationsLoading(false));
    }, [backend, showNotifications, user]);

    const openNotification = async (item: NotificationItem) => {
        if (!item.read) { try { await backend.markNotificationRead(item.id); setNotifications((current) => current.map((entry) => entry.id === item.id ? { ...entry, read: true } : entry)); setUnreadNotifications((current) => Math.max(0, current - 1)); } catch { /* navigation remains available */ } }
        setShowNotifications(false);
        if (item.targetPath.startsWith('/')) navigate(item.targetPath);
    };

    const markAllNotificationsRead = async () => {
        try { await backend.markAllNotificationsRead(); setNotifications((current) => current.map((item) => ({ ...item, read: true }))); setUnreadNotifications(0); } catch { /* retain current list on a transient failure */ }
    };

    const handleSearchSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (searchQuery.trim()) {
            navigate(`/manga/search?title=${encodeURIComponent(searchQuery.trim())}`);
        }
    };

    const toggleDarkMode = () => {
        setDarkMode(!darkMode);
    };

    const handleLogout = async () => {
        try {
            await logout();
            navigate('/');
        } catch (error) {
            console.error('Logout failed:', error);
        }
    };

    return (
        <Navbar bg="dark" variant="dark" expand="lg">
            <Container>
                <Navbar.Brand as={Link} to="/">TaleHaven</Navbar.Brand>
                <Navbar.Toggle aria-controls="navbar-nav" />
                <Navbar.Collapse id="navbar-nav">
                    <Nav className="me-auto">
                        <Nav.Link as={Link} to="/">Home</Nav.Link>
                        <Nav.Link as={Link} to="/manga/search">Manga</Nav.Link>
                        <Nav.Link as="button" className="nav-recommendations-link" onClick={() => setShowRecommendations(true)}>Chatbot</Nav.Link>
                        {user && <Nav.Link as={Link} to="/chat">Chatrooms</Nav.Link>}
                    </Nav>
                    <Form className="d-flex align-items-center" onSubmit={handleSearchSubmit}>
                        <FormControl
                            type="search"
                            placeholder="Search manga..."
                            className="me-2"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            style={{ width: '300px' }}
                        />
                        <Button variant="outline-light" type="submit" className="me-2">Search</Button>
                        <Button 
                            variant={darkMode ? "outline-light" : "outline-dark"}
                            onClick={toggleDarkMode}
                            title={darkMode ? "Switch to Light Mode" : "Switch to Dark Mode"}
                            style={{ minWidth: '100px' }}
                            className="me-2"
                        >
                            {darkMode ? '☀️ Light' : '🌙 Dark'}
                        </Button>
                        {user ? (
                            <Button variant="outline-light" className="me-2 position-relative" onClick={() => setShowNotifications(true)} aria-label="Notifications">
                                🔔{unreadNotifications > 0 && <span className="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">{unreadNotifications > 99 ? '99+' : unreadNotifications}</span>}
                            </Button>
                        ) : null}
                        {user ? (
                            <NavDropdown 
                                title={`👤 ${user.displayName || user.email.split('@')[0]}`}
                                id="user-dropdown"
                                align="end"
                            >
                                <NavDropdown.Item as={Link} to="/profile">
                                    Settings
                                </NavDropdown.Item>
                                <NavDropdown.Item as={Link} to="/library">
                                    My Profile
                                </NavDropdown.Item>
                                {user.role === 'admin' && <NavDropdown.Item as={Link} to="/admin">
                                    Administration
                                </NavDropdown.Item>}
                                <NavDropdown.Divider />
                                <NavDropdown.Item onClick={handleLogout}>
                                    Logout
                                </NavDropdown.Item>
                            </NavDropdown>
                        ) : (
                            <>
                                <Button 
                                    variant="outline-light" 
                                    onClick={() => navigate('/login')}
                                    className="me-2"
                                >
                                    Login
                                </Button>
                                <Button 
                                    variant="light" 
                                    onClick={() => navigate('/register')}
                                >
                                    Register
                                </Button>
                            </>
                        )}
                    </Form>
                </Navbar.Collapse>
            </Container>
            <Offcanvas
                show={showRecommendations}
                onHide={() => setShowRecommendations(false)}
                placement="end"
                className={`recommendation-offcanvas chatbot-theme site-theme-${siteTheme}${hasSiteThemeImage ? ' theme-has-image' : ''}`}
                style={siteThemeStyle}
            >
                <Offcanvas.Header closeButton>
                    <Offcanvas.Title>Chatbot</Offcanvas.Title>
                </Offcanvas.Header>
                <Offcanvas.Body>
                    <GuidedRecommendationChat />
                </Offcanvas.Body>
            </Offcanvas>
            <Offcanvas show={showNotifications} onHide={() => setShowNotifications(false)} placement="end" className={`recommendation-offcanvas site-theme-${siteTheme}${hasSiteThemeImage ? ' theme-has-image' : ''}`} style={siteThemeStyle}>
                <Offcanvas.Header closeButton><Offcanvas.Title>Notifications</Offcanvas.Title></Offcanvas.Header>
                <Offcanvas.Body>{unreadNotifications > 0 && <Button variant="link" size="sm" className="p-0 mb-3" onClick={() => void markAllNotificationsRead()}>Mark all as read</Button>}{notificationsLoading ? <p className="text-muted">Loading notifications…</p> : notifications.length === 0 ? <p className="text-muted">You’re all caught up.</p> : <div className="d-grid gap-2">{notifications.map((item) => <button type="button" key={item.id} className={`text-start border rounded p-3 notification-item ${item.read ? 'is-read' : ''}`} onClick={() => void openNotification(item)}><strong>{item.title}</strong>{item.body && <span className="d-block small mt-1">{item.body}</span>}<small className="d-block text-muted mt-1">{new Date(item.createdAt).toLocaleString()}</small></button>)}</div>}</Offcanvas.Body>
            </Offcanvas>
        </Navbar>
    );
};

export default HorizontalNavbar;
