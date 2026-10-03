import * as React from 'react';
import { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { cx } from '@/utils/cx';
import { Input } from './Input';

interface PasswordInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  invalid?: boolean;
}

export const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ invalid, disabled, ...props }, ref) => {
    const [show, setShow] = useState(false);
    return (
      <div className={cx('input-affix input-affix--suffix', invalid && 'input-affix--invalid', disabled && 'input-affix--disabled')}>
        <Input
          ref={ref}
          type={show ? 'text' : 'password'}
          invalid={invalid}
          disabled={disabled}
          {...props}
        />
        <button
          type="button"
          onClick={() => setShow((s) => !s)}
          className="input-affix__action"
          aria-label={show ? 'Ocultar contraseña' : 'Mostrar contraseña'}
        >
          {show ? <EyeOff /> : <Eye />}
        </button>
      </div>
    );
  },
);
PasswordInput.displayName = 'PasswordInput';
