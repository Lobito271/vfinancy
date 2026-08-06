import * as React from 'react';
import { cn } from '@/lib/utils';
import { Input } from '@/components/input';

export interface MoneyInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'value' | 'onChange'> {
  value: number;
  onValueChange: (n: number) => void;
  currency?: string;
}

export const MoneyInput = React.forwardRef<HTMLInputElement, MoneyInputProps>(
  ({ className, value, onValueChange, currency = 'PEN', ...props }, ref) => {
    const [raw, setRaw] = React.useState<string>(() => formatRaw(value));

    React.useEffect(() => {
      setRaw(formatRaw(value));
    }, [value]);

    return (
      <div className={cn('relative', className)}>
        <span
          className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground"
          aria-hidden="true"
        >
          {currency === 'PEN' ? 'S/' : currency === 'USD' ? '$' : currency}
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
          className="pl-12 text-right tabular-nums"
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
