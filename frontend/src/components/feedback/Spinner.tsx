import * as React from 'react';
import { Loader2 } from 'lucide-react';
import { cx } from '@/utils/cx';

interface SpinnerProps extends React.HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg';
}

export function Spinner({ className, size = 'md', ...props }: SpinnerProps) {
  return (
    <div role="status" aria-live="polite" className={cx('spinner', `spinner--${size}`, className)} {...props}>
      <Loader2 />
      <span className="sr-only">Cargando</span>
    </div>
  );
}
