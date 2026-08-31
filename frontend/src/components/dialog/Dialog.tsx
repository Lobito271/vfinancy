import * as React from 'react';
import { Dialog as DialogPrimitive } from '@base-ui/react/dialog';
import { X } from 'lucide-react';
import { cx } from '@/utils/cx';

export const Dialog = DialogPrimitive.Root;

type Size = 'sm' | 'md' | 'lg' | 'xl';

const sizeClass: Record<Size, string> = {
  sm: 'dialog-content--sm',
  md: 'dialog-content--md',
  lg: 'dialog-content--lg',
  xl: 'dialog-content--xl',
};

interface DialogContentProps
  extends Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Popup>, 'className'> {
  className?: string;
  size?: Size;
}

export const DialogContent = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Popup>,
  DialogContentProps
>(({ className, children, size = 'md', ...props }, ref) => (
  <DialogPrimitive.Portal>
    <DialogPrimitive.Backdrop className="dialog-backdrop" />
    <DialogPrimitive.Popup
      ref={ref}
      className={cx('dialog-content', sizeClass[size], className)}
      {...props}
    >
      {children}
      <DialogPrimitive.Close className="dialog-close" aria-label="Cerrar">
        <X />
      </DialogPrimitive.Close>
    </DialogPrimitive.Popup>
  </DialogPrimitive.Portal>
));
DialogContent.displayName = 'DialogContent';

export const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cx('dialog-header', className)} {...props} />
);
DialogHeader.displayName = 'DialogHeader';

export const DialogFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cx('dialog-footer', className)} {...props} />
);
DialogFooter.displayName = 'DialogFooter';

export const DialogTitle = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Title>,
  Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title ref={ref} className={cx('dialog-title', className)} {...props} />
));
DialogTitle.displayName = 'DialogTitle';

export const DialogDescription = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Description>,
  Omit<React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description ref={ref} className={cx('dialog-description', className)} {...props} />
));
DialogDescription.displayName = 'DialogDescription';
