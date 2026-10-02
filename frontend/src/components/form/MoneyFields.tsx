import { useFormContext, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { NumberField as NumberFieldPrimitive } from '@base-ui/react/number-field';
import { Field, fieldDescribedBy } from './Field';
import { Currencies, type CurrencyCode, DefaultCurrency } from '@/constants/currencies';
import { formatCurrency } from '@/utils/format';
import { cx } from '@/utils/cx';

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
  const id = String(name);

  const cur = Currencies[currency] ?? Currencies[DefaultCurrency];
  const hint = description ?? (value != null && Number.isFinite(value) ? `≈ ${formatCurrency(value, currency)}` : undefined);

  return (
    <Field label={label} required={required} description={hint} error={error} className={className} htmlFor={id}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <div className={showSymbol ? cx('input-affix input-affix--prefix', error && 'input-affix--invalid', disabled && 'input-affix--disabled') : undefined}>
            {showSymbol && (
              <span className="input-affix__prefix" aria-hidden="true">
                {cur.symbol}
              </span>
            )}
            <NumberFieldPrimitive.Root
              className="number-field"
              locale="es-PE"
              value={typeof field.value === 'number' ? field.value : null}
              disabled={disabled}
              onValueChange={(v) => field.onChange((v ?? 0) as never)}
            >
              <NumberFieldPrimitive.Input
                id={id}
                className="input tabular"
                aria-invalid={!!error || undefined}
                aria-required={required || undefined}
                aria-describedby={fieldDescribedBy(id, hint, error)}
                style={{ textAlign: 'right' }}
              />
            </NumberFieldPrimitive.Root>
          </div>
        )}
      />
    </Field>
  );
}
