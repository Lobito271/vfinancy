import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type Theme = 'light' | 'dark' | 'system';

interface ThemeState {
  theme: Theme;
  resolved: 'light' | 'dark';
  setTheme: (t: Theme) => void;
  toggle: () => void;
  applyToDocument: () => void;
}

function resolve(theme: Theme): 'light' | 'dark' {
  if (theme === 'system') {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return theme;
}

function applyClass(resolved: 'light' | 'dark') {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  root.classList.toggle('dark', resolved === 'dark');
  root.style.colorScheme = resolved;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'system',
      resolved: 'light',
      setTheme: (t) => {
        const r = resolve(t);
        applyClass(r);
        set({ theme: t, resolved: r });
      },
      toggle: () => {
        const next = get().resolved === 'dark' ? 'light' : 'dark';
        applyClass(next);
        set({ theme: next, resolved: next });
      },
      applyToDocument: () => {
        applyClass(get().resolved);
      },
    }),
    {
      name: 'vfinancy.theme',
      storage: createJSONStorage(() => localStorage),
      partialize: (s) => ({ theme: s.theme }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.resolved = resolve(state.theme);
          state.applyToDocument();
        }
      },
    },
  ),
);
