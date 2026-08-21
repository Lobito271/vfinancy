import * as React from 'react';
import * as SwitchPrimitive from '@radix-ui/react-switch';
import { cx } from '@/utils/cx';

export const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitive.Root
    ref={ref}
    className={cx('switch', className)}
    {...props}
  >
    <SwitchPrimitive.Thumb className="switch__thumb" />
  </SwitchPrimitive.Root>
));
Switch.displayName = SwitchPrimitive.Root.displayName;
