import type { ReactNode } from 'react';
import { cn } from '@/utils/cn';

interface FieldProps {
  label?: string;
  required?: boolean;
  description?: string;
  error?: string;
  children: ReactNode;
  className?: string;
  htmlFor?: string;
}

export function Field({ label, required, description, error, children, className, htmlFor }: FieldProps) {
  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      {label && (
        <label htmlFor={htmlFor} className="text-sm font-medium text-foreground">
          {label}
          {required && (
            <span className="ml-0.5 text-destructive" aria-hidden="true">
              *
            </span>
          )}
        </label>
      )}
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
  );
}

export function FieldDescription({ children }: { children: ReactNode }) {
  return <p className="text-xs text-muted-foreground">{children}</p>;
}

export function FieldError({ children }: { children?: ReactNode }) {
  if (!children) return null;
  return (
    <p role="alert" className="text-xs text-destructive">
      {children}
    </p>
  );
}
