import type { LucideIcon } from 'lucide-react';
import { cx } from '@/utils/cx';

interface StatCardProps {
  label: string;
  value: string;
  icon?: LucideIcon;
  hint?: string;
  className?: string;
}

export function StatCard({ label, value, icon: Icon, hint, className }: StatCardProps) {
  return (
    <div className={cx('stat-card', className)}>
      <p className="stat-card__label">
        {Icon && <Icon aria-hidden="true" />}
        {label}
      </p>
      <p className="stat-card__value">{value}</p>
      {hint && <p className="stat-card__hint">{hint}</p>}
    </div>
  );
}
