import { Toast } from '@base-ui/react/toast';
import { CheckCircle2, AlertTriangle, Info, XCircle, X } from 'lucide-react';
import { cx } from '@/utils/cx';
import { toastManager, type ToastVariant } from '@/stores/notification';

const iconMap: Record<ToastVariant, typeof CheckCircle2> = {
  success: CheckCircle2,
  warning: AlertTriangle,
  destructive: XCircle,
  info: Info,
};

function ToastViewport() {
  const manager = Toast.useToastManager();

  return (
    <>
      {manager.toasts.map((toast) => {
        const variant = (toast.type ?? 'info') as ToastVariant;
        const Icon = iconMap[variant] ?? Info;
        return (
          <Toast.Root
            key={toast.id}
            toast={toast}
            className={cx('toast', `toast--${variant}`)}
            swipeDirection={['down', 'right']}
          >
            <span className="toast__icon" aria-hidden="true">
              <Icon />
            </span>
            <div className="toast__body">
              <Toast.Title className="toast__title">{toast.title}</Toast.Title>
              {toast.description && (
                <Toast.Description className="toast__description">{toast.description}</Toast.Description>
              )}
            </div>
            <Toast.Close className="toast__close" aria-label="Cerrar notificación">
              <X />
            </Toast.Close>
          </Toast.Root>
        );
      })}
      <Toast.Viewport className="toaster" />
    </>
  );
}

export function Toaster() {
  return (
    <Toast.Provider toastManager={toastManager}>
      <ToastViewport />
    </Toast.Provider>
  );
}
