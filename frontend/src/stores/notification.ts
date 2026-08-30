import { create } from 'zustand';

export type ToastVariant = 'success' | 'info' | 'warning' | 'destructive';

interface Toast {
  id: string;
  title: string;
  description?: string;
  variant: ToastVariant;
  duration?: number;
}

interface NotificationState {
  toasts: Toast[];
  push: (t: Omit<Toast, 'id'>) => string;
  dismiss: (id: string) => void;
  clear: () => void;
}

export const useNotificationStore = create<NotificationState>()((set) => ({
  toasts: [],
  push: (t) => {
    const id = crypto.randomUUID();
    set((s) => ({ toasts: [...s.toasts, { ...t, id }] }));
    return id;
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((x) => x.id !== id) })),
  clear: () => set({ toasts: [] }),
}));
