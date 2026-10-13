import React, { useState } from 'react';
import { Container, Card, Button, Modal, Form, Alert, Row, Col } from 'react-bootstrap';
import { useAuth } from '../../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';

const backend = new Backend();

const ProfilePage: React.FC = () => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();

    const [showChangePassword, setShowChangePassword] = useState(false);
    const [showDeleteAccount, setShowDeleteAccount] = useState(false);

    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmNewPassword, setConfirmNewPassword] = useState('');

    const [passwordError, setPasswordError] = useState('');
    const [passwordSuccess, setPasswordSuccess] = useState('');
    const [deleteError, setDeleteError] = useState('');

    const [loadingPassword, setLoadingPassword] = useState(false);
    const [loadingDelete, setLoadingDelete] = useState(false);

    if (!user) {
        return (
            <Container className="mt-5">
                <Alert variant="warning">Please login to view your profile.</Alert>
            </Container>
        );
    }

    const handleChangePassword = async (e: React.FormEvent) => {
        e.preventDefault();
        setPasswordError('');
        setPasswordSuccess('');

        if (newPassword.length < 8) {
            setPasswordError('New password must be at least 8 characters long');
            return;
        }

        if (newPassword !== confirmNewPassword) {
            setPasswordError('New passwords do not match');
            return;
        }

        try {
            setLoadingPassword(true);
            await backend.changePassword({
                old_password: oldPassword,
                new_password: newPassword,
            });
            setPasswordSuccess('Password changed successfully!');
            setOldPassword('');
            setNewPassword('');
            setConfirmNewPassword('');
            setTimeout(() => {
                setShowChangePassword(false);
                setPasswordSuccess('');
            }, 2000);
        } catch (err: any) {
            setPasswordError(err?.details?.message || err?.message || 'Failed to change password');
        } finally {
            setLoadingPassword(false);
        }
    };

    const handleDeleteAccount = async () => {
        setDeleteError('');
        try {
            setLoadingDelete(true);
            await backend.deleteAccount();
            await logout();
            navigate('/');
        } catch (err: any) {
            // Handle account deletion errors (requires valid JWT token)
            setDeleteError(err?.details?.message || err?.message || 'Failed to delete account');
        } finally {
            setLoadingDelete(false);
        }
    };

    const handleLogout = async () => {
        await logout();
        navigate('/');
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });
    };

    return (
        <Container className="mt-5" style={{ maxWidth: '800px' }}>
            <h2 className="mb-4">My Profile</h2>

            <Card className="mb-4">
                <Card.Body>
                    <Card.Title>Account Information</Card.Title>
                    <Row className="mt-3">
                        <Col sm={3}>
                            <strong>Email:</strong>
                        </Col>
                        <Col sm={9}>
                            {user.email}
                        </Col>
                    </Row>
                    <Row className="mt-2">
                        <Col sm={3}>
                            <strong>User ID:</strong>
                        </Col>
                        <Col sm={9}>
                            {user.id}
                        </Col>
                    </Row>
                    <Row className="mt-2">
                        <Col sm={3}>
                            <strong>Member Since:</strong>
                        </Col>
                        <Col sm={9}>
                            {formatDate(user.created_at)}
                        </Col>
                    </Row>
                </Card.Body>
            </Card>

            <Card className="mb-4">
                <Card.Body>
                    <Card.Title>Account Actions</Card.Title>
                    <div className="d-grid gap-2 mt-3">
                        <Button
                            variant="primary"
                            onClick={() => setShowChangePassword(true)}
                        >
                            Change Password
                        </Button>
                        <Button
                            variant="secondary"
                            onClick={handleLogout}
                        >
                            Logout
                        </Button>
                        <Button
                            variant="danger"
                            onClick={() => setShowDeleteAccount(true)}
                        >
                            Delete Account
                        </Button>
                    </div>
                </Card.Body>
            </Card>

            {/* Change Password Modal */}
            <Modal show={showChangePassword} onHide={() => setShowChangePassword(false)}>
                <Modal.Header closeButton>
                    <Modal.Title>Change Password</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    {passwordError && <Alert variant="danger">{passwordError}</Alert>}
                    {passwordSuccess && <Alert variant="success">{passwordSuccess}</Alert>}
                    <Form onSubmit={handleChangePassword}>
                        <Form.Group className="mb-3">
                            <Form.Label>Current Password</Form.Label>
                            <Form.Control
                                type="password"
                                value={oldPassword}
                                onChange={(e) => setOldPassword(e.target.value)}
                                required
                            />
                        </Form.Group>
                        <Form.Group className="mb-3">
                            <Form.Label>New Password</Form.Label>
                            <Form.Control
                                type="password"
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                                required
                                minLength={8}
                                placeholder="Min 8 characters"
                            />
                        </Form.Group>
                        <Form.Group className="mb-3">
                            <Form.Label>Confirm New Password</Form.Label>
                            <Form.Control
                                type="password"
                                value={confirmNewPassword}
                                onChange={(e) => setConfirmNewPassword(e.target.value)}
                                required
                            />
                        </Form.Group>
                        <div className="d-grid gap-2">
                            <Button type="submit" variant="primary" disabled={loadingPassword}>
                                {loadingPassword ? 'Changing...' : 'Change Password'}
                            </Button>
                            <Button
                                variant="secondary"
                                onClick={() => setShowChangePassword(false)}
                                disabled={loadingPassword}
                            >
                                Cancel
                            </Button>
                        </div>
                    </Form>
                </Modal.Body>
            </Modal>

            {/* Delete Account Modal */}
            <Modal show={showDeleteAccount} onHide={() => setShowDeleteAccount(false)}>
                <Modal.Header closeButton>
                    <Modal.Title>Delete Account</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    {deleteError && <Alert variant="danger">{deleteError}</Alert>}
                    <Alert variant="warning">
                        <strong>Warning!</strong> This action cannot be undone. All your reading progress and bookmarks will be permanently deleted.
                    </Alert>
                    <p>Are you sure you want to delete your account?</p>
                </Modal.Body>
                <Modal.Footer>
                    <Button
                        variant="secondary"
                        onClick={() => setShowDeleteAccount(false)}
                        disabled={loadingDelete}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant="danger"
                        onClick={handleDeleteAccount}
                        disabled={loadingDelete}
                    >
                        {loadingDelete ? 'Deleting...' : 'Yes, Delete My Account'}
                    </Button>
                </Modal.Footer>
            </Modal>
        </Container>
    );
};

export default ProfilePage;


