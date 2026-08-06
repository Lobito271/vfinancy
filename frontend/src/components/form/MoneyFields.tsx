import { useEffect, useState } from 'react';
import { useFormContext, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field } from './Field';
import { Input } from '@/components/input';
import { Currencies, type CurrencyCode, DefaultCurrency } from '@/constants/currencies';
import { formatCurrency } from '@/utils/format';
import { cn } from '@/utils/cn';

interface MoneyFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  currency?: CurrencyCode;
  showSymbol?: boolean;
  className?: string;
  disabled?: boolean;
}

export function MoneyField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  currency = DefaultCurrency,
  showSymbol = true,
  className,
  disabled,
}: MoneyFieldProps<T>) {
  const { register, formState, watch, setValue } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const value = watch(name) as number | undefined;
  const [raw, setRaw] = useState<string>(value != null ? String(value) : '');

  useEffect(() => {
    setRaw(value != null ? String(value) : '');
  }, [value]);

  const cur = Currencies[currency] ?? Currencies[DefaultCurrency];

  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <div className="relative">
        {showSymbol && (
          <span
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground"
            aria-hidden="true"
          >
            {cur.symbol}
          </span>
        )}
        <Input
          id={String(name)}
          type="text"
          inputMode="decimal"
          disabled={disabled}
          invalid={!!error}
          className={cn(showSymbol ? 'pl-10' : '', 'text-right tabular-nums')}
          value={raw}
          onChange={(e) => {
            const next = e.target.value.replace(/[^0-9.,-]/g, '').replace(',', '.');
            setRaw(next);
            const n = Number(next);
            if (!Number.isNaN(n)) {
              setValue(name, n as never, { shouldDirty: true, shouldValidate: false });
            }
          }}
          onBlur={() => {
            void register(name);
          }}
        />
      </div>
      {value != null && Number.isFinite(value) && (
        <p className="text-xs text-muted-foreground">≈ {formatCurrency(value, currency)}</p>
      )}
    </Field>
  );
}

interface PercentageFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function PercentageField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  className,
}: PercentageFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <div className="relative">
        <Input
          id={String(name)}
          type="number"
          inputMode="decimal"
          step="0.01"
          min={0}
          max={100}
          invalid={!!error}
          className="pr-8 text-right tabular-nums"
          {...register(name, { valueAsNumber: true })}
        />
        <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground" aria-hidden="true">%</span>
      </div>
    </Field>
  );
}

interface CurrencyFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
}

export function CurrencyField<T extends FieldValues>({ name, label, description, required, className }: CurrencyFieldProps<T>) {
  const { register, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <select
        id={String(name)}
        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        {...register(name)}
      >
        {Object.values(Currencies).map((c) => (
          <option key={c.code} value={c.code}>
            {c.code} — {c.name}
          </option>
        ))}
      </select>
    </Field>
  );
}
