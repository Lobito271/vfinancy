import * as React from 'react';
import { cx } from '@/utils/cx';

interface ListRowProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  title: React.ReactNode;
  meta?: React.ReactNode;
  trailing?: React.ReactNode;
}

export function ListRow({ title, meta, trailing, className, ...props }: ListRowProps) {
  return (
    <div className={cx('list-row', className)} {...props}>
      <span className="list-row__main">
        <strong className="list-row__title">{title}</strong>
        {meta && <small className="list-row__meta">{meta}</small>}
      </span>
      {trailing && <div className="list-row__trailing">{trailing}</div>}
    </div>
  );
}