import { cn } from '@/lib/utils';
import { formatCurrency } from '@/lib/utils';

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
      className={cn(
        'tabular-nums',
        signed && isNeg && 'text-destructive',
        signed && isPos && 'text-success',
        className,
      )}
    >
      {signed && isNeg ? `-${formatted}` : signed && isPos ? `+${formatted}` : formatted}
    </span>
  );
}
