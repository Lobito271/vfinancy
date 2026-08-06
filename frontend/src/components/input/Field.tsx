import * as React from 'react';
import { cn } from '@/lib/utils';

export interface FieldProps extends React.HTMLAttributes<HTMLDivElement> {
  error?: string;
  description?: string;
  required?: boolean;
}

export const Field = React.forwardRef<HTMLDivElement, FieldProps>(
  ({ className, error, description, children, ...props }, ref) => (
    <div ref={ref} className={cn('flex flex-col gap-1.5', className)} {...props}>
      {children}
      {description && !error && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
      {error && (
        <p role="alert" className="text-xs text-destructive">
          {error}
        </p>
      )}
    </div>
  ),
);
Field.displayName = 'Field';
