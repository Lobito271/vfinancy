import type { ReactNode } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/card';
import { Spinner } from '@/components/feedback';
import { cn } from '@/utils/cn';

export interface WidgetShellProps {
  title: string;
  description?: string;
  loading?: boolean;
  error?: Error | null;
  actions?: ReactNode;
  className?: string;
  children: ReactNode;
}

export function WidgetShell({ title, description, loading, error, actions, className, children }: WidgetShellProps) {
  return (
    <Card className={cn('h-full', className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0">
        <div className="space-y-0.5">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </div>
        {actions}
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex h-24 items-center justify-center">
            <Spinner />
          </div>
        ) : error ? (
          <p className="text-sm text-destructive">{error.message}</p>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}
