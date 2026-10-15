import { modeStorageKey, setMode } from 'mode-watcher';

export type ThemePreference = 'system' | 'light' | 'dark';

const preferenceStorageKey = 'hister-theme-preference';

function isThemePreference(value: string | null): value is ThemePreference {
  return value === 'system' || value === 'light' || value === 'dark';
}

export function getStoredThemePreference(): ThemePreference | null {
  const preference = localStorage.getItem(preferenceStorageKey);
  if (isThemePreference(preference)) return preference;
  if (preference !== null) return null;

  return migrateThemePreference();
}

export function setThemePreference(preference: ThemePreference): void {
  localStorage.setItem(preferenceStorageKey, preference);
  setMode(preference);
}

function migrateThemePreference(): ThemePreference | null {
  const legacyMode = localStorage.getItem(modeStorageKey.current);
  if (['light', 'dark'].includes(legacyMode ?? '')) {
    localStorage.setItem(preferenceStorageKey, legacyMode!);
    return legacyMode as ThemePreference;
  }

  // Record the migration without creating an explicit visitor preference.
  localStorage.setItem(preferenceStorageKey, '');
  return null;
}
