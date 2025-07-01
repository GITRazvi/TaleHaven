import React from 'react';
import { Navbar, Nav, Container, Form, FormControl, Button } from 'react-bootstrap';
import { Link, useNavigate } from 'react-router-dom';

interface NavbarProps {
    searchQuery: string;
    setSearchQuery: React.Dispatch<React.SetStateAction<string>>;
    setSearchResults: React.Dispatch<React.SetStateAction<any[]>>;
}

const HorizontalNavbar: React.FC<NavbarProps> = ({ searchQuery, setSearchQuery, setSearchResults }) => {
    const navigate = useNavigate();

    const handleSearchSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSearchResults([]); // Clear previous results (optional).
        navigate('/search'); // Navigate to the search page.
    };

    return (
        <Navbar bg="dark" variant="dark" expand="lg">
            <Container>
                <Navbar.Brand as={Link} to="/">TaleHaven</Navbar.Brand>
                <Navbar.Toggle aria-controls="navbar-nav" />
                <Navbar.Collapse id="navbar-nav">
                    <Nav className="me-auto">
                        <Nav.Link as={Link} to="/">Home</Nav.Link>
                    </Nav>
                    <Form className="d-flex" onSubmit={handleSearchSubmit}>
                        <FormControl
                            type="search"
                            placeholder="Search"
                            className="me-2"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            style={{ width: '300px' }}
                        />
                        <Button variant="outline-light" type="submit">Search</Button>
                    </Form>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    );
};

export default HorizontalNavbar;