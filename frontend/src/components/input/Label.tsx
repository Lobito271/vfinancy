import * as React from 'react';
import { cx } from '@/utils/cx';

export const Label = React.forwardRef<
  HTMLLabelElement,
  React.LabelHTMLAttributes<HTMLLabelElement> & { required?: boolean }
>(({ className, required, children, ...props }, ref) => (
  <label ref={ref} className={cx('label', className)} {...props}>
    {children}
    {required && (
      <span className="required-mark" aria-hidden="true">
        *
      </span>
    )}
  </label>
));
Label.displayName = 'Label';
