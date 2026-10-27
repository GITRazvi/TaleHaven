import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ChatRoomsPage from './chat-rooms-page';
import Backend from '../../app/clients/backend/backend';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('ChatRoomsPage', () => {
    it('shows public rooms and only private rooms the reader belongs to', async () => {
        const getChatRooms = jest.fn().mockResolvedValue({ result: 'ok', data: [
            { id: 1, name: 'Public readers', visibility: 'public', imagePath: '/images/chat/rooms/public.jpg', theme: 'forest', customThemeColor: '#387c5d', ownerUserId: 2, ownerDisplayName: 'Host', memberCount: 4, isMember: false, isOwner: false, createdAt: '2026-10-27T00:00:00Z' },
            { id: 2, name: 'Private club', visibility: 'invite_only', imagePath: '', theme: 'midnight', customThemeColor: '', ownerUserId: 1, ownerDisplayName: 'Reader', memberCount: 2, isMember: true, isOwner: true, createdAt: '2026-10-27T00:00:00Z' },
        ] });
        mockedBackend.mockImplementation(() => ({ getChatRooms, getBaseUrl: () => 'http://localhost:3001' }));

        render(<ChatRoomsPage />, { wrapper: MemoryRouter });

        expect(await screen.findByRole('heading', { name: 'Public readers' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Join room' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Your private rooms' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Private club' })).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Cover for Public readers' })).toBeInTheDocument();
        expect(getChatRooms).toHaveBeenCalledTimes(1);
    });
});
