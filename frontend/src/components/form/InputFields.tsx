import { useState } from 'react';
import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Eye, EyeOff } from 'lucide-react';
import { Field } from './Field';
import { Input } from '@/components/input';
import { TextField } from './BasicFields';

interface PasswordFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  autoComplete?: string;
}

export function PasswordField<T extends FieldValues>({ name, label, description, required, className, autoComplete = 'current-password' }: PasswordFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const [show, setShow] = useState(false);
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <div className="input-affix input-affix--suffix">
        <Input
          id={String(name)}
          type={show ? 'text' : 'password'}
          autoComplete={autoComplete}
          invalid={!!error}
          {...register(name)}
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
    </Field>
  );
}

interface EmailFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function EmailField<T extends FieldValues>(props: EmailFieldProps<T>) {
  return <TextField {...props} type="email" />;
}

void TextField;
