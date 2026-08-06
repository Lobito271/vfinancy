import * as React from 'react';
import { cn } from '@/lib/utils';

export function PageContainer({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('mx-auto w-full max-w-7xl space-y-6 p-6', className)} {...props} />;
}

export function PageHeader({
  title,
  subtitle,
  actions,
  className,
}: {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between', className)}>
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {subtitle && <p className="text-sm text-muted-foreground">{subtitle}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}

export function Section({
  title,
  description,
  actions,
  className,
  children,
}: {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section className={cn('space-y-3', className)}>
      {(title || actions) && (
        <div className="flex items-end justify-between gap-3">
          <div className="space-y-0.5">
            {title && <h2 className="text-base font-semibold">{title}</h2>}
            {description && <p className="text-sm text-muted-foreground">{description}</p>}
          </div>
          {actions}
        </div>
      )}
      {children}
    </section>
  );
}

export function Stack({
  direction = 'col',
  gap = 4,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { direction?: 'row' | 'col'; gap?: 1 | 2 | 3 | 4 | 6 | 8 }) {
  return (
    <div
      className={cn(
        'flex',
        direction === 'row' ? 'flex-row' : 'flex-col',
        `gap-${gap}`,
        className,
      )}
      {...props}
    />
  );
}

export function Grid({
  cols = 3,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { cols?: 1 | 2 | 3 | 4 | 5 | 6 }) {
  return (
    <div
      className={cn(
        'grid gap-4',
        cols === 1 && 'grid-cols-1',
        cols === 2 && 'grid-cols-1 sm:grid-cols-2',
        cols === 3 && 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
        cols === 4 && 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
        cols === 5 && 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-5',
        cols === 6 && 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6',
        className,
      )}
      {...props}
    />
  );
}
