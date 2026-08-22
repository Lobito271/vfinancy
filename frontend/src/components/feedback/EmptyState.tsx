import { Inbox, type LucideIcon } from 'lucide-react';
import { cx } from '@/utils/cx';
import { Button } from '@/components/button';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
  className?: string;
}

export function EmptyState({
  icon: Icon = Inbox,
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div className={cx('empty-state', className)}>
      <div className="empty-state__icon">
        <Icon aria-hidden="true" />
      </div>
      <div className="empty-state__body">
        <h3 className="empty-state__title">{title}</h3>
        {description && (
          <p className="empty-state__description">{description}</p>
        )}
      </div>
      {action && <Button onClick={action.onClick}>{action.label}</Button>}
    </div>
  );
}
