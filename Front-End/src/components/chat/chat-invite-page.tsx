import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Container, Spinner } from 'react-bootstrap';
import { useNavigate, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';

const ChatInvitePage: React.FC = () => {
    const backend = useMemo(() => new Backend(), []);
    const navigate = useNavigate();
    const { token } = useParams<{ token: string }>();
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const join = async () => {
            if (!token) { setError('This invite link is invalid.'); return; }
            try {
                const response = await backend.joinChatInvite(token);
                if (!response.roomId) throw new Error('This invite did not identify a chatroom.');
                navigate(`/chat/rooms/${response.roomId}`, { replace: true });
            } catch (requestError) {
                setError(requestError instanceof Error ? requestError.message : 'Unable to use this invite link.');
            }
        };
        void join();
    }, [backend, navigate, token]);

    return <Container className="py-5" style={{ maxWidth: '42rem' }}>{error ? <Alert variant="danger">{error}</Alert> : <div className="d-flex gap-2 align-items-center" role="status"><Spinner animation="border" size="sm" />Joining chatroom…</div>}</Container>;
};

export default ChatInvitePage;
