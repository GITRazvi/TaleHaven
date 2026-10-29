import React, { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Container, Form, Spinner } from 'react-bootstrap';
import { Link, useParams } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { ProfileComment, PublicProfile } from '../../app/clients/backend/types';
import { useAuth } from '../../contexts/AuthContext';

const formatMemberDate = (value: string): string => {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Unknown' : new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(date);
};

const PublicProfilePage: React.FC = () => {
  const backend = useMemo(() => new Backend(), []);
  const { user } = useAuth();
  const { userID } = useParams<{ userID: string }>();
  const parsedUserID = Number(userID);
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [error, setError] = useState('');
  const [comments, setComments] = useState<ProfileComment[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [commentsError, setCommentsError] = useState('');
  const [commentBody, setCommentBody] = useState('');
  const [commentSaving, setCommentSaving] = useState(false);
  const [deletingCommentID, setDeletingCommentID] = useState<number | null>(null);

  useEffect(() => {
    if (!Number.isInteger(parsedUserID) || parsedUserID < 1) {
      setError('This profile address is invalid.');
      return;
    }
    let current = true;
    backend.getPublicProfile(parsedUserID)
      .then((response) => { if (current) setProfile(response.profile); })
      .catch((requestError) => { if (current) setError(requestError instanceof Error ? requestError.message : 'Unable to load this profile.'); });
    return () => { current = false; };
  }, [backend, parsedUserID]);

  useEffect(() => {
    if (!profile?.commentsEnabled) {
      setComments([]);
      return;
    }
    let current = true;
    setCommentsLoading(true);
    setCommentsError('');
    backend.getProfileComments(profile.id)
      .then((response) => { if (current) setComments(response.data); })
      .catch((requestError) => { if (current) setCommentsError(requestError instanceof Error ? requestError.message : 'Unable to load comments.'); })
      .finally(() => { if (current) setCommentsLoading(false); });
    return () => { current = false; };
  }, [backend, profile]);

  if (error) return <Container className="py-5"><Alert variant="danger">{error}</Alert><Link className="btn btn-outline-primary" to="/chat">Back to chatrooms</Link></Container>;
  if (!profile) return <Container className="py-5" role="status"><Spinner animation="border" size="sm" className="me-2" />Loading profile…</Container>;

  const imageUrl = profile.profileImagePath ? `${backend.getBaseUrl()}${profile.profileImagePath}` : '';
  const initial = profile.displayName.trim().charAt(0).toUpperCase() || 'R';
  const submitComment = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!commentBody.trim() || commentSaving) return;
    setCommentSaving(true);
    setCommentsError('');
    try {
      const response = await backend.createProfileComment(profile.id, commentBody.trim());
      setComments((current) => [response.comment, ...current]);
      setCommentBody('');
    } catch (requestError) {
      setCommentsError(requestError instanceof Error ? requestError.message : 'Unable to post your comment.');
    } finally {
      setCommentSaving(false);
    }
  };
  const deleteComment = async (commentID: number) => {
    setDeletingCommentID(commentID);
    setCommentsError('');
    try {
      await backend.deleteProfileComment(profile.id, commentID);
      setComments((current) => current.filter((comment) => comment.id !== commentID));
    } catch (requestError) {
      setCommentsError(requestError instanceof Error ? requestError.message : 'Unable to delete this comment.');
    } finally {
      setDeletingCommentID(null);
    }
  };
  return <Container className="py-4 py-md-5 public-profile-page"><Link to="/chat" className="d-inline-block mb-3">← Back to chatrooms</Link><Card className="shadow-sm"><Card.Body className="d-flex flex-column flex-sm-row align-items-center gap-4 p-4 p-md-5 text-center text-sm-start">{imageUrl ? <img className="public-profile-avatar" src={imageUrl} alt={`${profile.displayName}'s profile`} /> : <div className="public-profile-avatar public-profile-avatar-fallback" aria-hidden="true">{initial}</div>}<div><p className="text-muted text-uppercase small mb-1">Reader profile</p><h1 className="display-6 mb-2">{profile.displayName}</h1><p className="text-muted mb-0">Member since {formatMemberDate(profile.createdAt)}</p></div></Card.Body></Card><section className="manga-comments-section mt-4" aria-labelledby="profile-comments-heading"><h2 id="profile-comments-heading" className="h3 mb-3">Comments</h2>{!profile.commentsEnabled ? <p className="text-muted mb-0">This user has turned off comments on their profile.</p> : <>{user && <Form onSubmit={submitComment} className="mb-3"><Form.Label htmlFor="profile-comment-body">Leave a comment</Form.Label><Form.Control as="textarea" id="profile-comment-body" rows={3} maxLength={1000} value={commentBody} onChange={(event) => setCommentBody(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter' && !event.shiftKey && !event.ctrlKey) { event.preventDefault(); event.currentTarget.form?.requestSubmit(); } }} placeholder={`Write to ${profile.displayName}…`} /><div className="d-flex justify-content-between align-items-center gap-3 mt-2"><Form.Text>{commentBody.length}/1000</Form.Text><Button type="submit" size="sm" disabled={commentSaving || !commentBody.trim()}>{commentSaving ? 'Posting…' : 'Post comment'}</Button></div></Form>}{commentsError && <Alert variant="danger">{commentsError}</Alert>}{commentsLoading ? <div className="text-center py-3"><Spinner animation="border" size="sm" className="me-2" />Loading comments…</div> : comments.length === 0 ? <p className="text-muted mb-0">No comments yet.</p> : <div className="d-grid gap-3">{comments.map((comment) => { const avatarURL = comment.profileImagePath ? `${backend.getBaseUrl()}${comment.profileImagePath}` : ''; const commentInitial = comment.authorDisplayName.trim().charAt(0).toUpperCase() || 'R'; const canDelete = user?.id === comment.userId || user?.id === profile.id; return <article key={comment.id} className="manga-comment"><div className="d-flex align-items-start gap-3"><Link to={`/profile/${comment.userId}`} className="manga-comment-avatar" aria-label={`View ${comment.authorDisplayName}'s profile`}>{avatarURL ? <img src={avatarURL} alt="" /> : <span aria-hidden="true">{commentInitial}</span>}</Link><div className="flex-grow-1 min-width-0"><div className="d-flex flex-wrap align-items-center gap-2"><Link to={`/profile/${comment.userId}`} className="manga-comment-author">{comment.authorDisplayName}</Link>{canDelete && <Button variant="link" size="sm" className="ms-auto p-0 text-danger" onClick={() => void deleteComment(comment.id)} disabled={deletingCommentID === comment.id}>{deletingCommentID === comment.id ? 'Deleting…' : 'Delete'}</Button>}</div><p className="mb-0 mt-1" style={{ whiteSpace: 'pre-wrap' }}>{comment.body}</p></div></div></article>; })}</div>}</>}</section></Container>;
};

export default PublicProfilePage;
