import type { ReactNode } from 'react';
import { cx } from '@/utils/cx';

interface FieldProps {
  label?: string;
  required?: boolean;
  description?: string;
  error?: string;
  children: ReactNode;
  className?: string;
  htmlFor?: string;
}

export function fieldDescribedBy(id: string, description?: string, error?: string) {
  if (error) return `${id}-error`;
  return description ? `${id}-hint` : undefined;
}

export function Field({ label, required, description, error, children, className, htmlFor }: FieldProps) {
  return (
    <div className={cx('field', className)}>
      {label && (
        <label htmlFor={htmlFor} className="label">
          {label}
          {required && (
            <span className="required-mark" aria-hidden="true">
              *
            </span>
          )}
        </label>
      )}
      {children}
      {description && !error && (
        <p id={htmlFor ? `${htmlFor}-hint` : undefined} className="field-hint">{description}</p>
      )}
      {error && (
        <p id={htmlFor ? `${htmlFor}-error` : undefined} role="alert" className="field-error">
          {error}
        </p>
      )}
    </div>
  );
}
