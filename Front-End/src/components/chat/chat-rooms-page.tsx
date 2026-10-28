import React, { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Badge, Button, Card, Col, Container, Form, Row, Spinner } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { ChatRoom, ChatRoomVisibility } from '../../app/clients/backend/types';
import { presentationThemeStyle } from '../../app/theme-style';

const ChatRoomsPage: React.FC = () => {
    const backend = useMemo(() => new Backend(), []);
    const navigate = useNavigate();
    const [rooms, setRooms] = useState<ChatRoom[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [creating, setCreating] = useState(false);
    const [joiningRoomID, setJoiningRoomID] = useState<number | null>(null);
    const [name, setName] = useState('');
    const [visibility, setVisibility] = useState<ChatRoomVisibility>('public');
    const [showCreateForm, setShowCreateForm] = useState(false);

    const loadRooms = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await backend.getChatRooms();
            setRooms(response.data);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to load chat rooms.');
        } finally {
            setLoading(false);
        }
    }, [backend]);

    useEffect(() => { void loadRooms(); }, [loadRooms]);

    const createRoom = async (event: FormEvent) => {
        event.preventDefault();
        if (!name.trim()) return;
        setCreating(true);
        setError(null);
        try {
            const response = await backend.createChatRoom({ name: name.trim(), visibility });
            navigate(`/chat/rooms/${response.room.id}`);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to create this room.');
        } finally {
            setCreating(false);
        }
    };

    const joinRoom = async (room: ChatRoom) => {
        if (room.isMember) {
            navigate(`/chat/rooms/${room.id}`);
            return;
        }
        setJoiningRoomID(room.id);
        setError(null);
        try {
            await backend.joinPublicChatRoom(room.id);
            navigate(`/chat/rooms/${room.id}`);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to join this room.');
        } finally {
            setJoiningRoomID(null);
        }
    };

    const publicRooms = rooms.filter((room) => room.visibility === 'public');
    const privateRooms = rooms.filter((room) => room.visibility === 'invite_only');

    const renderRoom = (room: ChatRoom) => (
        <Col key={room.id} md={6} xl={4}>
            <Card className={`chat-room-card chat-room-theme-${room.theme || 'midnight'}${room.themeImagePath ? ' theme-has-image' : ''} h-100 shadow-sm`} style={presentationThemeStyle(room.customThemeColor, room.themeImagePath ? `${backend.getBaseUrl()}${room.themeImagePath}` : '')}>
                <Card.Body className="d-flex flex-column">
                    <div className="d-flex align-items-start justify-content-between gap-2">
                        <div className="d-flex align-items-start gap-3">
                            <div className="chat-room-cover chat-room-cover-placeholder">{room.imagePath ? <img src={`${backend.getBaseUrl()}${room.imagePath}`} alt={`Cover for ${room.name}`} /> : '💬'}</div>
                            <div><h2 className="h5 mb-1">{room.name}</h2><p className="small text-muted mb-0">Created by {room.ownerDisplayName}</p></div>
                        </div>
                        <Badge bg={room.visibility === 'public' ? 'success' : 'secondary'}>{room.visibility === 'public' ? 'Public' : 'Private'}</Badge>
                    </div>
                    <p className="small text-muted mt-3 mb-3">{room.memberCount} {room.memberCount === 1 ? 'member' : 'members'}</p>
                    <Button className="mt-auto" variant={room.isMember ? 'outline-primary' : 'primary'} disabled={joiningRoomID === room.id} onClick={() => void joinRoom(room)}>
                        {joiningRoomID === room.id ? 'Joining…' : room.isMember ? 'Open room' : 'Join room'}
                    </Button>
                </Card.Body>
            </Card>
        </Col>
    );

    return (
        <Container className="py-4 py-md-5 chat-rooms-page">
            <div className="mb-4">
                <p className="text-uppercase small text-muted fw-semibold mb-1">Community</p>
                <div className="d-flex flex-wrap align-items-center justify-content-between gap-3"><h1 className="h2 mb-0">Chatrooms</h1><Button onClick={() => setShowCreateForm((current) => !current)} aria-expanded={showCreateForm} aria-controls="create-chatroom-panel">{showCreateForm ? 'Close creator' : 'Create chatroom'}</Button></div>
                <p className="text-muted mb-0">Join a public conversation or create a private room for people you invite. Private rooms never appear in public browsing.</p>
            </div>

            {error && <Alert variant="danger">{error}</Alert>}

            {showCreateForm && <Card id="create-chatroom-panel" className="chat-create-card shadow-sm mb-4">
                <Card.Body>
                    <div className="d-flex justify-content-between align-items-center gap-3"><h2 className="h5 mb-0">Create a chatroom</h2><Button variant="outline-secondary" size="sm" onClick={() => setShowCreateForm(false)} disabled={creating}>Cancel</Button></div>
                    <Form onSubmit={createRoom} className="row g-3 align-items-end">
                        <Form.Group as={Col} md={7} controlId="chat-room-name">
                            <Form.Label>Room name</Form.Label>
                            <Form.Control value={name} maxLength={80} minLength={2} required placeholder="For example, Fantasy manga readers" onChange={(event) => setName(event.target.value)} />
                        </Form.Group>
                        <Form.Group as={Col} md={3} controlId="chat-room-visibility">
                            <Form.Label>Access</Form.Label>
                            <Form.Select value={visibility} onChange={(event) => setVisibility(event.target.value as ChatRoomVisibility)}>
                                <option value="public">Public</option>
                                <option value="invite_only">Private</option>
                            </Form.Select>
                        </Form.Group>
                        <Col md={2}><Button type="submit" className="w-100" disabled={creating}>{creating ? 'Creating…' : 'Create'}</Button></Col>
                    </Form>
                </Card.Body>
            </Card>}

            {loading ? <div className="py-5 d-flex gap-2 align-items-center" role="status"><Spinner animation="border" size="sm" /> Loading chatrooms…</div> : <>
                <section aria-labelledby="public-rooms-heading">
                    <div className="d-flex align-items-center justify-content-between mb-3"><h2 id="public-rooms-heading" className="h4 mb-0">Public rooms</h2><Button variant="outline-secondary" size="sm" onClick={() => void loadRooms()}>Refresh</Button></div>
                    {publicRooms.length ? <Row className="g-3">{publicRooms.map(renderRoom)}</Row> : <p className="text-muted">There are no public rooms yet. Create the first one.</p>}
                </section>
                {privateRooms.length > 0 && <section className="mt-5" aria-labelledby="your-private-rooms-heading"><h2 id="your-private-rooms-heading" className="h4">Your private rooms</h2><p className="text-muted small">Only rooms you created or joined are shown here.</p><Row className="g-3">{privateRooms.map(renderRoom)}</Row></section>}
            </>}
        </Container>
    );
};

export default ChatRoomsPage;
