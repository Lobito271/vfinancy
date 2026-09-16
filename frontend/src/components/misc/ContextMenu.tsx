import * as React from 'react';
import { ContextMenu as ContextMenuPrimitive } from '@base-ui/react/context-menu';
import { cx } from '@/utils/cx';

export const ContextMenu = ContextMenuPrimitive.Root;

export const ContextMenuTrigger = React.forwardRef<
  React.ComponentRef<typeof ContextMenuPrimitive.Trigger>,
  Omit<React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Trigger>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <ContextMenuPrimitive.Trigger ref={ref} className={className} {...props} />
));
ContextMenuTrigger.displayName = 'ContextMenuTrigger';

export const ContextMenuContent = React.forwardRef<
  React.ComponentRef<typeof ContextMenuPrimitive.Popup>,
  Omit<React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Popup>, 'className'> & { className?: string }
>(({ className, children, ...props }, ref) => (
  <ContextMenuPrimitive.Portal>
    <ContextMenuPrimitive.Positioner className="menu-positioner">
      <ContextMenuPrimitive.Popup ref={ref} className={cx('menu-content', className)} {...props}>
        {children}
      </ContextMenuPrimitive.Popup>
    </ContextMenuPrimitive.Positioner>
  </ContextMenuPrimitive.Portal>
));
ContextMenuContent.displayName = 'ContextMenuContent';

export const ContextMenuItem = React.forwardRef<
  React.ComponentRef<typeof ContextMenuPrimitive.Item>,
  Omit<React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Item>, 'className'> & {
    className?: string;
    danger?: boolean;
    onSelect?: () => void;
  }
>(({ className, danger, onSelect, ...props }, ref) => (
  <ContextMenuPrimitive.Item
    ref={ref}
    onClick={onSelect}
    className={cx('menu-item', danger && 'menu-item--danger', className)}
    {...props}
  />
));
ContextMenuItem.displayName = 'ContextMenuItem';