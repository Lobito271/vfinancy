import * as React from 'react';
import { Menu } from '@base-ui/react/menu';
import { Circle } from 'lucide-react';
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

export const DropdownMenuRadioGroup = Menu.RadioGroup;

export const DropdownMenuRadioItem = React.forwardRef<
  React.ComponentRef<typeof Menu.RadioItem>,
  Omit<React.ComponentPropsWithoutRef<typeof Menu.RadioItem>, 'className'> & { className?: string }
>(({ className, children, ...props }, ref) => (
  <Menu.RadioItem ref={ref} className={cx('menu-radio-item', className)} {...props}>
    <span className="menu-item-indicator">
      <Menu.RadioItemIndicator>
        <Circle />
      </Menu.RadioItemIndicator>
    </span>
    {children}
  </Menu.RadioItem>
));
DropdownMenuRadioItem.displayName = 'DropdownMenuRadioItem';

export const DropdownMenuLabel = React.forwardRef<
  React.ComponentRef<typeof Menu.GroupLabel>,
  Omit<React.ComponentPropsWithoutRef<typeof Menu.GroupLabel>, 'className'> & {
    className?: string;
    inset?: boolean;
  }
>(({ className, inset, ...props }, ref) => (
  <Menu.Group>
    <Menu.GroupLabel
      ref={ref}
      className={cx('menu-label', inset && 'menu-label--inset', className)}
      {...props}
    />
  </Menu.Group>
));
DropdownMenuLabel.displayName = 'DropdownMenuLabel';

export const DropdownMenuSeparator = React.forwardRef<
  React.ComponentRef<typeof Menu.Separator>,
  Omit<React.ComponentPropsWithoutRef<typeof Menu.Separator>, 'className'> & { className?: string }
>(({ className, ...props }, ref) => (
  <Menu.Separator ref={ref} className={cx('menu-separator', className)} {...props} />
));
DropdownMenuSeparator.displayName = 'DropdownMenuSeparator';
