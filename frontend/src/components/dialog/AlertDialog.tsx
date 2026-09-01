import { AlertTriangle, CheckCircle2, Info, XCircle, type LucideIcon } from 'lucide-react';
import { cx } from '@/utils/cx';
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

const variantIcon: Record<Variant, LucideIcon> = {
  success: CheckCircle2,
  warning: AlertTriangle,
  destructive: XCircle,
  info: Info,
  confirmation: AlertTriangle,
};

const variantIconClass: Record<Variant, string> = {
  success: 'alert-icon--success',
  warning: 'alert-icon--warning',
  destructive: 'alert-icon--destructive',
  info: 'alert-icon--info',
  confirmation: 'alert-icon--warning',
};

interface AlertDialogProps {
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
  const Icon = variantIcon[variant];
  const isDestructive = variant === 'destructive';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size={size}>
        <DialogHeader>
          <div className="alert-header">
            <div className={cx('alert-icon', variantIconClass[variant])}>
              <Icon aria-hidden="true" />
            </div>
            <div className="alert-header__body">
              <DialogTitle>{title}</DialogTitle>
              {description && <DialogDescription>{description}</DialogDescription>}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            variant={isDestructive ? 'destructive' : 'primary'}
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
