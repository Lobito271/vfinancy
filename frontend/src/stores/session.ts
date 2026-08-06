import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Role } from '@/constants/permissions';

export interface SessionUser {
  id: string;
  fullName: string;
  email: string;
  username: string;
  roles: Role[];
  company: string;
}

interface SessionState {
  user: SessionUser | null;
  token: string | null;
  expiresAt: string | null;
  isAuthenticated: boolean;
  lastUsername: string;
  setUser: (user: SessionUser, token: string, expiresAt: string) => void;
  setLastUsername: (username: string) => void;
  logout: () => void;
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      expiresAt: null,
      isAuthenticated: false,
      lastUsername: '',
      setUser: (user, token, expiresAt) =>
        set({ user, token, expiresAt, isAuthenticated: true, lastUsername: user.username }),
      setLastUsername: (username) => set({ lastUsername: username }),
      logout: () =>
        set({ user: null, token: null, expiresAt: null, isAuthenticated: false }),
    }),
    { name: 'vfinancy.session' },
  ),
);
