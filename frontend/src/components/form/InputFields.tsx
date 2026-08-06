import { useState } from 'react';
import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Eye, EyeOff, Search } from 'lucide-react';
import { Field } from './Field';
import { Input } from '@/components/input';
import { Button } from '@/components/button';
import { TextField } from './BasicFields';
import { validators } from '@/utils/validators';

interface SearchFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  placeholder?: string;
  className?: string;
  onSearch?: (value: string) => void;
}

export function SearchField<T extends FieldValues>({ name, label, description, placeholder, className, onSearch }: SearchFieldProps<T>) {
  const { register, formState, watch, setValue } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const value = (watch(name) as string) ?? '';
  return (
    <Field label={label} description={description} error={error} className={className} htmlFor={String(name)}>
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          id={String(name)}
          type="search"
          placeholder={placeholder}
          className="pl-9 pr-9"
          invalid={!!error}
          {...register(name)}
          onChange={(e) => {
            setValue(name, e.target.value as never, { shouldDirty: true });
            onSearch?.(e.target.value);
          }}
        />
        {value && (
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            onClick={() => setValue(name, '' as never)}
            className="absolute right-1 top-1/2 -translate-y-1/2"
            aria-label="Limpiar"
          >
            <EyeOff className="h-4 w-4" />
          </Button>
        )}
      </div>
    </Field>
  );
}

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
      <div className="relative">
        <Input
          id={String(name)}
          type={show ? 'text' : 'password'}
          autoComplete={autoComplete}
          invalid={!!error}
          className="pr-10"
          {...register(name)}
        />
        <button
          type="button"
          onClick={() => setShow((s) => !s)}
          className="absolute right-1 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground"
          aria-label={show ? 'Ocultar contraseña' : 'Mostrar contraseña'}
        >
          {show ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
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

interface PhoneFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function PhoneField<T extends FieldValues>({ name, label, description, required, className }: PhoneFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <div className="relative">
        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">+51</span>
        <Input
          id={String(name)}
          type="tel"
          inputMode="numeric"
          maxLength={9}
          placeholder="999 999 999"
          invalid={!!error}
          className="pl-12 tabular-nums"
          {...register(name, {
            validate: (v) => !v || validators.isPhonePE(String(v)) || 'Teléfono inválido (9 dígitos)',
          })}
        />
      </div>
    </Field>
  );
}

void TextField;
