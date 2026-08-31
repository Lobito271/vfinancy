import { useFormContext, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { NumberField as NumberFieldPrimitive } from '@base-ui/react/number-field';
import type { ReactNode } from 'react';
import { Field } from './Field';
import { Input, Textarea } from '@/components/input';

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

interface TextFieldProps<T extends FieldValues> extends BaseFieldProps<T>, Omit<React.InputHTMLAttributes<HTMLInputElement>, 'name' | 'value' | 'onChange' | 'onBlur'> {
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

interface NumberFieldProps<T extends FieldValues> extends BaseFieldProps<T> {
  min?: number;
  max?: number;
  step?: number | 'any';
  placeholder?: string;
  readOnly?: boolean;
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
  disabled,
  readOnly,
  placeholder,
}: NumberFieldProps<T>) {
  const { control, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <NumberFieldPrimitive.Root
            id={String(name)}
            className="number-field"
            locale="es-PE"
            value={typeof field.value === 'number' ? field.value : null}
            min={min}
            max={max}
            step={step}
            disabled={disabled}
            readOnly={readOnly}
            onValueChange={(value) => field.onChange((value ?? 0) as never)}
          >
            <NumberFieldPrimitive.Input
              className="input tabular"
              aria-invalid={!!error || undefined}
              placeholder={placeholder}
            />
          </NumberFieldPrimitive.Root>
        )}
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
        {...register(name)}
        {...rest}
      />
    </Field>
  );
}
