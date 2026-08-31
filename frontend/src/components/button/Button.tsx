import * as React from 'react';
import { Button as BaseButton } from '@base-ui/react/button';
import { Loader2 } from 'lucide-react';
import { cx } from '@/utils/cx';

type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
type ButtonSize = 'sm' | 'md' | 'lg' | 'icon' | 'icon-sm';

const sizeClass: Record<ButtonSize, string> = {
  sm: 'btn--sm',
  md: 'btn--md',
  lg: 'btn--lg',
  icon: 'btn--icon',
  'icon-sm': 'btn--icon-sm',
};

interface ButtonProps
  extends Omit<React.ComponentPropsWithoutRef<typeof BaseButton>, 'className'> {
  className?: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
}

export function Button({
  className,
  variant = 'primary',
  size = 'md',
  loading = false,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <BaseButton
      className={cx('btn', `btn--${variant}`, sizeClass[size], className)}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <Loader2 aria-hidden="true" className="animate-spin" />}
      {children}
    </BaseButton>
  );
}
