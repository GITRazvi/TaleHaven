import { parseChatbotRequest } from './chatbot-request';

describe('parseChatbotRequest', () => {
  it('recognizes multiple genres, format, and publication status', () => {
    expect(parseChatbotRequest('I want an action romance one-shot that is completed')).toEqual({
      kind: 'filters',
      filters: { genres: ['Action', 'Romance'], format: 'ONE_SHOT', status: 'FINISHED' },
    });
  });

  it('recognizes similar-title and same-author requests before generic keywords', () => {
    expect(parseChatbotRequest('similar to manga Berserk')).toEqual({ kind: 'similar', title: 'Berserk' });
    expect(parseChatbotRequest('same author as Berserk')).toEqual({ kind: 'same-author', title: 'Berserk' });
  });
});
