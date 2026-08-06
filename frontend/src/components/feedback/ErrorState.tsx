import { AlertCircle, type LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';
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
    <div
      className={cn(
        'flex flex-col items-center justify-center gap-3 rounded-lg border border-destructive/30 bg-destructive/5 py-16 text-center',
        className,
      )}
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
        <Icon className="h-6 w-6 text-destructive" aria-hidden="true" />
      </div>
      <div className="space-y-1">
        <h3 className="text-sm font-semibold">{title}</h3>
        <p className="mx-auto max-w-sm text-sm text-muted-foreground">{description}</p>
      </div>
      {onRetry && <Button onClick={onRetry}>Reintentar</Button>}
    </div>
  );
}
