import * as React from 'react';
import * as LabelPrimitive from '@radix-ui/react-label';
import { cx } from '@/utils/cx';

export const Label = React.forwardRef<
  React.ElementRef<typeof LabelPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root> & { required?: boolean }
>(({ className, required, children, ...props }, ref) => (
  <LabelPrimitive.Root
    ref={ref}
    className={cx('label', className)}
    {...props}
  >
    {children}
    {required && (
      <span className="required-mark" aria-hidden="true">
        *
      </span>
    )}
  </LabelPrimitive.Root>
));
Label.displayName = LabelPrimitive.Root.displayName;
