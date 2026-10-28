import React, { useEffect, useState } from 'react';
import { Container, Card, Button, Modal, Form, Alert, Row, Col } from 'react-bootstrap';
import { useAuth } from '../../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { DisplayTheme, SiteThemeImageRotationMode } from '../../app/clients/backend/types';
import { presentationThemeStyle } from '../../app/theme-style';

const backend = new Backend();

const languageOptions = [
    ['all', 'All languages'], ['en', 'English'], ['ro', 'Romanian'],
    ['es', 'Spanish'], ['fr', 'French'], ['de', 'German'], ['it', 'Italian'],
    ['pt-br', 'Portuguese (Brazil)'], ['pl', 'Polish'], ['ru', 'Russian'],
    ['ja', 'Japanese'], ['ko', 'Korean'], ['zh', 'Chinese'],
];

const displayThemeOptions: Array<{ value: DisplayTheme; label: string; description: string }> = [
    { value: 'midnight', label: 'Midnight', description: 'Deep blue and cool silver.' },
    { value: 'forest', label: 'Forest', description: 'Evergreen with warm moss accents.' },
    { value: 'plum', label: 'Plum', description: 'Rich violet and soft lavender.' },
    { value: 'ember', label: 'Ember', description: 'Charcoal with copper warmth.' },
    { value: 'parchment', label: 'Parchment', description: 'A calm, light paper palette.' },
];

