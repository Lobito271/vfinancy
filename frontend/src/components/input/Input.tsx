import * as React from 'react';
import { cx } from '@/utils/cx';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type = 'text', invalid, ...props }, ref) => (
    <input
      type={type}
      ref={ref}
      aria-invalid={invalid || undefined}
      className={cx('input', invalid && 'input--invalid', className)}
      {...props}
    />
  ),
);
Input.displayName = 'Input';

export const Textarea = React.forwardRef<
  HTMLTextAreaElement,
  React.TextareaHTMLAttributes<HTMLTextAreaElement> & { invalid?: boolean }
>(({ className, invalid, ...props }, ref) => (
  <textarea
    ref={ref}
    aria-invalid={invalid || undefined}
    className={cx('textarea', invalid && 'input--invalid', className)}
    {...props}
  />
));
Textarea.displayName = 'Textarea';
