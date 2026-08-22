import { ArrowDownRight, ArrowUpRight, Minus, type LucideIcon } from 'lucide-react';
import { cx } from '@/utils/cx';

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
    <div className={cx('stat-card', className)}>
      <div className="stat-card__top">
        <p className="stat-card__label">{label}</p>
        {Icon && <Icon aria-hidden="true" />}
      </div>
      <p className="stat-card__value">{value}</p>
      {change !== undefined && (
        <div className="stat-card__meta">
          <span
            className={cx(
              'trend',
              isUp && 'trend--up',
              isDown && 'trend--down',
              !isUp && !isDown && 'trend--flat',
            )}
          >
            <Trend />
            {Math.abs(change).toFixed(1)}%
          </span>
          {changeLabel && <span className="stat-card__change-label">{changeLabel}</span>}
        </div>
      )}
    </div>
  );
}
