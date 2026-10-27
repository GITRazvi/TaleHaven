import { CSSProperties } from 'react';

type ThemeStyle = CSSProperties & Record<`--${string}`, string>;

const validHexColor = /^#[0-9a-f]{6}$/i;

const toRgb = (hex: string) => [
  Number.parseInt(hex.slice(1, 3), 16),
  Number.parseInt(hex.slice(3, 5), 16),
  Number.parseInt(hex.slice(5, 7), 16),
] as const;

const toHex = (red: number, green: number, blue: number) =>
  `#${[red, green, blue].map((value) => Math.round(Math.max(0, Math.min(255, value))).toString(16).padStart(2, '0')).join('')}`;

const blend = (source: readonly number[], target: readonly number[], targetAmount: number) =>
  toHex(
    source[0] * (1 - targetAmount) + target[0] * targetAmount,
    source[1] * (1 - targetAmount) + target[1] * targetAmount,
    source[2] * (1 - targetAmount) + target[2] * targetAmount,
  );

const luminance = ([red, green, blue]: readonly number[]) =>
  (0.2126 * red + 0.7152 * green + 0.0722 * blue) / 255;

// A selected base color is never used directly for every surface. Instead we
// derive a restrained page/surface set and an appropriate text color.
export const customThemeStyle = (customThemeColor?: string): ThemeStyle | undefined => {
  if (!customThemeColor || !validHexColor.test(customThemeColor)) return undefined;
  const base = toRgb(customThemeColor);
  const isLight = luminance(base) > 0.56;
  const light = [255, 255, 255] as const;
  const dark = [12, 16, 24] as const;
  const text = isLight ? '#211b18' : '#f8fbff';
  const muted = isLight ? '#594e47' : '#d3dce9';
  return {
    '--theme-page': blend(base, isLight ? light : dark, isLight ? 0.18 : 0.48),
    '--theme-surface': blend(base, isLight ? light : dark, isLight ? 0.54 : 0.30),
    '--theme-surface-strong': blend(base, isLight ? light : dark, isLight ? 0.33 : 0.16),
    '--theme-border': blend(base, isLight ? dark : light, isLight ? 0.24 : 0.34),
    '--theme-text': text,
    '--theme-muted': muted,
    '--theme-accent': customThemeColor.toLowerCase(),
  };
};

// Theme images use server-generated local URLs. The overlay preserves readable
// content even when a reader uploads a bright or high-contrast image.
export const themeBackgroundStyle = (imageUrl?: string): ThemeStyle | undefined => {
  if (!imageUrl) return undefined;
  const safeUrl = imageUrl.replace(/["\\\n\r]/g, '');
  return {
    backgroundImage: `linear-gradient(145deg, rgba(7, 12, 22, 0.80), rgba(7, 12, 22, 0.66)), url("${safeUrl}")`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
  };
};

export const presentationThemeStyle = (customThemeColor?: string, imageUrl?: string): ThemeStyle | undefined => {
  const colorStyle = customThemeStyle(customThemeColor);
  const imageStyle = themeBackgroundStyle(imageUrl);
  if (!colorStyle && !imageStyle) return undefined;
  return { ...colorStyle, ...imageStyle };
};
