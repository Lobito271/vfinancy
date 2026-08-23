import * as React from 'react';
import { cx } from '@/utils/cx';

export function PageContainer({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('page-container', className)} {...props} />;
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
    <div className={cx('page-header', className)}>
      <div className="page-header__titles">
        <h1 className="page-title">{title}</h1>
        {subtitle && <p className="page-subtitle">{subtitle}</p>}
      </div>
      {actions && <div className="page-header__actions">{actions}</div>}
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
    <section className={cx('section', className)}>
      {(title || actions) && (
        <div className="section__head">
          <div className="section__head-titles">
            {title && <h2 className="section-title">{title}</h2>}
            {description && <p className="section-description">{description}</p>}
          </div>
          {actions}
        </div>
      )}
      {children}
    </section>
  );
}

const gapRem: Record<number, string> = {
  1: '0.25rem',
  2: '0.5rem',
  3: '0.75rem',
  4: '1rem',
  6: '1.5rem',
  8: '2rem',
};

export function Stack({
  direction = 'col',
  gap = 4,
  className,
  style,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { direction?: 'row' | 'col'; gap?: 1 | 2 | 3 | 4 | 6 | 8 }) {
  return (
    <div
      className={cx('stack', className)}
      style={{
        flexDirection: direction === 'row' ? 'row' : 'column',
        gap: gapRem[gap],
        ...style,
      }}
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
      className={cx(`grid-${cols}`, className)}
      {...props}
    />
  );
}
