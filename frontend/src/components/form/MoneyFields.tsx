import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useFormContext, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { Field, fieldDescribedBy } from './Field';
import { Input } from '@/components/input';
import { Currencies, type CurrencyCode, DefaultCurrency } from '@/constants/currencies';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';
import { convertAmount, nextMoneyCurrency, parseAmount, roundMoney, type Currency } from '@/utils/money';
import { formatCurrency } from '@/utils/format';
import { cx } from '@/utils/cx';

interface MoneyFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  currency?: CurrencyCode;
  currencyField?: FieldPath<T>;
  exchangeRate?: number;
  className?: string;
  disabled?: boolean;
}

function formatAmount(value: number): string {
  return new Intl.NumberFormat('es-PE', {
    useGrouping: false,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function MoneyField<T extends FieldValues>({
  name,
  label,
  description,
  required,
  currency = DefaultCurrency,
  currencyField,
  exchangeRate,
  className,
  disabled,
}: MoneyFieldProps<T>) {
  const { control, formState, watch, setValue } = useFormContext<T>();
  const error = formState.errors[name]?.message as string | undefined;
  const id = String(name);
  const [displayCurrency, setDisplayCurrency] = useState<CurrencyCode>(currency);
  const [draft, setDraft] = useState<string | null>(null);
  const [focused, setFocused] = useState(false);
  const currencyValue = currencyField ? watch(currencyField) as CurrencyCode | undefined : undefined;
  const activeCurrency = currencyValue ?? displayCurrency;
  const suppliedRate = exchangeRate != null && Number.isFinite(exchangeRate) && exchangeRate > 0 ? exchangeRate : undefined;
  const rateQuery = useQuery({
    queryKey: queryKeys.treasury.exchangeRate('USD', 'PEN'),
    queryFn: () => wailsClient.latestExchangeRate(),
    staleTime: 60_000,
    enabled: !currencyField && suppliedRate == null,
  });
  const rate = suppliedRate ?? rateQuery.data?.rate;
  const hasRate = rate != null && Number.isFinite(rate) && rate > 0;

  useEffect(() => {
    setDisplayCurrency(currency);
  }, [currency]);

  const toggleCurrency = () => {
    const next = nextMoneyCurrency(activeCurrency);
    if (currencyField) {
      setValue(currencyField, next as never, { shouldDirty: true, shouldValidate: true });
    } else if (hasRate) {
      setDisplayCurrency(next);
    }
    setFocused(false);
    setDraft(null);
  };

  const symbol = Currencies[activeCurrency]?.symbol ?? Currencies.PEN.symbol;
  const nextCurrency = activeCurrency === 'USD' ? 'PEN' : 'USD';
  const currentRate = hasRate ? rate : 1;
  const hint = description;

  return (
    <Field label={label} required={required} description={hint} error={error} className={className} htmlFor={id}>
      <Controller
        control={control}
        name={name}
        render={({ field }) => {
          const baseAmount = typeof field.value === 'number' ? field.value : 0;
          const inputAmount = Number.isFinite(baseAmount)
            ? currencyField
              ? baseAmount
              : convertAmount(baseAmount, currency as Currency, activeCurrency as Currency, currentRate)
            : Number.NaN;
          const value = focused && draft !== null ? draft : Number.isFinite(inputAmount) ? formatAmount(inputAmount) : '';
          const valueHint = !description && Number.isFinite(inputAmount)
            ? `≈ ${formatCurrency(inputAmount, activeCurrency)}`
            : undefined;

          return (
            <>
              <div className={cx('input-affix', error && 'input-affix--invalid', disabled && 'input-affix--disabled')}>
                <button
                  type="button"
                  className="input-affix__currency"
                  onClick={toggleCurrency}
                  disabled={disabled || (!currencyField && !hasRate)}
                  aria-label={`Cambiar moneda a ${nextCurrency === 'USD' ? 'dólares' : 'soles'}`}
                  title={
                    !currencyField && !hasRate
                      ? rateQuery.error
                        ? 'No hay tipo de cambio disponible'
                        : 'Cargando tipo de cambio…'
                      : undefined
                  }
                >
                  {symbol}
                </button>
                <Input
                  ref={field.ref}
                  id={id}
                  type="text"
                  inputMode="decimal"
                  autoComplete="off"
                  className="tabular money-input__amount"
                  value={value}
                  disabled={disabled}
                  invalid={!!error}
                  aria-required={required || undefined}
                  aria-describedby={fieldDescribedBy(id, hint ?? valueHint, error)}
                  onFocus={() => {
                    setFocused(true);
                    setDraft(formatAmount(inputAmount));
                  }}
                  onChange={(event) => {
                    const raw = event.target.value;
                    setDraft(raw);
                    if (!raw.trim()) {
                      field.onChange(0);
                      return;
                    }
                    const parsed = parseAmount(raw);
                    if (parsed == null) {
                      field.onChange(Number.NaN);
                      return;
                    }
                    const baseValue = currencyField
                      ? parsed
                      : convertAmount(parsed, activeCurrency as Currency, currency as Currency, currentRate);
                    field.onChange(roundMoney(baseValue));
                  }}
                  onBlur={() => {
                    field.onBlur();
                    setFocused(false);
                    setDraft(null);
                  }}
                />
              </div>
              {valueHint && !error && <p id={`${id}-hint`} className="field-hint">{valueHint}</p>}
            </>
          );
        }}
      />
    </Field>
  );
}
