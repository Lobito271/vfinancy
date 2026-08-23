import { AlertCircle, type LucideIcon } from 'lucide-react';
import { cx } from '@/utils/cx';
import { Button } from '@/components/button';

interface ErrorStateProps {
  icon?: LucideIcon;
  title?: string;
  description?: string;
  onRetry?: () => void;
  className?: string;
}

export function ErrorState({
  icon: Icon = AlertCircle,
  title = 'Algo salió mal',
  description = 'No se pudo completar la operación. Reintente.',
  onRetry,
  className,
}: ErrorStateProps) {
  return (
    <div className={cx('error-state', className)}>
      <div className="error-state__icon">
        <Icon aria-hidden="true" />
      </div>
      <div className="error-state__body">
        <h3 className="error-state__title">{title}</h3>
        <p className="error-state__description">{description}</p>
      </div>
      {onRetry && <Button onClick={onRetry}>Reintentar</Button>}
    </div>
  );
}
