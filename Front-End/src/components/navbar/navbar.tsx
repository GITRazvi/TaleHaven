import React, { useEffect, useState } from 'react';
import { Navbar, Nav, Container, Form, FormControl, Button, NavDropdown } from 'react-bootstrap';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

interface NavbarProps {
    searchQuery: string;
    setSearchQuery: React.Dispatch<React.SetStateAction<string>>;
}

const HorizontalNavbar: React.FC<NavbarProps> = ({ searchQuery, setSearchQuery }) => {
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    const [darkMode, setDarkMode] = useState<boolean>(() => {
        const saved = localStorage.getItem('darkMode');
        return saved ? JSON.parse(saved) : true; // Default to dark mode
    });

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
                                title={`👤 ${user.email.split('@')[0]}`} 
                                id="user-dropdown"
                                align="end"
                            >
                                <NavDropdown.Item as={Link} to="/profile">
                                    My Profile
                                </NavDropdown.Item>
                                <NavDropdown.Item as={Link} to="/library">
                                    My Library
                                </NavDropdown.Item>
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
        </Navbar>
    );
};

export default HorizontalNavbar;
