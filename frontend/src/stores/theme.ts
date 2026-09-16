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
        const order: Theme[] = ['light', 'dark', 'system'];
        const next = order[(order.indexOf(get().theme) + 1) % order.length];
        const r = resolve(next);
        applyClass(r);
        set({ theme: next, resolved: r });
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
