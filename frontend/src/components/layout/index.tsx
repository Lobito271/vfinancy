import * as React from 'react';
import { cx } from '@/utils/cx';

export function PageContainer({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cx('page-container', className)} {...props} />;
}

export function PageHeader({
  title,
  subtitle,
  actions,
  eyebrow,
  className,
}: {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  eyebrow?: string;
  className?: string;
}) {
  return (
    <div className={cx('page-header', className)}>
      <div className="page-header__titles">
        {eyebrow && <span className="page-header__eyebrow">{eyebrow}</span>}
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
  withTick,
}: {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
  withTick?: boolean;
}) {
  return (
    <section className={cx('section', className)}>
      {(title || actions) && (
        <div className="section__head">
          <div className="section__head-titles">
            {title && <h2 className={cx('section-title', withTick && 'section-title--with-tick')}>{title}</h2>}
            {description && <p className="section-description">{description}</p>}
          </div>
          {actions}
        </div>
      )}
      {children}
    </section>
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
