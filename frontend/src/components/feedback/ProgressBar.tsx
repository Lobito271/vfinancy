import { cx } from '@/utils/cx';

interface ProgressBarProps extends React.HTMLAttributes<HTMLDivElement> {
  value: number;
  max?: number;
  variant?: 'default' | 'success' | 'warning' | 'destructive';
}

const variantClass = {
  default: '',
  success: 'progress__bar--success',
  warning: 'progress__bar--warning',
  destructive: 'progress__bar--destructive',
};

export function ProgressBar({ value, max = 100, variant = 'default', className, ...props }: ProgressBarProps) {
  const pct = Math.max(0, Math.min(100, (value / max) * 100));
  return (
    <div
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      className={cx('progress', className)}
      {...props}
    >
      <div
        className={cx('progress__bar', variantClass[variant])}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}
