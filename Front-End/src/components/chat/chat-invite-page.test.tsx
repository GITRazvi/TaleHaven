import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ChatInvitePage from './chat-invite-page';
import Backend from '../../app/clients/backend/backend';

const mockedBackend = Backend as unknown as jest.Mock;

jest.mock('../../app/clients/backend/backend', () => jest.fn());

describe('ChatInvitePage', () => {
    it('redeems an invite then opens the room', async () => {
        const joinChatInvite = jest.fn().mockResolvedValue({ result: 'ok', joined: true, roomId: 7 });
        mockedBackend.mockImplementation(() => ({ joinChatInvite }));

        render(<MemoryRouter initialEntries={['/chat/invite/safe-token']}><Routes><Route path="/chat/invite/:token" element={<ChatInvitePage />} /><Route path="/chat/rooms/:roomID" element={<h1>Joined room</h1>} /></Routes></MemoryRouter>);

        expect(await screen.findByRole('heading', { name: 'Joined room' })).toBeInTheDocument();
        expect(joinChatInvite).toHaveBeenCalledWith('safe-token');
    });
});
