import { cx } from '@/utils/cx';

export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cx('skeleton', className)}
      aria-hidden="true"
      {...props}
    />
  );
}
