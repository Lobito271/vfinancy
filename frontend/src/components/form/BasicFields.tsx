import { useFormContext, Controller, type FieldPath, type FieldValues, type Control } from 'react-hook-form';
import type { ReactNode, ReactElement } from 'react';
import { Field } from './Field';
import { Input, Textarea, type InputProps } from '@/components/input';
import { cn } from '@/utils/cn';

interface BaseFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  disabled?: boolean;
  loading?: boolean;
  className?: string;
  children?: ReactNode;
}

interface TextFieldProps<T extends FieldValues> extends BaseFieldProps<T>, Omit<InputProps, 'name' | 'value' | 'onChange' | 'onBlur'> {
  type?: 'text' | 'email' | 'tel' | 'url' | 'password';
}

export function TextField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
  type = 'text',
  ...inputProps
}: TextFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Input
        id={String(name)}
        type={type}
        invalid={!!error}
        {...register(name, { valueAsNumber: false })}
        {...inputProps}
      />
    </Field>
  );
}

export function NumberField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
  min,
  max,
  step,
  ...inputProps
}: TextFieldProps<T> & { min?: number; max?: number; step?: number }) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Input
        id={String(name)}
        type="number"
        inputMode="decimal"
        invalid={!!error}
        min={min}
        max={max}
        step={step}
        className="tabular-nums"
        {...register(name, { valueAsNumber: true })}
        {...inputProps}
      />
    </Field>
  );
}

interface TextareaFieldProps<T extends FieldValues> extends BaseFieldProps<T> {
  placeholder?: string;
  rows?: number;
  maxLength?: number;
}

export function TextareaField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
  rows = 3,
  ...rest
}: TextareaFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Textarea
        id={String(name)}
        rows={rows}
        invalid={!!error}
        className={cn('min-h-[80px]', className)}
        {...register(name)}
        {...rest}
      />
    </Field>
  );
}

interface CheckboxFieldProps<T extends FieldValues> extends BaseFieldProps<T> {
  label: string;
  description?: string;
}

export function CheckboxField<T extends FieldValues>({ name, label, description, className }: CheckboxFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field error={error} className={className}>
      <label className="inline-flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          className="h-4 w-4 rounded-sm border border-input accent-primary"
          {...register(name)}
        />
        <span>{label}</span>
      </label>
      {description && <p className="ml-6 text-xs text-muted-foreground">{description}</p>}
    </Field>
  );
}

export function FormController<T extends FieldValues>(props: {
  name: FieldPath<T>;
  control?: Control<T>;
  render: (args: { value: unknown; onChange: (v: unknown) => void; onBlur: () => void; error?: string }) => ReactElement;
}) {
  const ctx = useFormContext<T>();
  return (
    <Controller
      control={props.control ?? ctx.control}
      name={props.name}
      render={({ field, fieldState }) =>
        props.render({
          value: field.value,
          onChange: field.onChange,
          onBlur: field.onBlur,
          error: fieldState.error?.message,
        })
      }
    />
  );
}
