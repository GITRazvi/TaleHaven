import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Backend from '../../app/clients/backend/backend';
import { useAuth } from '../../contexts/AuthContext';
import HomePage from './home-page';

const mockNavigate = jest.fn();
const mockedBackend = Backend as unknown as jest.Mock;
const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

jest.mock('../../app/clients/backend/backend', () => jest.fn());
jest.mock('../../contexts/AuthContext', () => ({ useAuth: jest.fn() }));
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

const manga = {
  id: 'manga-1',
  title: 'Test Manga',
  coverUrl: '',
};

const secondManga = {
  id: 'manga-2',
  title: 'Second Test Manga',
  coverUrl: '',
};

const authenticatedUser = {
  id: 1,
  email: 'reader@example.com',
};

describe('HomePage library heart', () => {
  const getPopularManga = jest.fn();
  const getMyLibrary = jest.fn();
  const addMangaToMyLibrary = jest.fn();
  const removeMangaFromMyLibrary = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockedUseAuth.mockReturnValue({ user: authenticatedUser } as ReturnType<typeof useAuth>);
    getPopularManga.mockResolvedValue({ data: [manga] });
    getMyLibrary.mockResolvedValue({ data: [] });
    addMangaToMyLibrary.mockResolvedValue({ result: 'ok' });
    removeMangaFromMyLibrary.mockResolvedValue({ result: 'ok' });
    mockedBackend.mockImplementation(() => ({
      getPopularManga,
      getMyLibrary,
      addMangaToMyLibrary,
      removeMangaFromMyLibrary,
    }));
  });

  const renderPage = () => render(<HomePage />, { wrapper: MemoryRouter });

  it('adds a manga when its hollow heart is clicked', async () => {
    renderPage();

    const heart = await screen.findByRole('button', { name: 'Add Test Manga to My Library' });
    await userEvent.click(heart);

    expect(addMangaToMyLibrary).toHaveBeenCalledWith('manga-1');
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Remove Test Manga from My Library' })).toHaveAttribute('aria-pressed', 'true');
    });
    expect(removeMangaFromMyLibrary).not.toHaveBeenCalled();
  });

  it('removes a manga when its filled heart is clicked', async () => {
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'manga-1' }] });
    renderPage();

    const heart = await screen.findByRole('button', { name: 'Remove Test Manga from My Library' });
    await userEvent.click(heart);

    expect(removeMangaFromMyLibrary).toHaveBeenCalledWith('manga-1');
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Add Test Manga to My Library' })).toHaveAttribute('aria-pressed', 'false');
    });
    expect(addMangaToMyLibrary).not.toHaveBeenCalled();
  });

  it('keeps a hollow heart and shows an error when adding fails', async () => {
    addMangaToMyLibrary.mockRejectedValue(new Error('Unable to save manga'));
    renderPage();

    const heart = await screen.findByRole('button', { name: 'Add Test Manga to My Library' });
    await userEvent.click(heart);

    expect(addMangaToMyLibrary).toHaveBeenCalledWith('manga-1');
    expect(await screen.findByText('Unable to save manga')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Test Manga to My Library' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('keeps a filled heart and shows an error when removing fails', async () => {
    getMyLibrary.mockResolvedValue({ data: [{ mangaId: 'manga-1' }] });
    removeMangaFromMyLibrary.mockRejectedValue(new Error('Unable to remove manga'));
    renderPage();

    const heart = await screen.findByRole('button', { name: 'Remove Test Manga from My Library' });
    await userEvent.click(heart);

    expect(removeMangaFromMyLibrary).toHaveBeenCalledWith('manga-1');
    expect(await screen.findByText('Unable to remove manga')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Remove Test Manga from My Library' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('requests and appends another popular-manga batch when Load more is clicked', async () => {
    getPopularManga
      .mockResolvedValueOnce({ data: [manga], total: 2, limit: 20, offset: 0 })
      .mockResolvedValueOnce({ data: [secondManga], total: 2, limit: 20, offset: 1 });
    renderPage();

    expect(await screen.findByText('Test Manga')).toBeInTheDocument();
    expect(screen.queryByText('Second Test Manga')).not.toBeInTheDocument();
    expect(getPopularManga).toHaveBeenCalledWith({ limit: 20, offset: 0, orderBy: 'followedCount' });

    await userEvent.click(screen.getByRole('button', { name: 'Load more popular manga' }));

    await waitFor(() => {
      expect(getPopularManga).toHaveBeenLastCalledWith({ limit: 20, offset: 1, orderBy: 'followedCount' });
      expect(screen.getByText('Second Test Manga')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: 'Load more popular manga' })).not.toBeInTheDocument();
  });
});
