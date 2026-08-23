import * as React from 'react';
import { cx } from '@/utils/cx';

export type BadgeVariant =
  | 'primary'
  | 'secondary'
  | 'outline'
  | 'success'
  | 'warning'
  | 'destructive'
  | 'info'
  | 'muted';

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

export const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant = 'primary', ...props }, ref) => (
    <span ref={ref} className={cx('badge', `badge--${variant}`, className)} {...props} />
  ),
);
Badge.displayName = 'Badge';
