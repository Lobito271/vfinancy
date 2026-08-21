import * as React from 'react';
import { Input } from '@/components/input';

export interface MoneyInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'value' | 'onChange'> {
  value: number;
  onValueChange: (n: number) => void;
  currency?: string;
}

const symbols: Record<string, string> = { PEN: 'S/', USD: '$' };

export const MoneyInput = React.forwardRef<HTMLInputElement, MoneyInputProps>(
  ({ value, onValueChange, currency = 'PEN', ...props }, ref) => {
    const [raw, setRaw] = React.useState<string>(() => formatRaw(value));

    React.useEffect(() => {
      setRaw(formatRaw(value));
    }, [value]);

    return (
      <div className="input-affix input-affix--prefix">
        <span className="input-affix__prefix" aria-hidden="true">
          {symbols[currency] ?? currency}
        </span>
        <Input
          ref={ref}
          type="text"
          inputMode="decimal"
          value={raw}
          onChange={(e) => {
            const next = e.target.value.replace(/[^0-9.,-]/g, '').replace(',', '.');
            setRaw(next);
            const n = Number(next);
            if (!Number.isNaN(n)) onValueChange(n);
          }}
          {...props}
        />
      </div>
    );
  },
);
MoneyInput.displayName = 'MoneyInput';

function formatRaw(n: number): string {
  if (!n) return '';
  return String(n);
}
