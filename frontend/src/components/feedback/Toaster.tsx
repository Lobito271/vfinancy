import { createPortal } from 'react-dom';
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

  if (typeof document === 'undefined') return null;

  return createPortal(
    <Toast.Viewport
      style={{
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        zIndex: 99999,
        margin: 0,
        maxHeight: '100vh',
        display: 'flex',
        flexDirection: 'column-reverse',
        gap: '8px',
        padding: '16px',
        width: '100%',
        maxWidth: '420px',
        pointerEvents: 'none',
      }}
    >
      {manager.toasts.map((toast) => {
        const variant = (toast.type ?? 'info') as ToastVariant;
        const Icon = iconMap[variant] ?? Info;
        return (
          <Toast.Root
            key={toast.id}
            toast={toast}
            className={cx('toast', `toast--${variant}`)}
            style={{
              flexShrink: 0,
              flexGrow: 0,
              alignSelf: 'flex-end',
              width: '100%',
              maxHeight: '100px',
              height: 'auto',
            }}
            swipeDirection={['down', 'right']}
          >
            <span className="toast__icon" aria-hidden="true">
              <Icon strokeWidth={2.5} />
            </span>
            <div className="toast__body">
              <Toast.Title className="toast__title">{toast.title}</Toast.Title>
              {toast.description && (
                <Toast.Description className="toast__description">{toast.description}</Toast.Description>
              )}
            </div>
            <Toast.Close className="toast__close" aria-label="Cerrar notificación">
              <X strokeWidth={2.5} />
            </Toast.Close>
          </Toast.Root>
        );
      })}
    </Toast.Viewport>,
    document.body,
  );
}

export function Toaster() {
  return (
    <Toast.Provider toastManager={toastManager}>
      <ToastViewport />
    </Toast.Provider>
  );
}
