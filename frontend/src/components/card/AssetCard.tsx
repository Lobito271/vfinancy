import * as React from 'react';
import { cx } from '@/utils/cx';
import { Card, CardContent } from './Card';

interface AssetCardProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  title: React.ReactNode;
  meta?: React.ReactNode;
  value?: React.ReactNode;
  actions?: React.ReactNode;
}

export function AssetCard({ title, meta, value, actions, className, children, ...props }: AssetCardProps) {
  return (
    <Card className={cx('asset-card', className)} {...props}>
      <CardContent className="asset-card__body">
        <div className="asset-card__top">
          <strong className="asset-card__title">{title}</strong>
          {actions}
        </div>
        {meta && <p className="asset-card__meta">{meta}</p>}
        {value && <p className="asset-card__value">{value}</p>}
        {children}
      </CardContent>
    </Card>
  );
}