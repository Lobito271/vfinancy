import { useEffect, useState } from 'react';
import { useFormContext, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field } from './Field';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { Spinner } from '@/components/feedback';
import { cn } from '@/utils/cn';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
  description?: string;
}

interface SelectFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
  options: SelectOption[];
  loading?: boolean;
  clearable?: boolean;
  onChange?: (value: string) => void;
}

export function SelectField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  placeholder = 'Seleccione…',
  className,
  disabled,
  options,
  loading,
  clearable = true,
  onChange,
}: SelectFieldProps<T>) {
  const { control, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <div className="relative">
            <Select
              value={field.value ?? ''}
              onValueChange={(v) => {
                field.onChange(v);
                onChange?.(v);
              }}
              disabled={disabled || loading}
            >
              <SelectTrigger invalid={!!error} className={cn(loading && 'opacity-60')}>
                <SelectValue placeholder={loading ? 'Cargando…' : placeholder} />
              </SelectTrigger>
              <SelectContent>
                {options.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value} disabled={opt.disabled}>
                    {opt.label}
                  </SelectItem>
                ))}
                {options.length === 0 && !loading && (
                  <div className="p-2 text-center text-sm text-muted-foreground">Sin opciones</div>
                )}
              </SelectContent>
            </Select>
            {loading && (
              <Spinner size="sm" className="absolute right-10 top-1/2 -translate-y-1/2" />
            )}
            {clearable && field.value && !disabled && (
              <button
                type="button"
                onClick={() => field.onChange('')}
                className="absolute right-9 top-1/2 -translate-y-1/2 rounded p-1 text-xs text-muted-foreground hover:bg-accent hover:text-foreground"
                aria-label="Limpiar selección"
              >
                ×
              </button>
            )}
          </div>
        )}
      />
    </Field>
  );
}

interface AsyncSelectFieldProps<T extends FieldValues> extends Omit<SelectFieldProps<T>, 'options' | 'loading'> {
  loadOptions: () => Promise<SelectOption[]>;
  dependsOn?: unknown;
}

export function AsyncSelectField<T extends FieldValues>({
  loadOptions,
  dependsOn,
  ...rest
}: AsyncSelectFieldProps<T>) {
  const [options, setOptions] = useState<SelectOption[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void loadOptions().then((opts) => {
      if (!cancelled) {
        setOptions(opts);
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [dependsOn, loadOptions]);
  return <SelectField {...rest} options={options} loading={loading} />;
}
