import { create } from 'zustand';

interface UIState {
  globalSearch: string;
  setGlobalSearch: (s: string) => void;
}

export const useUIStore = create<UIState>((set) => ({
  globalSearch: '',
  setGlobalSearch: (globalSearch) => set({ globalSearch }),
}));
