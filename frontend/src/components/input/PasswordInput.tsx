import * as React from 'react';
import { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { Input } from './Input';

interface PasswordInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  invalid?: boolean;
}

export const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ invalid, ...props }, ref) => {
    const [show, setShow] = useState(false);
    return (
      <div className="input-affix input-affix--suffix">
        <Input
          ref={ref}
          type={show ? 'text' : 'password'}
          invalid={invalid}
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