const ProfilePage: React.FC = () => {
    const { user, logout, refreshUser } = useAuth();
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
    const [preferredTranslationLanguage, setPreferredTranslationLanguage] = useState('all');
    const [preferredReadingMode, setPreferredReadingMode] = useState<'vertical' | 'ltr' | 'rtl'>('vertical');
    const [readingSettingsError, setReadingSettingsError] = useState('');
    const [readingSettingsSuccess, setReadingSettingsSuccess] = useState('');
    const [loadingReadingSettings, setLoadingReadingSettings] = useState(true);
    const [savingReadingSettings, setSavingReadingSettings] = useState(false);
    const [displayName, setDisplayName] = useState('');
    const [profilePhoto, setProfilePhoto] = useState<File | null>(null);
    const [profileError, setProfileError] = useState('');
    const [profileSuccess, setProfileSuccess] = useState('');
    const [savingProfile, setSavingProfile] = useState(false);
    const [displayTheme, setDisplayTheme] = useState<DisplayTheme>('midnight');
    const [displayThemeError, setDisplayThemeError] = useState('');
    const [displayThemeSuccess, setDisplayThemeSuccess] = useState('');
    const [savingDisplayTheme, setSavingDisplayTheme] = useState(false);
    const [customThemeColor, setCustomThemeColor] = useState('#7ab4ff');
    const [useCustomThemeColor, setUseCustomThemeColor] = useState(false);
    const [themeImage, setThemeImage] = useState<File | null>(null);
    const [themeImageInputKey, setThemeImageInputKey] = useState(0);
    const [siteTheme, setSiteTheme] = useState<DisplayTheme>('midnight');
    const [siteCustomColor, setSiteCustomColor] = useState('#7ab4ff');
    const [useCustomSiteColor, setUseCustomSiteColor] = useState(false);
    const [siteThemeError, setSiteThemeError] = useState('');
    const [siteThemeSuccess, setSiteThemeSuccess] = useState('');
    const [savingSiteTheme, setSavingSiteTheme] = useState(false);
    const [siteThemeImages, setSiteThemeImages] = useState<File[]>([]);
    const [siteThemeImageInputKey, setSiteThemeImageInputKey] = useState(0);
    const [siteThemeImageRotationMode, setSiteThemeImageRotationMode] = useState<SiteThemeImageRotationMode>('static');
    const [siteThemeImageRotationMinutes, setSiteThemeImageRotationMinutes] = useState(15);

    useEffect(() => {
        if (user) {
            setDisplayName(user.displayName || user.email.split('@')[0]);
            setDisplayTheme(user.displayTheme || 'midnight');
            setCustomThemeColor(user.customThemeColor || '#7ab4ff');
            setUseCustomThemeColor(Boolean(user.customThemeColor));
            setSiteTheme(user.siteTheme || 'midnight');
            setSiteCustomColor(user.siteCustomColor || '#7ab4ff');
            setUseCustomSiteColor(Boolean(user.siteCustomColor));
            setSiteThemeImageRotationMode(user.siteThemeImageRotationMode || 'static');
            setSiteThemeImageRotationMinutes(user.siteThemeImageRotationMinutes || 15);
        }
    }, [user]);

    useEffect(() => {
        let isCurrent = true;
        backend.getReadingSettings()
            .then((response) => {
                if (isCurrent) {
                    setPreferredTranslationLanguage(response.preferredTranslationLanguage || 'all');
                    setPreferredReadingMode(response.preferredReadingMode || 'vertical');
                }
            })
            .catch((err) => {
                if (isCurrent) setReadingSettingsError(err?.message || 'Unable to load reading settings.');
            })
            .finally(() => {
                if (isCurrent) setLoadingReadingSettings(false);
            });
        return () => { isCurrent = false; };
    }, []);

    if (!user) {
        return (
            <Container className="mt-5">
                <Alert variant="warning">Please login to view your profile.</Alert>
            </Container>
        );
    }

    const profileImageUrl = user.profileImagePath
        ? `${process.env.REACT_APP_API_BASE_URL ?? 'http://localhost:3001'}${user.profileImagePath}`
        : '';
    const themeImageUrl = user.themeImagePath
        ? `${process.env.REACT_APP_API_BASE_URL ?? 'http://localhost:3001'}${user.themeImagePath}`
        : '';
    const siteThemeImagePaths = user.siteThemeImagePaths || [];
    const siteThemeImageUrl = siteThemeImagePaths[0]
        ? `${process.env.REACT_APP_API_BASE_URL ?? 'http://localhost:3001'}${siteThemeImagePaths[0]}`
        : '';
    const avatarInitial = (user.displayName || user.email).trim().charAt(0).toUpperCase() || 'R';

    const handleSaveProfile = async (event: React.FormEvent) => {
        event.preventDefault();
        const nextDisplayName = displayName.trim();
        if (!nextDisplayName) {
            setProfileError('Display name is required.');
            return;
        }
        try {
            setSavingProfile(true);
            setProfileError('');
            setProfileSuccess('');
            if (nextDisplayName !== user.displayName) {
                await backend.updateProfile({ displayName: nextDisplayName });
            }
            if (profilePhoto) {
                await backend.uploadProfilePhoto(profilePhoto);
                setProfilePhoto(null);
            }
            await refreshUser();
            setProfileSuccess('Profile updated.');
        } catch (err: any) {
            setProfileError(err?.details?.message || err?.message || 'Unable to update your profile.');
        } finally {
            setSavingProfile(false);
        }
    };

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
            // The backend revokes every session after a password change.
            // Clear the local user immediately and tell the user why Login is
            // required again instead of waiting for a later request to fail.
            await logout();
            navigate('/login', {
                replace: true,
                state: { notice: 'Password changed. Please log in again.' },
            });
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
            setDeleteError(err?.details?.message || err?.message || 'Failed to delete account');
        } finally {
            setLoadingDelete(false);
        }
    };

    const handleLogout = async () => {
        await logout();
        navigate('/');
    };

    const handleSaveReadingSettings = async () => {
        setReadingSettingsError('');
        setReadingSettingsSuccess('');
        try {
            setSavingReadingSettings(true);
            const response = await backend.updateReadingSettings({ preferredTranslationLanguage, preferredReadingMode });
            setPreferredTranslationLanguage(response.preferredTranslationLanguage);
            setPreferredReadingMode(response.preferredReadingMode || 'vertical');
            setReadingSettingsSuccess('Reading preferences saved.');
        } catch (err: any) {
            setReadingSettingsError(err?.message || 'Unable to save reading settings.');
        } finally {
            setSavingReadingSettings(false);
        }
    };

    const handleSaveDisplayTheme = async () => {
        setDisplayThemeError('');
        setDisplayThemeSuccess('');
        try {
            setSavingDisplayTheme(true);
            await backend.updateProfileTheme({ displayTheme, customThemeColor: useCustomThemeColor ? customThemeColor : '' });
            await refreshUser();
            setDisplayThemeSuccess('Profile display theme saved.');
        } catch (err: any) {
            setDisplayThemeError(err?.details?.message || err?.message || 'Unable to save display theme.');
        } finally {
            setSavingDisplayTheme(false);
        }
    };

    const handleSaveSiteTheme = async () => {
        setSiteThemeError('');
        setSiteThemeSuccess('');
        try {
            setSavingSiteTheme(true);
            if (siteThemeImages.length) {
                await backend.uploadSiteThemeImages(siteThemeImages);
                setSiteThemeImages([]);
                setSiteThemeImageInputKey((current) => current + 1);
            }
            await backend.updateSiteTheme({ siteTheme, siteCustomColor: useCustomSiteColor ? siteCustomColor : '', siteThemeImageRotationMode, siteThemeImageRotationMinutes });
            await refreshUser();
            setSiteThemeSuccess('Site theme saved.');
        } catch (err: any) {
            setSiteThemeError(err?.details?.message || err?.message || 'Unable to save site theme.');
        } finally {
            setSavingSiteTheme(false);
        }
    };

    const handleRemoveSiteThemeImage = async (imagePath: string) => {
        setSiteThemeError('');
        setSiteThemeSuccess('');
        try {
            setSavingSiteTheme(true);
            await backend.removeSiteThemeImage(imagePath);
            await refreshUser();
            setSiteThemeSuccess('Site theme image removed.');
        } catch (err: any) {
            setSiteThemeError(err?.details?.message || err?.message || 'Unable to remove site theme image.');
        } finally {
            setSavingSiteTheme(false);
        }
    };

    const handleUploadThemeImage = async () => {
        if (!themeImage) return;
        setDisplayThemeError('');
        setDisplayThemeSuccess('');
        try {
            setSavingDisplayTheme(true);
            await backend.uploadProfileThemeImage(themeImage);
            setThemeImage(null);
            setThemeImageInputKey((current) => current + 1);
            await refreshUser();
            setDisplayThemeSuccess('Profile theme image saved.');
        } catch (err: any) {
            setDisplayThemeError(err?.details?.message || err?.message || 'Unable to upload the profile theme image.');
        } finally {
            setSavingDisplayTheme(false);
        }
    };

    const handleRemoveThemeImage = async () => {
        setDisplayThemeError('');
        setDisplayThemeSuccess('');
        try {
            setSavingDisplayTheme(true);
            await backend.removeProfileThemeImage();
            await refreshUser();
            setDisplayThemeSuccess('Profile theme image removed.');
        } catch (err: any) {
            setDisplayThemeError(err?.details?.message || err?.message || 'Unable to remove the profile theme image.');
        } finally {
            setSavingDisplayTheme(false);
        }
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
            <h2 className="mb-4">Settings</h2>

            <Card className="mb-4 profile-settings-card shadow-sm">
                <Card.Body className="p-4">
                    <div className="d-flex align-items-center gap-3 mb-4">
                        {profileImageUrl ? (
                            <img className="profile-avatar-image" src={profileImageUrl} alt={`${user.displayName || user.email}'s profile`} />
                        ) : (
                            <div className="profile-avatar-fallback" aria-label="Profile photo placeholder">{avatarInitial}</div>
                        )}
                        <div>
                            <h3 className="h5 mb-1">Profile</h3>
                            <p className="text-muted mb-0">Choose the name and photo shown on My Profile.</p>
                        </div>
                    </div>
                    {profileError && <Alert variant="danger">{profileError}</Alert>}
                    {profileSuccess && <Alert variant="success">{profileSuccess}</Alert>}
                    <Form onSubmit={handleSaveProfile}>
                        <Form.Group className="mb-3" controlId="display-name">
                            <Form.Label>Display name</Form.Label>
                            <Form.Control value={displayName} maxLength={80} onChange={(event) => setDisplayName(event.target.value)} required />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="profile-photo">
                            <Form.Label>Profile photo</Form.Label>
                            <Form.Control type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => setProfilePhoto((event.target as HTMLInputElement).files?.[0] ?? null)} />
                            <Form.Text className="text-muted">JPEG, PNG, or WebP; up to 5 MB.</Form.Text>
                        </Form.Group>
                        <Button type="submit" disabled={savingProfile}>
                            {savingProfile ? 'Saving profile...' : 'Save profile'}
                        </Button>
                    </Form>
                </Card.Body>
            </Card>

            <Card className={`mb-4 profile-settings-card profile-theme-preview profile-theme-${siteTheme}${siteThemeImageUrl ? ' theme-has-image' : ''} shadow-sm`} style={presentationThemeStyle(useCustomSiteColor ? siteCustomColor : undefined, siteThemeImageUrl)}>
                <Card.Body className="p-4">
                    <h3 className="h5 mb-1">Site theme</h3>
                    <p className="mb-3">Choose the palette for Home, manga browsing and search, and the chatroom browser. My Profile and an open chatroom keep their own themes.</p>
                    {siteThemeError && <Alert variant="danger">{siteThemeError}</Alert>}
                    {siteThemeSuccess && <Alert variant="success">{siteThemeSuccess}</Alert>}
                    <Form.Group className="mb-3" controlId="site-display-theme">
                        <Form.Label>Palette</Form.Label>
                        <Form.Select value={siteTheme} onChange={(event) => { setSiteTheme(event.target.value as DisplayTheme); setUseCustomSiteColor(false); }} disabled={savingSiteTheme}>
                            {displayThemeOptions.map((option) => <option key={option.value} value={option.value}>{option.label} — {option.description}</option>)}
                        </Form.Select>
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="site-custom-theme-color">
                        <Form.Label>Custom color</Form.Label>
                        <div className="d-flex align-items-center gap-2">
                            <Form.Control type="color" value={siteCustomColor} disabled={savingSiteTheme} onChange={(event) => { setSiteCustomColor(event.target.value); setUseCustomSiteColor(true); }} aria-label="Custom site theme color" />
                            <span className="font-monospace text-uppercase">{siteCustomColor}</span>
                        </div>
                        <Form.Text className="text-muted">Picking a color activates a custom palette for general site pages.</Form.Text>
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="site-theme-images">
                        <Form.Label>Site background images</Form.Label>
                        <div className="d-flex flex-wrap align-items-center gap-2">
                            <Form.Control key={siteThemeImageInputKey} className="theme-image-input" type="file" multiple accept="image/jpeg,image/png,image/webp,image/gif" disabled={savingSiteTheme} onChange={(event) => setSiteThemeImages(Array.from((event.target as HTMLInputElement).files || []))} />
                        </div>
                        <Form.Text className="text-muted">JPEG, PNG, WebP, or animated GIF up to 5 MB each; keep up to 10. GIFs play as backgrounds. A 1920 × 1080 image works best. Selected images are uploaded when you save the site theme.</Form.Text>
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="site-theme-image-rotation">
                        <Form.Label>Image rotation</Form.Label>
                        <Form.Select value={siteThemeImageRotationMode} disabled={savingSiteTheme} onChange={(event) => setSiteThemeImageRotationMode(event.target.value as SiteThemeImageRotationMode)}>
                            <option value="static">Static (first image)</option>
                            <option value="refresh">Choose a different image on every page refresh</option>
                            <option value="session">Choose one image per browser session</option>
                            <option value="interval">Rotate every X minutes</option>
                        </Form.Select>
                        {siteThemeImageRotationMode === 'interval' && <div className="d-flex align-items-center gap-2 mt-2">
                            <Form.Control type="number" min={1} max={1440} value={siteThemeImageRotationMinutes} disabled={savingSiteTheme} onChange={(event) => setSiteThemeImageRotationMinutes(Number(event.target.value))} aria-label="Site theme rotation interval in minutes" />
                            <span>minutes</span>
                        </div>}
                        <Form.Text className="text-muted">Rotation needs at least two saved images. Timed rotation accepts 1–1440 minutes; save the site theme to apply the choice.</Form.Text>
                    </Form.Group>
                    {siteThemeImagePaths.length > 0 && <div className="site-theme-image-list mb-3" aria-label="Saved site background images">
                        {siteThemeImagePaths.map((imagePath, index) => <div className="site-theme-image-item" key={imagePath}>
                            <img src={`${process.env.REACT_APP_API_BASE_URL ?? 'http://localhost:3001'}${imagePath}`} alt={`Site background ${index + 1}`} />
                            <Button variant="outline-secondary" size="sm" onClick={() => handleRemoveSiteThemeImage(imagePath)} disabled={savingSiteTheme}>Remove</Button>
                        </div>)}
                    </div>}
                    <div className="profile-theme-swatch mb-3" aria-label={`${siteTheme} site theme preview`}>
                        <span /><span /><span />
                    </div>
                    <Button onClick={handleSaveSiteTheme} disabled={savingSiteTheme}>{savingSiteTheme ? 'Saving theme...' : 'Save site theme'}</Button>
                </Card.Body>
            </Card>

            <Card className={`mb-4 profile-settings-card profile-theme-preview profile-theme-${displayTheme}${themeImageUrl ? ' theme-has-image' : ''} shadow-sm`} style={presentationThemeStyle(useCustomThemeColor ? customThemeColor : undefined, themeImageUrl)}>
                <Card.Body className="p-4">
                    <h3 className="h5 mb-1">My Profile display theme</h3>
                    <p className="mb-3">Choose the palette used on your My Profile page. It does not change the rest of the site.</p>
                    {displayThemeError && <Alert variant="danger">{displayThemeError}</Alert>}
                    {displayThemeSuccess && <Alert variant="success">{displayThemeSuccess}</Alert>}
                    <Form.Group className="mb-3" controlId="profile-display-theme">
                        <Form.Label>Palette</Form.Label>
                        <Form.Select value={displayTheme} onChange={(event) => { setDisplayTheme(event.target.value as DisplayTheme); setUseCustomThemeColor(false); }} disabled={savingDisplayTheme}>
                            {displayThemeOptions.map((option) => <option key={option.value} value={option.value}>{option.label} — {option.description}</option>)}
                        </Form.Select>
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="profile-custom-theme-color">
                        <Form.Label>Custom color</Form.Label>
                        <div className="d-flex align-items-center gap-2">
                            <Form.Control type="color" value={customThemeColor} disabled={savingDisplayTheme} onChange={(event) => { setCustomThemeColor(event.target.value); setUseCustomThemeColor(true); }} aria-label="Custom profile theme color" />
                            <span className="font-monospace text-uppercase">{customThemeColor}</span>
                        </div>
                        <Form.Text className="text-muted">Picking a color activates a custom palette; the app adjusts surfaces and text for readability.</Form.Text>
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="profile-theme-image">
                        <Form.Label>Theme background image</Form.Label>
                        <div className="d-flex flex-wrap align-items-center gap-2">
                            <Form.Control key={themeImageInputKey} className="theme-image-input" type="file" accept="image/jpeg,image/png,image/webp,image/gif" disabled={savingDisplayTheme} onChange={(event) => setThemeImage((event.target as HTMLInputElement).files?.[0] ?? null)} />
                            <Button variant="outline-primary" onClick={handleUploadThemeImage} disabled={!themeImage || savingDisplayTheme}>{savingDisplayTheme ? 'Uploading…' : 'Upload image'}</Button>
                            {themeImageUrl && <Button variant="outline-secondary" onClick={handleRemoveThemeImage} disabled={savingDisplayTheme}>Remove image</Button>}
                        </div>
                        <Form.Text className="text-muted">JPEG, PNG, WebP, or animated GIF up to 5 MB. GIFs play as backgrounds. A 1920 × 1080 image works best. Uploading replaces the current background.</Form.Text>
                    </Form.Group>
                    <div className="profile-theme-swatch mb-3" aria-label={`${displayTheme} theme preview`}>
                        <span /><span /><span />
                    </div>
                    <Button onClick={handleSaveDisplayTheme} disabled={savingDisplayTheme}>{savingDisplayTheme ? 'Saving theme...' : 'Save theme'}</Button>
                </Card.Body>
            </Card>

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
                    <Card.Title>Reading</Card.Title>
                    <Card.Text className="text-muted">
                        Choose the translation language and page direction used when you open a manga. You can still change a manga's language and scanlation group.
                    </Card.Text>
                    {readingSettingsError && <Alert variant="danger">{readingSettingsError}</Alert>}
                    {readingSettingsSuccess && <Alert variant="success">{readingSettingsSuccess}</Alert>}
                    <Form.Group className="mb-3">
                        <Form.Label htmlFor="preferred-translation-language">Preferred translation language</Form.Label>
                        <div className="d-flex gap-2">
                            <Form.Select
                                id="preferred-translation-language"
                                aria-label="Preferred translation language"
                                value={preferredTranslationLanguage}
                                disabled={loadingReadingSettings || savingReadingSettings}
                                onChange={(event) => setPreferredTranslationLanguage(event.target.value)}
                            >
                                {languageOptions.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                            </Form.Select>
                            <Button onClick={handleSaveReadingSettings} disabled={loadingReadingSettings || savingReadingSettings}>
                                {savingReadingSettings ? 'Saving...' : 'Save'}
                            </Button>
                        </div>
                    </Form.Group>
                    <Form.Group>
                        <Form.Label htmlFor="preferred-reading-mode">Reader layout</Form.Label>
                        <Form.Select
                            id="preferred-reading-mode"
                            aria-label="Reader layout"
                            value={preferredReadingMode}
                            disabled={loadingReadingSettings || savingReadingSettings}
                            onChange={(event) => setPreferredReadingMode(event.target.value as 'vertical' | 'ltr' | 'rtl')}
                        >
                            <option value="vertical">Continuous vertical</option>
                            <option value="ltr">Left to right pages</option>
                            <option value="rtl">Right to left pages</option>
                        </Form.Select>
                    </Form.Group>
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


