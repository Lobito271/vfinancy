import * as React from 'react';
import { cx } from '@/utils/cx';
import { formatDate } from '@/utils/format';

interface DateInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  invalid?: boolean;
  placeholder?: string;
}

// Forces the visible date format to DD/MM/YYYY regardless of browser
// locale: the formatted text is rendered below the native date input,
// which stays invisible but still opens the native picker on click.
export const DateInput = React.forwardRef<HTMLInputElement, DateInputProps>(
  ({ className, invalid, value, placeholder, ...props }, ref) => (
    <span className={cx('date-input', invalid && 'date-input--invalid', className)}>
      <span className="date-input__display" aria-hidden="true">
        {value ? formatDate(value as string) : placeholder ?? ''}
      </span>
      <input
        type="date"
        ref={ref}
        value={value}
        aria-invalid={invalid || undefined}
        className="date-input__native"
        {...props}
      />
    </span>
  ),
);
DateInput.displayName = 'DateInput';