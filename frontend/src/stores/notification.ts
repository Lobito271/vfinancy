import { create } from 'zustand';
import { Toast } from '@base-ui/react/toast';

export type ToastVariant = 'success' | 'info' | 'warning' | 'destructive';

interface ToastInput {
  title: string;
  description?: string;
  variant?: ToastVariant;
  duration?: number;
}

const manager = Toast.createToastManager();

export const toastManager = manager;

interface NotificationState {
  push: (toast: ToastInput) => string;
  dismiss: (id: string) => void;
  clear?: () => void;
}

export const useNotificationStore = create<NotificationState>(() => ({
  push: (toast) =>
    manager.add({
      title: toast.title,
      description: toast.description,
      type: toast.variant ?? 'info',
      timeout: toast.duration ?? 5000,
    }),
  dismiss: (id) => manager.close(id),
}));
