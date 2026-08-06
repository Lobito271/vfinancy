import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field } from './Field';
import { Input } from '@/components/input';
import { formatDate } from '@/utils/format';

interface DateFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  min?: string;
  max?: string;
}

export function DateField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
  min,
  max,
}: DateFieldProps<T>) {
  const { register, formState, watch } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const value = watch(name) as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Input
        id={String(name)}
        type="date"
        invalid={!!error}
        min={min}
        max={max}
        {...register(name)}
      />
      {value && <p className="text-xs text-muted-foreground">{formatDate(value)}</p>}
    </Field>
  );
}

interface DateRangeFieldProps<T extends FieldValues> {
  fromName: FieldPath<T>;
  toName: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function DateRangeField<T extends FieldValues>({
  fromName,
  toName,
  label,
  description,
  required,
  className,
}: DateRangeFieldProps<T>) {
  return (
    <Field label={label} required={required} description={description} className={className}>
      <div className="grid grid-cols-2 gap-2">
        <DateField name={fromName} label="Desde" />
        <DateField name={toName} label="Hasta" />
      </div>
    </Field>
  );
}

interface DateTimeFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function DateTimeField<T extends FieldValues>({ name, label, description, required, className }: DateTimeFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Input id={String(name)} type="datetime-local" invalid={!!error} {...register(name)} />
    </Field>
  );
}
