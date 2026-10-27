import React, { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';
import { Alert, Badge, Button, Card, Container, Form, Spinner } from 'react-bootstrap';
import { Link, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { ChatMessage, ChatRoom, DisplayTheme } from '../../app/clients/backend/types';
import { presentationThemeStyle } from '../../app/theme-style';

const REFRESH_INTERVAL_MS = 5000;

const chatThemeOptions: Array<{ value: DisplayTheme; label: string }> = [
    { value: 'midnight', label: 'Midnight' },
    { value: 'forest', label: 'Forest' },
    { value: 'plum', label: 'Plum' },
    { value: 'ember', label: 'Ember' },
    { value: 'parchment', label: 'Parchment' },
];

const ChatRoomPage: React.FC = () => {
    const backend = useMemo(() => new Backend(), []);
    const { roomID } = useParams<{ roomID: string }>();
    const parsedRoomID = Number(roomID);
    const [room, setRoom] = useState<ChatRoom | null>(null);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [body, setBody] = useState('');
    const [messageImage, setMessageImage] = useState<File | null>(null);
    const [messageImageInputKey, setMessageImageInputKey] = useState(0);
    const [sending, setSending] = useState(false);
    const [inviteLink, setInviteLink] = useState<string | null>(null);
    const [creatingInvite, setCreatingInvite] = useState(false);
    const [changingVisibility, setChangingVisibility] = useState(false);
    const [roomImage, setRoomImage] = useState<File | null>(null);
    const [roomImageInputKey, setRoomImageInputKey] = useState(0);
    const [uploadingRoomImage, setUploadingRoomImage] = useState(false);
    const [themeDraft, setThemeDraft] = useState<DisplayTheme>('midnight');
    const [savingTheme, setSavingTheme] = useState(false);
    const [customThemeColor, setCustomThemeColor] = useState('#7ab4ff');
    const [useCustomThemeColor, setUseCustomThemeColor] = useState(false);
    const [themeImage, setThemeImage] = useState<File | null>(null);
    const [themeImageInputKey, setThemeImageInputKey] = useState(0);

    const imageUrl = (imagePath: string) => imagePath ? `${backend.getBaseUrl()}${imagePath}` : '';

    const loadMessages = useCallback(async () => {
        const response = await backend.getChatMessages(parsedRoomID);
        setMessages(response.data);
    }, [backend, parsedRoomID]);

    const loadRoom = useCallback(async (initial = false) => {
        if (!Number.isInteger(parsedRoomID) || parsedRoomID < 1) {
            setError('This chatroom address is invalid.');
            setLoading(false);
            return;
        }
        if (initial) setLoading(true);
        try {
            const [roomResponse, messagesResponse] = await Promise.all([backend.getChatRoom(parsedRoomID), backend.getChatMessages(parsedRoomID)]);
            setRoom(roomResponse.room);
            setMessages(messagesResponse.data);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to open this chatroom.');
        } finally {
            if (initial) setLoading(false);
        }
    }, [backend, parsedRoomID]);

    useEffect(() => {
        void loadRoom(true);
    }, [loadRoom]);

    useEffect(() => {
        if (!room) return undefined;
        const interval = window.setInterval(() => { void loadMessages().catch(() => undefined); }, REFRESH_INTERVAL_MS);
        return () => window.clearInterval(interval);
    }, [loadMessages, room]);

    useEffect(() => {
        if (!room) return;
        setThemeDraft(room.theme || 'midnight');
        setCustomThemeColor(room.customThemeColor || '#7ab4ff');
        setUseCustomThemeColor(Boolean(room.customThemeColor));
    }, [room]);

    const sendMessage = async (event: FormEvent) => {
        event.preventDefault();
        const value = body.trim();
        if (!value && !messageImage) return;
        setSending(true);
        try {
            const response = messageImage
                ? await backend.sendChatMessageWithImage(parsedRoomID, value, messageImage)
                : await backend.sendChatMessage(parsedRoomID, value);
            setMessages((current) => [...current, response.message]);
            setBody('');
            setMessageImage(null);
            setMessageImageInputKey((current) => current + 1);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to send message.');
        } finally {
            setSending(false);
        }
    };

    const uploadRoomImage = async () => {
        if (!roomImage) return;
        setUploadingRoomImage(true);
        try {
            const response = await backend.uploadChatRoomImage(parsedRoomID, roomImage);
            setRoom(response.room);
            setRoomImage(null);
			setRoomImageInputKey((current) => current + 1);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to upload room image.');
        } finally {
            setUploadingRoomImage(false);
        }
    };

    const uploadRoomThemeImage = async () => {
        if (!themeImage) return;
        setSavingTheme(true);
        try {
            const response = await backend.uploadChatRoomThemeImage(parsedRoomID, themeImage);
            setRoom(response.room);
            setThemeImage(null);
            setThemeImageInputKey((current) => current + 1);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to upload chatroom theme image.');
        } finally {
            setSavingTheme(false);
        }
    };

    const removeRoomThemeImage = async () => {
        setSavingTheme(true);
        try {
            const response = await backend.removeChatRoomThemeImage(parsedRoomID);
            setRoom(response.room);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to remove chatroom theme image.');
        } finally {
            setSavingTheme(false);
        }
    };

    const saveRoomTheme = async () => {
        setSavingTheme(true);
        try {
            const response = await backend.updateChatRoomTheme(parsedRoomID, themeDraft, useCustomThemeColor ? customThemeColor : '');
            setRoom(response.room);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to update chatroom theme.');
        } finally {
            setSavingTheme(false);
        }
    };

    const createInvite = async () => {
        setCreatingInvite(true);
        try {
            const response = await backend.createChatInvite(parsedRoomID);
            const link = `${window.location.origin}/chat/invite/${response.inviteToken}`;
            setInviteLink(link);
            if (navigator.clipboard?.writeText) await navigator.clipboard.writeText(link);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to create an invite link.');
        } finally {
            setCreatingInvite(false);
        }
    };

    const toggleVisibility = async () => {
        if (!room) return;
        const visibility = room.visibility === 'public' ? 'invite_only' : 'public';
        setChangingVisibility(true);
        try {
            const response = await backend.updateChatRoomVisibility(parsedRoomID, visibility);
            setRoom(response.room);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to update room access.');
        } finally {
            setChangingVisibility(false);
        }
    };

    if (loading) return <Container className="py-5" role="status"><Spinner animation="border" size="sm" className="me-2" />Opening chatroom…</Container>;
    if (error || !room) return <Container className="py-5"><Alert variant="danger">{error || 'This chatroom is unavailable.'}</Alert><Link className="btn btn-outline-primary" to="/chat">Back to chatrooms</Link></Container>;

    return <Container className={`py-4 py-md-5 chat-room-page chat-room-theme-${room.theme || 'midnight'}`} style={presentationThemeStyle(room.customThemeColor, imageUrl(room.themeImagePath))}>
        <div className="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-4">
            <div><Link to="/chat" className="d-inline-block mb-2">← All chatrooms</Link><div className="d-flex align-items-start gap-3"><div className="chat-room-cover chat-room-cover-placeholder">{room.imagePath ? <img src={imageUrl(room.imagePath)} alt={`Cover for ${room.name}`} /> : '💬'}</div><div><div className="d-flex gap-2 align-items-center"><h1 className="h2 mb-0">{room.name}</h1><Badge bg={room.visibility === 'public' ? 'success' : 'secondary'}>{room.visibility === 'public' ? 'Public' : 'Private'}</Badge></div><p className="text-muted mb-0 mt-2">{room.memberCount} {room.memberCount === 1 ? 'member' : 'members'} · Created by {room.ownerDisplayName}</p></div></div></div>
            {room.isOwner && <div className="d-flex flex-wrap gap-2"><Button variant="outline-primary" onClick={() => void toggleVisibility()} disabled={changingVisibility}>{changingVisibility ? 'Saving access…' : room.visibility === 'public' ? 'Make private' : 'Make public'}</Button><Button variant="outline-primary" onClick={() => void createInvite()} disabled={creatingInvite}>{creatingInvite ? 'Creating invite…' : 'Create invite link'}</Button></div>}
        </div>

        {inviteLink && <Alert variant="info"><strong>Invite link ready.</strong> It has been copied when your browser allows it.<Form.Control className="mt-2" aria-label="Invite link" readOnly value={inviteLink} onFocus={(event) => event.currentTarget.select()} /></Alert>}
        {error && <Alert variant="danger">{error}</Alert>}

        {room.isOwner && <Card className="chat-room-image-card shadow-sm mb-3"><Card.Body><Form.Label htmlFor="chat-room-image" className="mb-1">Room image</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Control key={roomImageInputKey} id="chat-room-image" className="chat-image-input" type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => setRoomImage((event.currentTarget as HTMLInputElement).files?.[0] || null)} /><Button variant="outline-primary" onClick={() => void uploadRoomImage()} disabled={!roomImage || uploadingRoomImage}>{uploadingRoomImage ? 'Uploading…' : 'Upload image'}</Button></div><p className="text-muted small mb-0 mt-2">JPEG, PNG, or WebP · up to 5 MB. Replacing the image removes the old one.</p></Card.Body></Card>}

        {room.isOwner && <Card className="chat-room-theme-card shadow-sm mb-3"><Card.Body><Form.Label htmlFor="chat-room-theme" className="mb-1">Chatroom theme</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Select id="chat-room-theme" className="chat-theme-select" value={themeDraft} disabled={savingTheme} onChange={(event) => { setThemeDraft(event.target.value as DisplayTheme); setUseCustomThemeColor(false); }}>{chatThemeOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</Form.Select><Button variant="outline-primary" onClick={() => void saveRoomTheme()} disabled={savingTheme || (themeDraft === room.theme && (useCustomThemeColor ? customThemeColor : '') === room.customThemeColor)}>{savingTheme ? 'Saving theme…' : 'Save theme'}</Button></div><Form.Group className="mt-3" controlId="chat-room-custom-theme-color"><Form.Label className="small">Custom color</Form.Label><div className="d-flex align-items-center gap-2"><Form.Control type="color" value={customThemeColor} disabled={savingTheme} onChange={(event) => { setCustomThemeColor(event.target.value); setUseCustomThemeColor(true); }} aria-label="Custom chatroom theme color" /><span className="font-monospace text-uppercase">{customThemeColor}</span></div><Form.Text className="text-muted">Picking a color activates a custom room palette for every member.</Form.Text></Form.Group><Form.Group className="mt-3" controlId="chat-room-theme-image"><Form.Label className="small">Theme background image</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Control key={themeImageInputKey} className="theme-image-input" type="file" accept="image/jpeg,image/png,image/webp" disabled={savingTheme} onChange={(event) => setThemeImage((event.target as HTMLInputElement).files?.[0] ?? null)} /><Button variant="outline-primary" onClick={() => void uploadRoomThemeImage()} disabled={!themeImage || savingTheme}>{savingTheme ? 'Uploading…' : 'Upload image'}</Button>{room.themeImagePath && <Button variant="outline-secondary" onClick={() => void removeRoomThemeImage()} disabled={savingTheme}>Remove image</Button>}</div><Form.Text className="text-muted">JPEG, PNG, or WebP up to 5 MB. A 1920 × 1080 image works best and replaces the current background.</Form.Text></Form.Group><p className="text-muted small mb-0 mt-2">This theme is shown to everyone who opens this chatroom.</p></Card.Body></Card>}

        <Card className="chat-messages-card shadow-sm"><Card.Body className="chat-messages-list" aria-live="polite">
            {messages.length === 0 ? <p className="text-muted mb-0">No messages yet. Start the conversation.</p> : messages.map((message) => <article key={message.id} className="chat-message"><div className="d-flex justify-content-between gap-3"><strong>{message.authorDisplayName}</strong><time className="text-muted small" dateTime={message.createdAt}>{new Date(message.createdAt).toLocaleString()}</time></div>{message.body && <p className="mb-0 mt-1">{message.body}</p>}{message.imagePath && <a href={imageUrl(message.imagePath)} target="_blank" rel="noreferrer"><img className="chat-message-image mt-2" src={imageUrl(message.imagePath)} alt={`Shared by ${message.authorDisplayName}`} /></a>}</article>)}
        </Card.Body></Card>

        <Form className="mt-3" onSubmit={sendMessage}><Form.Label htmlFor="chat-message">Message</Form.Label><div className="d-flex gap-2"><Form.Control as="textarea" id="chat-message" rows={2} maxLength={2000} value={body} placeholder="Write a message…" onChange={(event) => setBody(event.target.value)} /><Button type="submit" disabled={sending || (!body.trim() && !messageImage)}>{sending ? 'Sending…' : 'Send'}</Button></div><Form.Control key={messageImageInputKey} className="chat-image-input mt-2" type="file" accept="image/jpeg,image/png,image/webp" aria-label="Attach image" onChange={(event) => setMessageImage((event.currentTarget as HTMLInputElement).files?.[0] || null)} />{messageImage && <p className="text-muted small mt-1 mb-0">Attaching {messageImage.name}</p>}</Form>
    </Container>;
};

export default ChatRoomPage;
