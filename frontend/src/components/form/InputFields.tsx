import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field, fieldDescribedBy } from './Field';
import { PasswordInput } from '@/components/input';
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
  const id = String(name);
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={id}>
      <PasswordInput
        id={id}
        autoComplete={autoComplete}
        invalid={!!error}
        aria-required={required || undefined}
        aria-describedby={fieldDescribedBy(id, description, error)}
        {...register(name)}
      />
    </Field>
  );
}

interface EmailFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  autoComplete?: string;
}

export function EmailField<T extends FieldValues>(props: EmailFieldProps<T>) {
  return <TextField {...props} type="email" />;
}

void TextField;
