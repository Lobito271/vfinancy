import type { ReactNode } from 'react';
import { cx } from '@/utils/cx';

export type WidgetSize = 'sm' | 'md' | 'lg' | 'xl' | 'full';

export interface DashboardGridItem {
  id: string;
  size?: WidgetSize;
  content: ReactNode;
}

interface DashboardGridProps {
  items: DashboardGridItem[];
  className?: string;
}

const sizeClass: Record<WidgetSize, string> = {
  sm: 'widget--sm',
  md: 'widget--md',
  lg: 'widget--lg',
  xl: 'widget--xl',
  full: '',
};

export function DashboardGrid({ items, className }: DashboardGridProps) {
  return (
    <div className={cx('widget-grid', className)}>
      {items.map((item) => (
        <div key={item.id} className={cx(sizeClass[item.size ?? 'md'])}>
          {item.content}
        </div>
      ))}
    </div>
  );
}
