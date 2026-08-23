import { createPortal } from 'react-dom';
import { CheckCircle2, AlertTriangle, Info, XCircle, X } from 'lucide-react';
import { cx } from '@/utils/cx';
import { useNotificationStore, type ToastVariant } from '@/stores/notification';

const iconMap: Record<ToastVariant, typeof CheckCircle2> = {
  success: CheckCircle2,
  warning: AlertTriangle,
  destructive: XCircle,
  info: Info,
};

export function Toaster() {
  const toasts = useNotificationStore((s) => s.toasts);
  const dismiss = useNotificationStore((s) => s.dismiss);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <div
      className="toaster"
      aria-live="polite"
      aria-atomic="true"
    >
      {toasts.map((t) => {
        const Icon = iconMap[t.variant];
        return (
          <div
            key={t.id}
            role="status"
            className={cx('toast', `toast--${t.variant}`)}
          >
            <div className="toast__bar" aria-hidden="true" />
            <Icon className="toast__icon" aria-hidden="true" />
            <div className="toast__body">
              <p className="toast__title">{t.title}</p>
              {t.description && (
                <p className="toast__description">{t.description}</p>
              )}
            </div>
            <button
              type="button"
              onClick={() => dismiss(t.id)}
              className="toast__close"
              aria-label="Cerrar notificación"
            >
              <X />
            </button>
          </div>
        );
      })}
    </div>,
    document.body,
  );
}
