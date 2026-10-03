import * as React from 'react';
import { CalendarDays, X } from 'lucide-react';
import { cx } from '@/utils/cx';
import { formatDate } from '@/utils/format';

interface DateInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  invalid?: boolean;
  placeholder?: string;
  onClear?: () => void;
}

// Forces the visible date format to DD/MM/YYYY regardless of browser
// locale: the formatted text is rendered below the native date input,
// which stays invisible but still opens the native picker on click.
export const DateInput = React.forwardRef<HTMLInputElement, DateInputProps>(
  ({ className, invalid, value, placeholder, onClear, ...props }, ref) => (
    <span className={cx('date-input', invalid && 'date-input--invalid', Boolean(onClear && value) && 'date-input--clearable', className)}>
      <input
        type="date"
        ref={ref}
        value={value}
        aria-invalid={invalid || undefined}
        className="date-input__native"
        {...props}
      />
      <span className={cx('date-input__display', !value && placeholder && 'date-input__display--placeholder')} aria-hidden="true">
        {value ? formatDate(value as string) : placeholder}
      </span>
      <CalendarDays className="date-input__icon" aria-hidden="true" />
      {onClear && value && (
        <button
          type="button"
          className="date-input__clear"
          aria-label="Limpiar fecha"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            onClear();
          }}
        >
          <X aria-hidden="true" />
        </button>
      )}
    </span>
  ),
);
DateInput.displayName = 'DateInput';
