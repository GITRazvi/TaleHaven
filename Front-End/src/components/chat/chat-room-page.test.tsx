import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ChatRoomPage from './chat-room-page';
import Backend from '../../app/clients/backend/backend';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('ChatRoomPage', () => {
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
        expect(screen.getByRole('button', { name: 'Make public' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Create invite link' })).toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Make public' }));

        expect(updateChatRoomVisibility).toHaveBeenCalledWith(1, 'public');
        expect(await screen.findByRole('button', { name: 'Make private' })).toBeInTheDocument();

        await userEvent.selectOptions(screen.getByLabelText('Chatroom theme'), 'forest');
        await userEvent.click(screen.getByRole('button', { name: 'Save theme' }));
        expect(updateChatRoomTheme).toHaveBeenCalledWith(1, 'forest', '');
    });
});
