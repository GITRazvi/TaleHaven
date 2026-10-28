import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ChatRoomPage from './chat-room-page';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';

const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));

describe('ChatRoomPage', () => {
    beforeEach(() => {
        mockedUseAuth.mockReturnValue({ user: { id: 1, email: 'reader@example.test', created_at: '' } } as ReturnType<typeof useAuth>);
    });

    it('shows a member room and its messages', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 1, name: 'Fantasy talk', visibility: 'invite_only', imagePath: '/images/chat/rooms/fantasy.jpg', theme: 'midnight', customThemeColor: '', ownerUserId: 1, ownerDisplayName: 'Reader', memberCount: 2, isMember: true, isOwner: true, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [{ id: 1, roomId: 1, userId: 2, authorDisplayName: 'Friend', body: 'Welcome!', imagePath: '/images/chat/messages/welcome.jpg', createdAt: '2026-10-27T00:00:00Z' }] });
        const updateChatRoomVisibility = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 1, name: 'Fantasy talk', visibility: 'public', imagePath: '/images/chat/rooms/fantasy.jpg', theme: 'midnight', customThemeColor: '', ownerUserId: 1, ownerDisplayName: 'Reader', memberCount: 2, isMember: true, isOwner: true, createdAt: '2026-10-27T00:00:00Z' } });
        const updateChatRoomTheme = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 1, name: 'Fantasy talk', visibility: 'public', imagePath: '/images/chat/rooms/fantasy.jpg', theme: 'forest', customThemeColor: '', ownerUserId: 1, ownerDisplayName: 'Reader', memberCount: 2, isMember: true, isOwner: true, createdAt: '2026-10-27T00:00:00Z' } });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, updateChatRoomVisibility, updateChatRoomTheme, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/1']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        expect(await screen.findByRole('heading', { name: 'Fantasy talk' })).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Cover for Fantasy talk' })).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Shared by Friend' })).toBeInTheDocument();
        expect(screen.getByText('Welcome!')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Room settings' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Make public' })).not.toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Room settings' }));
        expect(screen.getByRole('button', { name: 'Make public' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Create invite link' })).toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Make public' }));

        expect(updateChatRoomVisibility).toHaveBeenCalledWith(1, 'public');
        expect(await screen.findByRole('button', { name: 'Make private' })).toBeInTheDocument();

        await userEvent.selectOptions(screen.getByLabelText('Chatroom theme'), 'forest');
        await userEvent.click(screen.getByRole('button', { name: 'Save theme' }));
        expect(updateChatRoomTheme).toHaveBeenCalledWith(1, 'forest', '');
    });

    it('does not show room settings to a member who is not the creator', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 2, name: 'Readers', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 2, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [] });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/2']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        expect(await screen.findByRole('heading', { name: 'Readers' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Room settings' })).not.toBeInTheDocument();
        expect(screen.queryByRole('heading', { name: 'Room settings' })).not.toBeInTheDocument();
    });

    it('attaches an image pasted into the message box without rendering a file chooser', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 3, name: 'Paste room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 1, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [] });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/3']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        const messageBox = await screen.findByLabelText('Message');
        const image = new File(['image'], 'clipboard.png', { type: 'image/png' });
        fireEvent.paste(messageBox, { clipboardData: { items: [{ kind: 'file', type: 'image/png', getAsFile: () => image }] } });

        expect(screen.getByText('Pasted image attached: clipboard.png')).toBeInTheDocument();
        expect(screen.queryByLabelText('Attach image')).not.toBeInTheDocument();
    });

    it('sends a message when Enter is pressed but keeps Shift+Enter as a new line', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 4, name: 'Keyboard room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 1, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [] });
        const sendChatMessage = jest.fn().mockResolvedValue({ result: 'ok', message: { id: 1, roomId: 4, userId: 1, authorDisplayName: 'Creator', body: 'Sent with Enter', imagePath: '', createdAt: '2026-10-27T00:00:00Z' } });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, sendChatMessage, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/4']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        const messageBox = await screen.findByLabelText('Message');
        await userEvent.type(messageBox, 'First line{shift>}{enter}{/shift}Second line');
        expect(messageBox).toHaveValue('First line\nSecond line');
        expect(sendChatMessage).not.toHaveBeenCalled();

        await userEvent.clear(messageBox);
        await userEvent.type(messageBox, 'Sent with Enter{enter}');
        expect(sendChatMessage).toHaveBeenCalledWith(4, 'Sent with Enter', 0);
        expect(await screen.findByText('Sent with Enter')).toBeInTheDocument();
    });

    it('opens at the newest message and resumes following when the reader returns there', async () => {
        const originalScrollHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'scrollHeight');
        const originalClientHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'clientHeight');
        Object.defineProperty(HTMLElement.prototype, 'scrollHeight', { configurable: true, get: () => 900 });
        Object.defineProperty(HTMLElement.prototype, 'clientHeight', { configurable: true, get: () => 400 });
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 5, name: 'Follow room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 1, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [{ id: 1, roomId: 5, userId: 2, authorDisplayName: 'Friend', body: 'Latest message', imagePath: '', createdAt: '2026-10-27T00:00:00Z' }] });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, getBaseUrl: () => 'http://localhost:3001' }));

        try {
            render(<MemoryRouter initialEntries={['/chat/rooms/5']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

            await screen.findByRole('heading', { name: 'Follow room' });
            const messageList = document.querySelector('.chat-messages-list') as HTMLDivElement;
            expect(messageList.scrollTop).toBe(900);

            messageList.scrollTop = 100;
            fireEvent.scroll(messageList);
            expect(messageList.scrollTop).toBe(100);

            messageList.scrollTop = 500;
            fireEvent.scroll(messageList);
            expect(messageList.scrollTop).toBe(900);
        } finally {
            if (originalScrollHeight) Object.defineProperty(HTMLElement.prototype, 'scrollHeight', originalScrollHeight);
            else delete (HTMLElement.prototype as { scrollHeight?: number }).scrollHeight;
            if (originalClientHeight) Object.defineProperty(HTMLElement.prototype, 'clientHeight', originalClientHeight);
            else delete (HTMLElement.prototype as { clientHeight?: number }).clientHeight;
        }
    });

    it('shows delete only for the signed-in user while other messages offer reporting', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 6, name: 'Actions room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 2, isMember: true, isOwner: true, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [{ id: 1, roomId: 6, userId: 1, authorDisplayName: 'Reader', body: 'My message', imagePath: '', createdAt: '2026-10-27T00:00:00Z' }, { id: 2, roomId: 6, userId: 2, authorDisplayName: 'Friend', body: 'Friend message', imagePath: '', createdAt: '2026-10-27T00:00:00Z' }] });
        const deleteChatMessage = jest.fn().mockResolvedValue({ result: 'ok', message: 'Message deleted' });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, deleteChatMessage, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/6']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        expect(await screen.findByText('My message')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Message actions for My message' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Message actions for Friend message' })).toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Message actions for My message' }));
        await userEvent.click(screen.getByText('Delete'));

        expect(deleteChatMessage).toHaveBeenCalledWith(6, 1);
        await waitFor(() => expect(screen.queryByText('My message')).not.toBeInTheDocument());
        expect(screen.getByText('Friend message')).toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Message actions for Friend message' }));
        expect(screen.getByText('Reply')).toBeInTheDocument();
        expect(screen.getByText('Report')).toBeInTheDocument();
    });

    it('jumps to the original message when a reply reference is clicked', async () => {
        const scrollIntoView = jest.fn();
        Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', { configurable: true, value: scrollIntoView });
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 7, name: 'Reply room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 2, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [{ id: 10, roomId: 7, userId: 2, authorDisplayName: 'Friend', body: 'Original message', imagePath: '', replyToMessageId: 0, createdAt: '2026-10-27T00:00:00Z' }, { id: 11, roomId: 7, userId: 1, authorDisplayName: 'Reader', body: 'A reply', imagePath: '', replyToMessageId: 10, createdAt: '2026-10-27T00:01:00Z' }] });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/7']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        await userEvent.click(await screen.findByRole('button', { name: 'View replied message from Friend' }));
        expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'center' });
    });

    it('shows an author profile preview when a message name is hovered', async () => {
        const getChatRoom = jest.fn().mockResolvedValue({ result: 'ok', room: { id: 8, name: 'Preview room', visibility: 'public', imagePath: '', theme: 'midnight', customThemeColor: '', themeImagePath: '', ownerUserId: 1, ownerDisplayName: 'Creator', memberCount: 2, isMember: true, isOwner: false, createdAt: '2026-10-27T00:00:00Z' } });
        const getChatMessages = jest.fn().mockResolvedValue({ result: 'ok', limit: 100, data: [{ id: 12, roomId: 8, userId: 2, authorDisplayName: 'Friend', profileImagePath: '/images/profiles/friend.jpg', body: 'Hello', imagePath: '', replyToMessageId: 0, createdAt: '2026-10-27T00:00:00Z' }] });
        mockedBackend.mockImplementation(() => ({ getChatRoom, getChatMessages, getBaseUrl: () => 'http://localhost:3001' }));

        render(<MemoryRouter initialEntries={['/chat/rooms/8']}><Routes><Route path="/chat/rooms/:roomID" element={<ChatRoomPage />} /></Routes></MemoryRouter>);

        await userEvent.hover(await screen.findByRole('button', { name: 'Friend' }));
        const profileImage = await screen.findByRole('img', { name: "Friend's profile" });
        expect(profileImage).toHaveAttribute('src', 'http://localhost:3001/images/profiles/friend.jpg');
        expect(screen.getAllByText('Friend').length).toBeGreaterThan(1);
    });
});
