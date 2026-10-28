import { AniListMangaSearchParams } from '../../app/clients/backend/types';

export type ChatbotRequest = {
  kind: 'filters' | 'similar' | 'same-author';
  filters?: Omit<AniListMangaSearchParams, 'page' | 'perPage' | 'sort'>;
  title?: string;
  error?: string;
};

const genreAliases: Array<[string, string]> = [
  ['action', 'Action'], ['adventure', 'Adventure'], ['comedy', 'Comedy'], ['drama', 'Drama'],
  ['ecchi', 'Ecchi'], ['fantasy', 'Fantasy'], ['horror', 'Horror'], ['mahou shoujo', 'Mahou Shoujo'],
  ['magical girl', 'Mahou Shoujo'], ['mecha', 'Mecha'], ['music', 'Music'], ['mystery', 'Mystery'],
  ['psychological', 'Psychological'], ['romance', 'Romance'], ['sci-fi', 'Sci-Fi'], ['sci fi', 'Sci-Fi'],
  ['science fiction', 'Sci-Fi'], ['slice of life', 'Slice of Life'], ['sports', 'Sports'], ['supernatural', 'Supernatural'],
  ['thriller', 'Thriller'],
];

const includesPhrase = (text: string, phrase: string) =>
  new RegExp(`(^|[^a-z])${phrase.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}($|[^a-z])`, 'i').test(text);

const cleanReferencedTitle = (value: string) => value
  .replace(/^\s*(?:manga|manhwa|manhua)\s+/i, '')
  .replace(/[?.!]+$/, '')
  .trim();

export const parseChatbotRequest = (rawRequest: string): ChatbotRequest => {
  const request = rawRequest.trim();
  const normalized = request.toLowerCase();
  if (!request) return { kind: 'filters', error: 'Tell me what you would like to read first.' };

  const sameAuthor = request.match(/(?:same author as|by (?:the )?same author as)\s+(.+)$/i);
  if (sameAuthor) {
    const title = cleanReferencedTitle(sameAuthor[1]);
    return title
      ? { kind: 'same-author', title }
      : { kind: 'filters', error: 'Add the manga title after “same author as”.' };
  }

  const similar = request.match(/similar\s+to\s+(.+)$/i)
    || (normalized.startsWith('like ') ? [request, request.slice(5)] : null);
  if (similar) {
    const title = cleanReferencedTitle(similar[1]);
    return title
      ? { kind: 'similar', title }
      : { kind: 'filters', error: 'Add the manga title after “similar to”.' };
  }

  const genres = genreAliases
    .filter(([phrase]) => includesPhrase(normalized, phrase))
    .map(([, genre]) => genre)
    .filter((genre, index, values) => values.indexOf(genre) === index);
  const format = includesPhrase(normalized, 'one shot') || includesPhrase(normalized, 'one-shot') || includesPhrase(normalized, 'oneshot')
    ? 'ONE_SHOT'
    : includesPhrase(normalized, 'manga') || includesPhrase(normalized, 'series') ? 'MANGA' : '';
  const status = ['completed', 'complete', 'finished'].some((term) => includesPhrase(normalized, term))
    ? 'FINISHED'
    : ['ongoing', 'publishing', 'releasing'].some((term) => includesPhrase(normalized, term)) ? 'RELEASING' : '';

  if (!genres.length && !format && !status) {
    return {
      kind: 'filters',
      error: 'Try genres such as action or romance, manga or one-shot, and completed or ongoing. You can also ask for “similar to Berserk” or “same author as Berserk”.',
    };
  }

  return {
    kind: 'filters',
    filters: {
      ...(genres.length ? { genres } : {}),
      ...(format ? { format } : {}),
      ...(status ? { status } : {}),
    },
  };
};
