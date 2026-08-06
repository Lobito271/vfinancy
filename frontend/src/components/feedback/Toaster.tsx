import { createPortal } from 'react-dom';
import { CheckCircle2, AlertTriangle, Info, XCircle, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useNotificationStore, type ToastVariant } from '@/stores/notification';

const iconMap: Record<ToastVariant, typeof CheckCircle2> = {
  success: CheckCircle2,
  warning: AlertTriangle,
  destructive: XCircle,
  info: Info,
};

const styleMap: Record<ToastVariant, { bar: string; icon: string }> = {
  success: { bar: 'bg-success', icon: 'text-success' },
  warning: { bar: 'bg-warning', icon: 'text-warning' },
  destructive: { bar: 'bg-destructive', icon: 'text-destructive' },
  info: { bar: 'bg-info', icon: 'text-info' },
};

export function Toaster() {
  const toasts = useNotificationStore((s) => s.toasts);
  const dismiss = useNotificationStore((s) => s.dismiss);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-full max-w-sm flex-col gap-2"
      aria-live="polite"
      aria-atomic="true"
    >
      {toasts.map((t) => {
        const Icon = iconMap[t.variant];
        const style = styleMap[t.variant];
        return (
          <div
            key={t.id}
            role="status"
            className={cn(
              'pointer-events-auto relative flex w-full items-start gap-3 overflow-hidden rounded-lg border bg-card p-4 shadow-lg',
              'animate-in slide-in-from-right-full',
            )}
          >
            <div className={cn('absolute left-0 top-0 h-full w-1', style.bar)} aria-hidden="true" />
            <Icon className={cn('mt-0.5 h-5 w-5 shrink-0', style.icon)} aria-hidden="true" />
            <div className="flex-1 space-y-1 pl-2">
              <p className="text-sm font-semibold">{t.title}</p>
              {t.description && (
                <p className="text-sm text-muted-foreground">{t.description}</p>
              )}
            </div>
            <button
              type="button"
              onClick={() => dismiss(t.id)}
              className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              aria-label="Cerrar notificación"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        );
      })}
    </div>,
    document.body,
  );
}
