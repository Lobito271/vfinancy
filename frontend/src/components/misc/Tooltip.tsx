import * as React from 'react';
import { Tooltip as TooltipPrimitive } from '@base-ui/react/tooltip';
import { cx } from '@/utils/cx';

export function TooltipProvider({
  delayDuration,
  ...props
}: Omit<React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Provider>, 'className' | 'delay'> & {
  className?: string;
  delayDuration?: number;
}) {
  return <TooltipPrimitive.Provider delay={delayDuration} {...props} />;
}

export const Tooltip = TooltipPrimitive.Root;

export function TooltipTrigger({
  asChild,
  children,
  ...props
}: Omit<React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Trigger>, 'className'> & {
  className?: string;
  asChild?: boolean;
}) {
  return (
    <TooltipPrimitive.Trigger
      {...props}
      render={asChild && React.isValidElement(children) ? children : undefined}
    >
      {children}
    </TooltipPrimitive.Trigger>
  );
}

export const TooltipContent = React.forwardRef<
  React.ComponentRef<typeof TooltipPrimitive.Popup>,
  Omit<React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Popup>, 'className'> & {
    className?: string;
    side?: React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Positioner>['side'];
    align?: React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Positioner>['align'];
    sideOffset?: number;
  }
>(({ className, side, align, sideOffset = 6, ...props }, ref) => (
  <TooltipPrimitive.Portal>
    <TooltipPrimitive.Positioner side={side} align={align} sideOffset={sideOffset} className="tooltip-positioner">
      <TooltipPrimitive.Popup ref={ref} className={cx('tooltip-content', className)} {...props} />
    </TooltipPrimitive.Positioner>
  </TooltipPrimitive.Portal>
));
TooltipContent.displayName = 'TooltipContent';
