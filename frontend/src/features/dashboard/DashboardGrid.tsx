import type { ReactNode } from 'react';
import { cn } from '@/utils/cn';

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

const sizeToCol: Record<WidgetSize, string> = {
  sm: 'col-span-12 sm:col-span-6 lg:col-span-3',
  md: 'col-span-12 sm:col-span-6 lg:col-span-4',
  lg: 'col-span-12 lg:col-span-6',
  xl: 'col-span-12 lg:col-span-8',
  full: 'col-span-12',
};

export function DashboardGrid({ items, className }: DashboardGridProps) {
  return (
    <div className={cn('grid grid-cols-12 gap-4', className)}>
      {items.map((item) => (
        <div key={item.id} className={cn(sizeToCol[item.size ?? 'md'])}>
          {item.content}
        </div>
      ))}
    </div>
  );
}
