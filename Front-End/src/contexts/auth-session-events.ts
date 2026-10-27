export const sessionExpiredEvent = 'talehaven:session-expired';

// Backend client instances are created outside React. A browser event lets the
// auth provider clear its user state when any protected request returns 401.
export const notifySessionExpired = (): void => {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event(sessionExpiredEvent));
  }
};
