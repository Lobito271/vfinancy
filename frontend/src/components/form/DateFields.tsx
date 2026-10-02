import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field, fieldDescribedBy } from './Field';
import { DateInput } from '@/components/input';
import { formatDate } from '@/utils/format';

interface DateFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  min?: string;
  max?: string;
  showFormatted?: boolean;
}

export function DateField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
  min,
  max,
  showFormatted = false,
}: DateFieldProps<T>) {
  const { register, formState, watch, setValue } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const value = watch(name) as string | undefined;
  const id = String(name);
  const hint = description ?? (showFormatted && value ? formatDate(value) : undefined);
  return (
    <Field label={label} required={required} description={hint} error={error} className={className} htmlFor={id}>
      <DateInput
        id={id}
        invalid={!!error}
        aria-required={required || undefined}
        aria-describedby={fieldDescribedBy(id, hint, error)}
        min={min}
        max={max}
        value={value ?? ''}
        placeholder={!required && !value ? 'DD/MM/AAAA' : undefined}
        onClear={
          !required
            ? () => (setValue as (n: FieldPath<T>, v: string) => void)(name, '')
            : undefined
        }
        {...register(name)}
      />
    </Field>
  );
}
