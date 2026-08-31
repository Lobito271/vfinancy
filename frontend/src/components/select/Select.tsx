import * as React from 'react';
import { Select as SelectPrimitive } from '@base-ui/react/select';
import { Check, ChevronDown } from 'lucide-react';
import { cx } from '@/utils/cx';

export const Select = SelectPrimitive.Root;
export const SelectValue = SelectPrimitive.Value;

export const SelectTrigger = React.forwardRef<
  React.ComponentRef<typeof SelectPrimitive.Trigger>,
  Omit<React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>, 'className'> & { className?: string; invalid?: boolean }
>(({ className, children, invalid, ...props }, ref) => (
  <SelectPrimitive.Trigger
    ref={ref}
    className={cx('select-trigger', invalid && 'select-trigger--invalid', className)}
    {...props}
  >
    {children}
    <SelectPrimitive.Icon className="select-trigger__icon">
      <ChevronDown />
    </SelectPrimitive.Icon>
  </SelectPrimitive.Trigger>
));
SelectTrigger.displayName = 'SelectTrigger';

export const SelectContent = React.forwardRef<
  React.ComponentRef<typeof SelectPrimitive.Popup>,
  Omit<React.ComponentPropsWithoutRef<typeof SelectPrimitive.Popup>, 'className'> & {
    className?: string;
    align?: React.ComponentPropsWithoutRef<typeof SelectPrimitive.Positioner>['align'];
    sideOffset?: number;
  }
>(({ className, children, align = 'start', sideOffset = 6, ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Positioner align={align} sideOffset={sideOffset} className="select-positioner">
      <SelectPrimitive.Popup ref={ref} className={cx('select-popup', className)} {...props}>
        {children}
      </SelectPrimitive.Popup>
    </SelectPrimitive.Positioner>
  </SelectPrimitive.Portal>
));
SelectContent.displayName = 'SelectContent';

export const SelectItem = React.forwardRef<
  React.ComponentRef<typeof SelectPrimitive.Item>,
  Omit<React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>, 'className'> & { className?: string }
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item ref={ref} className={cx('select-item', className)} {...props}>
    <SelectPrimitive.ItemIndicator className="select-item__indicator">
      <Check />
    </SelectPrimitive.ItemIndicator>
    <SelectPrimitive.ItemText className="select-item__text">{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
));
SelectItem.displayName = 'SelectItem';
