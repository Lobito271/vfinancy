import * as React from 'react';
import { Drawer as DrawerPrimitive } from '@base-ui/react/drawer';
import { cx } from '@/utils/cx';

interface DrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: React.ReactNode;
  description?: React.ReactNode;
  footer?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}

export function Drawer({ open, onOpenChange, title, description, footer, children, className }: DrawerProps) {
  return (
    <DrawerPrimitive.Root open={open} onOpenChange={onOpenChange} swipeDirection="right">
      <DrawerPrimitive.Portal>
        <DrawerPrimitive.Backdrop className="drawer-backdrop" />
        <DrawerPrimitive.Viewport className="drawer-viewport">
          <DrawerPrimitive.Popup className={cx('drawer-popup', className)}>
            <DrawerPrimitive.Content className="drawer-content">
              {(title || description) && (
                <div className="drawer-header">
                  {title && <DrawerPrimitive.Title className="drawer-title">{title}</DrawerPrimitive.Title>}
                  {description && (
                    <DrawerPrimitive.Description className="drawer-description">
                      {description}
                    </DrawerPrimitive.Description>
                  )}
                </div>
              )}
              <div className="drawer-body">{children}</div>
              {footer && <div className="drawer-footer">{footer}</div>}
            </DrawerPrimitive.Content>
          </DrawerPrimitive.Popup>
        </DrawerPrimitive.Viewport>
      </DrawerPrimitive.Portal>
    </DrawerPrimitive.Root>
  );
}
