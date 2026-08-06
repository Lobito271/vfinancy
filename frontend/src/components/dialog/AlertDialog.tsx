import { AlertTriangle, CheckCircle2, Info, XCircle, type LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './Dialog';
import { Button } from '@/components/button';

type Variant = 'success' | 'warning' | 'destructive' | 'info' | 'confirmation';

const variantConfig: Record<
  Variant,
  { icon: LucideIcon; iconClass: string; bgClass: string }
> = {
  success: {
    icon: CheckCircle2,
    iconClass: 'text-success',
    bgClass: 'bg-success/10',
  },
  warning: {
    icon: AlertTriangle,
    iconClass: 'text-warning',
    bgClass: 'bg-warning/10',
  },
  destructive: {
    icon: XCircle,
    iconClass: 'text-destructive',
    bgClass: 'bg-destructive/10',
  },
  info: {
    icon: Info,
    iconClass: 'text-info',
    bgClass: 'bg-info/10',
  },
  confirmation: {
    icon: AlertTriangle,
    iconClass: 'text-warning',
    bgClass: 'bg-warning/10',
  },
};

export interface AlertDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  variant?: Variant;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm?: () => void;
  loading?: boolean;
  size?: 'sm' | 'md';
}

export function AlertDialog({
  open,
  onOpenChange,
  variant = 'info',
  title,
  description,
  confirmLabel = 'Aceptar',
  cancelLabel = 'Cancelar',
  onConfirm,
  loading = false,
  size = 'sm',
}: AlertDialogProps) {
  const cfg = variantConfig[variant];
  const Icon = cfg.icon;
  const isDestructive = variant === 'destructive';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size={size} className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-start gap-3">
            <div className={cn('flex h-10 w-10 shrink-0 items-center justify-center rounded-full', cfg.bgClass)}>
              <Icon className={cn('h-5 w-5', cfg.iconClass)} aria-hidden="true" />
            </div>
            <div className="flex-1 space-y-1.5">
              <DialogTitle>{title}</DialogTitle>
              {description && <DialogDescription>{description}</DialogDescription>}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter className="sm:gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            variant={isDestructive ? 'destructive' : 'default'}
            onClick={() => {
              onConfirm?.();
            }}
            loading={loading}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  description?: string;
  onConfirm: () => void;
  loading?: boolean;
  confirmLabel?: string;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title = '¿Confirmar acción?',
  description = '¿Está seguro que desea continuar?',
  onConfirm,
  loading,
  confirmLabel = 'Confirmar',
}: ConfirmDialogProps) {
  return (
    <AlertDialog
      open={open}
      onOpenChange={onOpenChange}
      variant="confirmation"
      title={title}
      description={description}
      confirmLabel={confirmLabel}
      onConfirm={onConfirm}
      loading={loading}
    />
  );
}
