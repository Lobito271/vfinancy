import * as React from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import { cx } from '@/utils/cx';

export const Dialog = DialogPrimitive.Root;

const DialogPortal = DialogPrimitive.Portal;

export const DialogOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cx('dialog-overlay', className)}
    {...props}
  />
));
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName;

type Size = 'sm' | 'md' | 'lg' | 'xl';

const sizeClass: Record<Size, string> = {
  sm: 'dialog-content--sm',
  md: 'dialog-content--md',
  lg: 'dialog-content--lg',
  xl: 'dialog-content--xl',
};

export interface DialogContentProps
  extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> {
  size?: Size;
}

export const DialogContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  DialogContentProps
>(({ className, children, size = 'md', ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay>
      <DialogPrimitive.Content
        ref={ref}
        className={cx('dialog-content', sizeClass[size], className)}
        {...props}
      >
        {children}
        <DialogPrimitive.Close className="dialog-close" aria-label="Cerrar">
          <X />
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogOverlay>
  </DialogPortal>
));
DialogContent.displayName = DialogPrimitive.Content.displayName;

export const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cx('dialog-header', className)} {...props} />
);
DialogHeader.displayName = 'DialogHeader';

export const DialogFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cx('dialog-footer', className)} {...props} />
);
DialogFooter.displayName = 'DialogFooter';

export const DialogTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    ref={ref}
    className={cx('dialog-title', className)}
    {...props}
  />
));
DialogTitle.displayName = DialogPrimitive.Title.displayName;

export const DialogDescription = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    ref={ref}
    className={cx('dialog-description', className)}
    {...props}
  />
));
DialogDescription.displayName = DialogPrimitive.Description.displayName;
