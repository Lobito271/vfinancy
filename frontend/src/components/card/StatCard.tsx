import { ArrowDownRight, ArrowUpRight, Minus, type LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';

interface StatCardProps {
  label: string;
  value: string;
  icon?: LucideIcon;
  change?: number;
  changeLabel?: string;
  className?: string;
}

export function StatCard({ label, value, icon: Icon, change, changeLabel, className }: StatCardProps) {
  const isUp = change !== undefined && change > 0;
  const isDown = change !== undefined && change < 0;
  const Trend = isUp ? ArrowUpRight : isDown ? ArrowDownRight : Minus;

  return (
    <div className={cn('rounded-lg border bg-card p-6 shadow-sm', className)}>
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{label}</p>
        {Icon && <Icon className="h-4 w-4 text-muted-foreground" aria-hidden="true" />}
      </div>
      <p className="mt-2 text-3xl font-semibold tabular-nums tracking-tight">{value}</p>
      {change !== undefined && (
        <div className="mt-2 flex items-center gap-1 text-xs">
          <span
            className={cn(
              'inline-flex items-center gap-0.5 rounded-full px-1.5 py-0.5 font-medium',
              isUp && 'bg-success/10 text-success',
              isDown && 'bg-destructive/10 text-destructive',
              !isUp && !isDown && 'bg-muted text-muted-foreground',
            )}
          >
            <Trend className="h-3 w-3" />
            {Math.abs(change).toFixed(1)}%
          </span>
          {changeLabel && <span className="text-muted-foreground">{changeLabel}</span>}
        </div>
      )}
    </div>
  );
}
