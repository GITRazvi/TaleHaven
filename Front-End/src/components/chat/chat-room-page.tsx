import React, { FormEvent, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { Alert, Badge, Button, Card, Container, Dropdown, Form, Modal, OverlayTrigger, Popover, Spinner } from 'react-bootstrap';
import { Link, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { ChatMessage, ChatRoom, DisplayTheme } from '../../app/clients/backend/types';
import { customThemeStyle, presentationThemeStyle, themeBackgroundStyle } from '../../app/theme-style';
import { useAuth } from '../../contexts/AuthContext';

const REFRESH_INTERVAL_MS = 5000;

const chatThemeOptions: Array<{ value: DisplayTheme; label: string }> = [
    { value: 'midnight', label: 'Midnight' },
    { value: 'forest', label: 'Forest' },
    { value: 'plum', label: 'Plum' },
    { value: 'ember', label: 'Ember' },
    { value: 'parchment', label: 'Parchment' },
];

const chatThemeBackgrounds: Record<DisplayTheme, string> = {
    midnight: 'linear-gradient(145deg, #0b1220, #182b4b)',
    forest: 'linear-gradient(145deg, #0a1d18, #1a473a)',
    plum: 'linear-gradient(145deg, #1b1022, #432652)',
    ember: 'linear-gradient(145deg, #201310, #503127)',
    parchment: 'linear-gradient(145deg, #e8dfca, #d6bea0)',
};

const ChatRoomPage: React.FC = () => {
    const backend = useMemo(() => new Backend(), []);
    const { user } = useAuth();
    const { roomID } = useParams<{ roomID: string }>();
    const parsedRoomID = Number(roomID);
    const [room, setRoom] = useState<ChatRoom | null>(null);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [body, setBody] = useState('');
    const [messageImage, setMessageImage] = useState<File | null>(null);
    const [sending, setSending] = useState(false);
    const [deletingMessageID, setDeletingMessageID] = useState<number | null>(null);
    const [replyToMessage, setReplyToMessage] = useState<ChatMessage | null>(null);
    const [reportMessage, setReportMessage] = useState<ChatMessage | null>(null);
    const [reportReason, setReportReason] = useState('');
    const [reporting, setReporting] = useState(false);
    const [notice, setNotice] = useState<string | null>(null);
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
    const [showRoomSettings, setShowRoomSettings] = useState(false);
    const [followingNewestMessage, setFollowingNewestMessage] = useState(true);
    const messageListRef = useRef<HTMLDivElement>(null);

    const imageUrl = useCallback((imagePath: string) => imagePath ? `${backend.getBaseUrl()}${imagePath}` : '', [backend]);

    const scrollToMessage = useCallback((messageID: number) => {
        document.querySelector<HTMLElement>(`[data-chat-message-id="${messageID}"]`)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }, []);

    const authorPreview = (message: ChatMessage) => {
        const initial = message.authorDisplayName.trim().charAt(0).toUpperCase() || 'R';
        return <OverlayTrigger trigger={['hover', 'focus']} placement="right" overlay={<Popover className="chat-author-preview"><Popover.Body>{message.profileImagePath ? <img src={imageUrl(message.profileImagePath)} alt={`${message.authorDisplayName}'s profile`} /> : <div className="chat-author-preview-fallback" aria-hidden="true">{initial}</div>}<span>{message.authorDisplayName}</span></Popover.Body></Popover>}><button type="button" className="chat-author-name">{message.authorDisplayName}</button></OverlayTrigger>;
    };

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

    const scrollToNewestMessage = useCallback(() => {
        const messageList = messageListRef.current;
        if (!messageList || !followingNewestMessage) return;
        messageList.scrollTop = messageList.scrollHeight;
    }, [followingNewestMessage]);

    useLayoutEffect(() => {
        scrollToNewestMessage();
    }, [messages, scrollToNewestMessage]);

    useEffect(() => {
        if (!room) return;
        setThemeDraft(room.theme || 'midnight');
        setCustomThemeColor(room.customThemeColor || '#7ab4ff');
        setUseCustomThemeColor(Boolean(room.customThemeColor));
    }, [room]);

    useEffect(() => {
        if (!room) return undefined;
        const customStyle = customThemeStyle(room.customThemeColor);
        const imageStyle = themeBackgroundStyle(imageUrl(room.themeImagePath));
        const background = imageStyle?.backgroundImage || (customStyle
            ? `linear-gradient(145deg, ${customStyle['--theme-page']}, ${customStyle['--theme-surface-strong']})`
            : chatThemeBackgrounds[room.theme || 'midnight']);
        const previousBodyBackground = document.body.style.background;
        const previousBodyBackgroundColor = document.body.style.backgroundColor;
        const previousBodyBackgroundImage = document.body.style.backgroundImage;
        const previousBodyBackgroundSize = document.body.style.backgroundSize;
        const previousBodyBackgroundPosition = document.body.style.backgroundPosition;
        const previousBodyBackgroundAttachment = document.body.style.backgroundAttachment;
        const previousDocumentBackground = document.documentElement.style.background;
        const previousDocumentBackgroundColor = document.documentElement.style.backgroundColor;
        const previousDocumentBackgroundImage = document.documentElement.style.backgroundImage;
        const previousDocumentBackgroundSize = document.documentElement.style.backgroundSize;
        const previousDocumentBackgroundPosition = document.documentElement.style.backgroundPosition;
        const previousDocumentBackgroundAttachment = document.documentElement.style.backgroundAttachment;
        document.body.style.background = background;
        document.documentElement.style.background = background;
        if (imageStyle?.backgroundImage) {
            document.body.classList.add('scoped-theme-image-active');
            document.body.style.backgroundImage = imageStyle.backgroundImage;
            document.body.style.backgroundSize = 'cover';
            document.body.style.backgroundPosition = 'center';
            document.body.style.backgroundAttachment = 'fixed';
            document.documentElement.style.backgroundImage = imageStyle.backgroundImage;
            document.documentElement.style.backgroundSize = 'cover';
            document.documentElement.style.backgroundPosition = 'center';
            document.documentElement.style.backgroundAttachment = 'fixed';
        }
        return () => {
            document.body.classList.remove('scoped-theme-image-active');
            document.body.style.background = previousBodyBackground;
            document.body.style.backgroundColor = previousBodyBackgroundColor;
            document.body.style.backgroundImage = previousBodyBackgroundImage;
            document.body.style.backgroundSize = previousBodyBackgroundSize;
            document.body.style.backgroundPosition = previousBodyBackgroundPosition;
            document.body.style.backgroundAttachment = previousBodyBackgroundAttachment;
            document.documentElement.style.background = previousDocumentBackground;
            document.documentElement.style.backgroundColor = previousDocumentBackgroundColor;
            document.documentElement.style.backgroundImage = previousDocumentBackgroundImage;
            document.documentElement.style.backgroundSize = previousDocumentBackgroundSize;
            document.documentElement.style.backgroundPosition = previousDocumentBackgroundPosition;
            document.documentElement.style.backgroundAttachment = previousDocumentBackgroundAttachment;
        };
    }, [room, imageUrl]);

    const sendMessage = async (event: FormEvent) => {
        event.preventDefault();
        const value = body.trim();
        if (!value && !messageImage) return;
        setSending(true);
        try {
            const response = messageImage
                ? await backend.sendChatMessageWithImage(parsedRoomID, value, messageImage, replyToMessage?.id ?? 0)
                : await backend.sendChatMessage(parsedRoomID, value, replyToMessage?.id ?? 0);
            setMessages((current) => [...current, response.message]);
            setBody('');
            setMessageImage(null);
            setReplyToMessage(null);
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to send message.');
        } finally {
            setSending(false);
        }
    };

    const attachPastedImage = (event: React.ClipboardEvent<HTMLTextAreaElement>) => {
        const imageItem = Array.from(event.clipboardData.items).find((item) => item.kind === 'file' && item.type.startsWith('image/'));
        const image = imageItem?.getAsFile();
        if (!image) return;
        event.preventDefault();
        setMessageImage(image);
        setError(null);
    };

    const handleMessageKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (event.key !== 'Enter' || event.shiftKey || event.nativeEvent.isComposing) return;
        event.preventDefault();
        event.currentTarget.form?.requestSubmit();
    };

    const handleMessageListScroll = (event: React.UIEvent<HTMLDivElement>) => {
        const messageList = event.currentTarget;
        const distanceFromNewest = messageList.scrollHeight - messageList.scrollTop - messageList.clientHeight;
        setFollowingNewestMessage(distanceFromNewest <= 24);
    };

    const deleteMessage = async (message: ChatMessage) => {
        setDeletingMessageID(message.id);
        try {
            await backend.deleteChatMessage(parsedRoomID, message.id);
            setMessages((current) => current.filter((currentMessage) => currentMessage.id !== message.id));
            setError(null);
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to delete message.');
        } finally {
            setDeletingMessageID(null);
        }
    };

    const submitReport = async (event: FormEvent) => {
        event.preventDefault();
        if (!reportMessage || !reportReason.trim()) return;
        setReporting(true);
        try {
            await backend.reportChatMessage(parsedRoomID, reportMessage.id, reportReason.trim());
            setNotice('Report submitted. Thank you for letting us know.');
            setReportMessage(null);
            setReportReason('');
        } catch (requestError) {
            setError(requestError instanceof Error ? requestError.message : 'Unable to report message.');
        } finally {
            setReporting(false);
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
    if (!room) return <Container className="py-5"><Alert variant="danger">{error || 'This chatroom is unavailable.'}</Alert><Link className="btn btn-outline-primary" to="/chat">Back to chatrooms</Link></Container>;

    return <Container className={`py-4 py-md-5 chat-room-page chat-room-theme-${room.theme || 'midnight'}${room.themeImagePath ? ' theme-has-image' : ''}`} style={presentationThemeStyle(room.customThemeColor, imageUrl(room.themeImagePath))}>
        <div className="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-4">
            <div><Link to="/chat" className="d-inline-block mb-2">← All chatrooms</Link><div className="d-flex align-items-start gap-3"><div className="chat-room-cover chat-room-cover-placeholder">{room.imagePath ? <img src={imageUrl(room.imagePath)} alt={`Cover for ${room.name}`} /> : '💬'}</div><div><div className="d-flex gap-2 align-items-center"><h1 className="h2 mb-0">{room.name}</h1><Badge bg={room.visibility === 'public' ? 'success' : 'secondary'}>{room.visibility === 'public' ? 'Public' : 'Private'}</Badge></div><p className="text-muted mb-0 mt-2">{room.memberCount} {room.memberCount === 1 ? 'member' : 'members'} · Created by {room.ownerDisplayName}</p></div></div></div>
            {room.isOwner && <Button variant="outline-primary" onClick={() => setShowRoomSettings((current) => !current)} aria-expanded={showRoomSettings} aria-controls="chat-room-settings">{showRoomSettings ? 'Close settings' : 'Room settings'}</Button>}
        </div>

        {error && <Alert variant="danger">{error}</Alert>}
        {notice && <Alert variant="success" dismissible onClose={() => setNotice(null)}>{notice}</Alert>}

        {room.isOwner && showRoomSettings && <Card id="chat-room-settings" className="chat-room-theme-card shadow-sm mb-3"><Card.Body>
            <div className="d-flex flex-wrap align-items-center justify-content-between gap-2"><h2 className="h5 mb-0">Room settings</h2><Button variant="outline-secondary" size="sm" onClick={() => setShowRoomSettings(false)}>Close</Button></div>
            <hr />
            <h3 className="h6">Access and invitations</h3>
            <div className="d-flex flex-wrap gap-2"><Button variant="outline-primary" onClick={() => void toggleVisibility()} disabled={changingVisibility}>{changingVisibility ? 'Saving access…' : room.visibility === 'public' ? 'Make private' : 'Make public'}</Button><Button variant="outline-primary" onClick={() => void createInvite()} disabled={creatingInvite}>{creatingInvite ? 'Creating invite…' : 'Create invite link'}</Button></div>
            {inviteLink && <Alert className="mt-3 mb-0" variant="info"><strong>Invite link ready.</strong> It has been copied when your browser allows it.<Form.Control className="mt-2" aria-label="Invite link" readOnly value={inviteLink} onFocus={(event) => event.currentTarget.select()} /></Alert>}
            <hr />
            <Form.Label htmlFor="chat-room-image" className="mb-1">Room image</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Control key={roomImageInputKey} id="chat-room-image" className="chat-image-input" type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => setRoomImage((event.currentTarget as HTMLInputElement).files?.[0] || null)} /><Button variant="outline-primary" onClick={() => void uploadRoomImage()} disabled={!roomImage || uploadingRoomImage}>{uploadingRoomImage ? 'Uploading…' : 'Upload image'}</Button></div><p className="text-muted small mb-0 mt-2">JPEG, PNG, or WebP · up to 5 MB. Replacing the image removes the old one.</p>
            <hr />
            <Form.Label htmlFor="chat-room-theme" className="mb-1">Chatroom theme</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Select id="chat-room-theme" className="chat-theme-select" value={themeDraft} disabled={savingTheme} onChange={(event) => { setThemeDraft(event.target.value as DisplayTheme); setUseCustomThemeColor(false); }}>{chatThemeOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</Form.Select><Button variant="outline-primary" onClick={() => void saveRoomTheme()} disabled={savingTheme || (themeDraft === room.theme && (useCustomThemeColor ? customThemeColor : '') === room.customThemeColor)}>{savingTheme ? 'Saving theme…' : 'Save theme'}</Button></div><Form.Group className="mt-3" controlId="chat-room-custom-theme-color"><Form.Label className="small">Custom color</Form.Label><div className="d-flex align-items-center gap-2"><Form.Control type="color" value={customThemeColor} disabled={savingTheme} onChange={(event) => { setCustomThemeColor(event.target.value); setUseCustomThemeColor(true); }} aria-label="Custom chatroom theme color" /><span className="font-monospace text-uppercase">{customThemeColor}</span></div><Form.Text className="text-muted">Picking a color activates a custom room palette for every member.</Form.Text></Form.Group><Form.Group className="mt-3" controlId="chat-room-theme-image"><Form.Label className="small">Theme background image</Form.Label><div className="d-flex flex-wrap align-items-center gap-2"><Form.Control key={themeImageInputKey} className="theme-image-input" type="file" accept="image/jpeg,image/png,image/webp,image/gif" disabled={savingTheme} onChange={(event) => setThemeImage((event.target as HTMLInputElement).files?.[0] ?? null)} /><Button variant="outline-primary" onClick={() => void uploadRoomThemeImage()} disabled={!themeImage || savingTheme}>{savingTheme ? 'Uploading…' : 'Upload image'}</Button>{room.themeImagePath && <Button variant="outline-secondary" onClick={() => void removeRoomThemeImage()} disabled={savingTheme}>Remove image</Button>}</div><Form.Text className="text-muted">JPEG, PNG, WebP, or animated GIF up to 5 MB. GIFs play as backgrounds. A 1920 × 1080 image works best and replaces the current background.</Form.Text></Form.Group>
        </Card.Body></Card>}

        <Card className="chat-messages-card shadow-sm"><Card.Body ref={messageListRef} className="chat-messages-list" aria-live="polite" onScroll={handleMessageListScroll}>
            {messages.length === 0 ? <p className="text-muted mb-0">No messages yet. Start the conversation.</p> : messages.map((message) => { const parent = messages.find((candidate) => candidate.id === message.replyToMessageId); return <article key={message.id} data-chat-message-id={message.id} className="chat-message"><div className="d-flex justify-content-between gap-3">{authorPreview(message)}<div className="d-flex align-items-center gap-2"><time className="text-muted small" dateTime={message.createdAt}>{new Date(message.createdAt).toLocaleString()}</time><Dropdown align="end"><Dropdown.Toggle size="sm" variant="link" className="chat-message-actions p-0 border-0" aria-label={`Message actions for ${message.body || 'image message'}`}>•••</Dropdown.Toggle><Dropdown.Menu><Dropdown.Item onClick={() => setReplyToMessage(message)}>Reply</Dropdown.Item>{message.userId === user?.id ? <Dropdown.Item className="text-danger" onClick={() => void deleteMessage(message)} disabled={deletingMessageID === message.id}>{deletingMessageID === message.id ? 'Deleting…' : 'Delete'}</Dropdown.Item> : <Dropdown.Item className="text-warning" onClick={() => { setReportMessage(message); setReportReason(''); }}>Report</Dropdown.Item>}</Dropdown.Menu></Dropdown></div></div>{message.replyToMessageId > 0 && <Button variant="link" className="chat-reply-reference p-0 mt-1 text-start" onClick={() => scrollToMessage(message.replyToMessageId)} disabled={!parent} aria-label={parent ? `View replied message from ${parent.authorDisplayName}` : 'Original replied message is unavailable'}>{parent ? <>↪ {parent.authorDisplayName}: {parent.body || 'Image message'}</> : '↪ Original message is unavailable'}</Button>}{message.body && <p className="mb-0 mt-1">{message.body}</p>}{message.imagePath && <a href={imageUrl(message.imagePath)} target="_blank" rel="noreferrer"><img className="chat-message-image mt-2" src={imageUrl(message.imagePath)} alt={`Shared by ${message.authorDisplayName}`} onLoad={scrollToNewestMessage} /></a>}</article>; })}
        </Card.Body></Card>

        <Form className="mt-3" onSubmit={sendMessage}>{replyToMessage && <Alert variant="secondary" className="py-2 d-flex justify-content-between align-items-center">Replying to <strong>{replyToMessage.authorDisplayName}</strong><Button size="sm" variant="link" onClick={() => setReplyToMessage(null)}>Cancel</Button></Alert>}<Form.Label htmlFor="chat-message">Message</Form.Label><div className="d-flex gap-2"><Form.Control as="textarea" id="chat-message" rows={2} maxLength={2000} value={body} placeholder="Write a message or paste an image…" onChange={(event) => setBody(event.target.value)} onKeyDown={handleMessageKeyDown} onPaste={attachPastedImage} /><Button type="submit" disabled={sending || (!body.trim() && !messageImage)}>{sending ? 'Sending…' : 'Send'}</Button></div>{messageImage && <div className="d-flex align-items-center gap-2 mt-2"><p className="text-muted small mb-0">Pasted image attached: {messageImage.name || 'image'}</p><Button size="sm" variant="outline-secondary" onClick={() => setMessageImage(null)}>Remove image</Button></div>}<Form.Text className="text-muted">Press Enter to send, Shift+Enter for a new line. Paste an image directly into the message box to attach it.</Form.Text></Form>
        <Modal show={reportMessage !== null} onHide={() => !reporting && setReportMessage(null)} centered><Form onSubmit={submitReport}><Modal.Header closeButton><Modal.Title>Report message</Modal.Title></Modal.Header><Modal.Body><p className="text-muted">Tell us why you are reporting this message from {reportMessage?.authorDisplayName}.</p><Form.Control as="textarea" rows={4} maxLength={500} required value={reportReason} onChange={(event) => setReportReason(event.target.value)} placeholder="Describe the issue…" aria-label="Report reason" /></Modal.Body><Modal.Footer><Button variant="outline-secondary" onClick={() => setReportMessage(null)} disabled={reporting}>Cancel</Button><Button variant="danger" type="submit" disabled={reporting || !reportReason.trim()}>{reporting ? 'Sending…' : 'Submit report'}</Button></Modal.Footer></Form></Modal>
    </Container>;
};

export default ChatRoomPage;
