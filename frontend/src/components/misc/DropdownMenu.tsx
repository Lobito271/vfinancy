import * as React from 'react';
import { Menu } from '@base-ui/react/menu';
import { cx } from '@/utils/cx';

export const DropdownMenu = Menu.Root;

export function DropdownMenuTrigger({
  asChild,
  children,
  ...props
}: Omit<React.ComponentPropsWithoutRef<typeof Menu.Trigger>, 'className'> & {
  className?: string;
  asChild?: boolean;
}) {
  return (
    <Menu.Trigger
      {...props}
      render={asChild && React.isValidElement(children) ? children : undefined}
    >
      {children}
    </Menu.Trigger>
  );
}

export const DropdownMenuContent = React.forwardRef<
  React.ComponentRef<typeof Menu.Popup>,
  Omit<React.ComponentPropsWithoutRef<typeof Menu.Popup>, 'className'> & {
    className?: string;
    align?: React.ComponentPropsWithoutRef<typeof Menu.Positioner>['align'];
    sideOffset?: number;
  }
>(({ className, align, sideOffset = 6, ...props }, ref) => (
  <Menu.Portal>
    <Menu.Positioner align={align} sideOffset={sideOffset} className="menu-positioner">
      <Menu.Popup ref={ref} className={cx('menu-content', className)} {...props} />
    </Menu.Positioner>
  </Menu.Portal>
));
DropdownMenuContent.displayName = 'DropdownMenuContent';

export const DropdownMenuItem = React.forwardRef<
  React.ComponentRef<typeof Menu.Item>,
  Omit<React.ComponentPropsWithoutRef<typeof Menu.Item>, 'className'> & {
    className?: string;
    inset?: boolean;
    danger?: boolean;
    onSelect?: () => void;
  }
>(({ className, inset, danger, onSelect, ...props }, ref) => (
  <Menu.Item
    ref={ref}
    onClick={onSelect}
    className={cx('menu-item', inset && 'menu-item--inset', danger && 'menu-item--danger', className)}
    {...props}
  />
));
DropdownMenuItem.displayName = 'DropdownMenuItem';
