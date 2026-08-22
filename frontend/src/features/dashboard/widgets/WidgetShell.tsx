import type { ReactNode } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/card';
import { Spinner } from '@/components/feedback';
import { cx } from '@/utils/cx';

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
    <Card className={cx('fill-parent', className)}>
      <CardHeader className="card-header--row">
        <div className="stack stack--xs">
          <CardTitle className="card-title--sm">{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </div>
        {actions}
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="widget-loading">
            <Spinner />
          </div>
        ) : error ? (
          <p className="error-text">{error.message}</p>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}
