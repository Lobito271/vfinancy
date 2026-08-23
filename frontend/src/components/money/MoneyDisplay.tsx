import { cx } from '@/utils/cx';
import { formatCurrency } from '@/utils/format';

export interface MoneyDisplayProps {
  value: number;
  currency?: string;
  className?: string;
  signed?: boolean;
}

export function MoneyDisplay({ value, currency = 'PEN', className, signed = false }: MoneyDisplayProps) {
  const isNeg = value < 0;
  const isPos = value > 0;
  const formatted = formatCurrency(Math.abs(value), currency);
  return (
    <span
      className={cx(
        'tabular',
        signed && isNeg && 'color-destructive',
        signed && isPos && 'color-success',
        className,
      )}
    >
      {signed && isNeg ? `-${formatted}` : signed && isPos ? `+${formatted}` : formatted}
    </span>
  );
}
