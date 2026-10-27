export type BrowserCacheEntry<T> = {
  savedAt: number;
  value: T;
};

export const readBrowserCache = <T>(key: string, maxAgeMs: number): T | null => {
  if (typeof window === "undefined") return null;

  try {
    const serialized = window.localStorage.getItem(key);
    if (!serialized) return null;

    const entry = JSON.parse(serialized) as BrowserCacheEntry<T>;
    if (
      !entry ||
      typeof entry.savedAt !== "number" ||
      Date.now() - entry.savedAt > maxAgeMs
    ) {
      window.localStorage.removeItem(key);
      return null;
    }

    return entry.value;
  } catch {
    return null;
  }
};

export const writeBrowserCache = <T>(key: string, value: T) => {
  if (typeof window === "undefined") return;

  try {
    const entry: BrowserCacheEntry<T> = { savedAt: Date.now(), value };
    window.localStorage.setItem(key, JSON.stringify(entry));
  } catch {
    // Browser storage is only an optional resilience layer.
  }
};
