import React, { useEffect, useState } from 'react';
import { Navbar, Nav, Container, Form, FormControl, Button, NavDropdown, Offcanvas } from 'react-bootstrap';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import GuidedRecommendationChat from '../home/guided-recommendation-chat';

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
    const [darkMode, setDarkMode] = useState<boolean>(() => {
        const saved = localStorage.getItem('darkMode');
        return saved ? JSON.parse(saved) : true; // Default to dark mode
    });
    const [showRecommendations, setShowRecommendations] = useState(false);

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
        </Navbar>
    );
};

export default HorizontalNavbar;
