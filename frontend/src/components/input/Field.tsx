import * as React from 'react';
import { cx } from '@/utils/cx';

export interface FieldProps extends React.HTMLAttributes<HTMLDivElement> {
  error?: string;
  description?: string;
  required?: boolean;
}

export const Field = React.forwardRef<HTMLDivElement, FieldProps>(
  ({ className, error, description, children, ...props }, ref) => (
    <div ref={ref} className={cx('field', className)} {...props}>
      {children}
      {description && !error && (
        <p className="field-hint">{description}</p>
      )}
      {error && (
        <p role="alert" className="field-error">
          {error}
        </p>
      )}
    </div>
  ),
);
Field.displayName = 'Field';
