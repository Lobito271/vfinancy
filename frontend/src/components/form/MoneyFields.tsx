import { useFormContext, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { NumberField as NumberFieldPrimitive } from '@base-ui/react/number-field';
import { Field } from './Field';
import { Currencies, type CurrencyCode, DefaultCurrency } from '@/constants/currencies';
import { formatCurrency } from '@/utils/format';

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
  const { control, formState, watch } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const value = watch(name) as number | undefined;

  const cur = Currencies[currency] ?? Currencies[DefaultCurrency];

  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <div className={showSymbol ? 'input-affix input-affix--prefix' : undefined}>
            {showSymbol && (
              <span className="input-affix__prefix" aria-hidden="true">
                {cur.symbol}
              </span>
            )}
            <NumberFieldPrimitive.Root
              id={String(name)}
              className="number-field"
              locale="es-PE"
              value={typeof field.value === 'number' ? field.value : null}
              disabled={disabled}
              onValueChange={(v) => field.onChange((v ?? 0) as never)}
            >
              <NumberFieldPrimitive.Input
                className="input tabular"
                aria-invalid={!!error || undefined}
                style={{ textAlign: 'right' }}
              />
            </NumberFieldPrimitive.Root>
          </div>
        )}
      />
      {value != null && Number.isFinite(value) && (
        <p className="field-hint">≈ {formatCurrency(value, currency)}</p>
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
  const { control, formState } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  return (
    <Field label={label} required={required} description={description} error={error} className={className} htmlFor={String(name)}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <div className="input-affix input-affix--suffix">
            <NumberFieldPrimitive.Root
              id={String(name)}
              className="number-field"
              locale="es-PE"
              min={0}
              max={100}
              step={0.01}
              value={typeof field.value === 'number' ? field.value : null}
              onValueChange={(v) => field.onChange((v ?? 0) as never)}
            >
              <NumberFieldPrimitive.Input
                className="input tabular"
                aria-invalid={!!error || undefined}
                style={{ textAlign: 'right' }}
              />
            </NumberFieldPrimitive.Root>
            <span className="input-affix__suffix" aria-hidden="true">%</span>
          </div>
        )}
      />
    </Field>
  );
}
