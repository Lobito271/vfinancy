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
      {showFormatted && value && <p className="field-hint">{formatDate(value)}</p>}
    </Field>
  );
}
