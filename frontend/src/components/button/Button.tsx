import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { Loader2 } from 'lucide-react';
import { cx } from '@/utils/cx';

type ButtonVariant =
  | 'primary'
  | 'secondary'
  | 'outline'
  | 'ghost'
  | 'link'
  | 'destructive'
  | 'success';

type ButtonSize = 'sm' | 'md' | 'lg' | 'icon' | 'icon-sm';

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  asChild?: boolean;
  loading?: boolean;
}

const sizeClass: Record<ButtonSize, string> = {
  sm: 'btn--sm',
  md: 'btn--md',
  lg: 'btn--lg',
  icon: 'btn--icon',
  'icon-sm': 'btn--icon-sm',
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', asChild = false, loading = false, children, disabled, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button';
    return (
      <Comp
        ref={ref}
        className={cx('btn', `btn--${variant}`, sizeClass[size], className)}
        disabled={disabled || loading}
        {...props}
      >
        {loading && <Loader2 aria-hidden="true" />}
        {children}
      </Comp>
    );
  },
);
Button.displayName = 'Button';
