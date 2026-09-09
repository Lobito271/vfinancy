import { useEffect, useMemo, useState } from 'react';
import { useWatch, useFormContext } from 'react-hook-form';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import {
  Form,
  CustomerSelectField,
  DateField,
  NumberField,
  TextareaField,
  LineItemsEditor,
  type SaleLineItemFormValues,
  type ProductLineOption,
} from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useCreateSale } from '@/features/sales/hooks/useSales';
import { useProducts } from '@/features/products/hooks/useProducts';
import { settingsService } from '@/services/settings';
import { treasuryService } from '@/services/treasury';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';
import { formatCurrency } from '@/utils/format';
import { cx } from '@/utils/cx';

const SaleFormSchema = z
  .object({
    customerId: z.string().min(1, 'Seleccione un cliente'),
    date: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
    dueDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido').optional().or(z.literal('')),
    exchangeRate: z.number().min(0.01, 'Tipo de cambio inválido'),
    notes: z.string().optional(),
    items: z
      .array(
        z.object({
          productId: z.string().min(1, 'Seleccione un producto'),
          quantity: z.number().positive('Cantidad debe ser mayor a 0'),
          unitPrice: z.number().min(0.01, 'Ingrese el precio de venta'),
          discountPercent: z.number().min(0).max(100),
          discountAmount: z.number().min(0),
          taxRate: z.number().min(0).max(100),
          taxAmount: z.number().min(0),
          costSnapshot: z.number().min(0),
          description: z.string(),
        }),
      )
      .min(1, 'Agregue al menos una línea'),
  })
  .superRefine((data, ctx) => {
    if (data.dueDate && data.date && data.dueDate < data.date) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['dueDate'],
        message: 'La fecha de vencimiento debe ser igual o posterior a la fecha de emisión.',
      });
    }
  });

type SaleFormValues = z.infer<typeof SaleFormSchema>;

const emptyLine = (): SaleLineItemFormValues => ({
  productId: '',
  quantity: 1,
  unitPrice: 0,
  discountPercent: 0,
  discountAmount: 0,
  taxRate: 0,
  taxAmount: 0,
  costSnapshot: 0,
  description: '',
});

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

function ExchangeRateSeed({ rate }: { rate: number | undefined }) {
  const { setValue } = useFormContext();
  useEffect(() => {
    if (rate != null && rate > 0) {
      setValue('exchangeRate', rate, { shouldValidate: true });
    }
  }, [rate, setValue]);
  return null;
}

function SaleFinancialSummary({ productMeta }: { productMeta: Map<string, { label: string; costUSD: number }> }) {
  const { control } = useFormContext();
  const items = useWatch({ control, name: 'items' }) as SaleFormValues['items'] | undefined;
  const exchangeRate = (useWatch({ control, name: 'exchangeRate' }) as number) ?? 0;

  const subtotal = round2(
    (items ?? []).reduce((s, it) => s + (it.unitPrice ?? 0) * (it.quantity ?? 0), 0),
  );
  const discount = round2(
    (items ?? []).reduce(
      (s, it) => s + (it.unitPrice ?? 0) * (it.quantity ?? 0) * ((it.discountPercent ?? 0) / 100),
      0,
    ),
  );
  const tax = round2(
    (items ?? []).reduce((s, it) => {
      const base = (it.unitPrice ?? 0) * (it.quantity ?? 0) * (1 - (it.discountPercent ?? 0) / 100);
      return s + base * ((it.taxRate ?? 0) / 100);
    }, 0),
  );
  const total = round2(subtotal - discount + tax);

  const totalCostPEN = round2(
    (items ?? []).reduce((s, it) => {
      const meta = productMeta.get(it.productId);
      const costUSD = meta?.costUSD ?? 0;
      return s + costUSD * (it.quantity ?? 0) * exchangeRate;
    }, 0),
  );
  const profit = round2(total - totalCostPEN);

  return (
    <div className="stack">
      <div className="fact-grid">
        <div>
          <div className="fact-grid__label">Subtotal</div>
          <div className="fact-grid__value">{formatCurrency(subtotal)}</div>
        </div>
        <div>
          <div className="fact-grid__label">IGV</div>
          <div className="fact-grid__value">{formatCurrency(tax)}</div>
        </div>
        <div>
          <div className="fact-grid__label">Total</div>
          <div className="fact-grid__value">{formatCurrency(total)}</div>
        </div>
        <div>
          <div className="fact-grid__label">Utilidad</div>
          <div className={`fact-grid__value ${profit < 0 ? 'text-destructive' : 'text-success'}`}>
            {formatCurrency(profit)}
          </div>
        </div>
      </div>
    </div>
  );
}

interface SaleFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SaleFormDialog({ open, onOpenChange }: SaleFormDialogProps) {
  const create = useCreateSale();
  const push = useNotificationStore((s) => s.push);
  const productsQuery = useProducts();
  const taxesQuery = useQuery({
    queryKey: ['catalog', 'taxes'],
    queryFn: () => settingsService.getTaxes(),
    staleTime: 5 * 60 * 1000,
  });
  const rateQuery = useQuery({
    queryKey: queryKeys.treasury.exchangeRate('USD', 'PEN'),
    queryFn: () => treasuryService.getExchangeRate('USD', 'PEN'),
    staleTime: 60 * 1000,
  });
  const prefsQuery = useQuery({
    queryKey: queryKeys.settings.preferences,
    queryFn: () => settingsService.getPreferences(),
    staleTime: 5 * 60 * 1000,
  });
  const isContingency = rateQuery.isError || rateQuery.data == null;
  const fallbackRate = prefsQuery.data?.fallbackExchangeRate ?? 0;

  const taxRateById = useMemo(
    () => new Map((taxesQuery.data ?? []).map((t) => [t.id, t.defaultRate])),
    [taxesQuery.data],
  );

  const productOptions = useMemo<ProductLineOption[]>(
    () =>
      (productsQuery.data?.items ?? []).map((p) => ({
        value: p.id,
        label: `${p.sku} — ${p.description}`,
        unitCost: p.costUSD,
        salePrice: 0,
        taxRate: p.taxId ? (taxRateById.get(p.taxId) ?? 0) : 0,
      })),
    [productsQuery.data, taxRateById],
  );

  const productMeta = useMemo(
    () => new Map(productOptions.map((p) => [p.value, { label: p.label, costUSD: p.unitCost }])),
    [productOptions],
  );

  const defaults = useMemo<SaleFormValues>(
    () => ({ customerId: '', date: today(), dueDate: '', exchangeRate: 0, notes: '', items: [emptyLine()] }),
    [],
  );

  const handleSubmit = async (values: SaleFormValues) => {
    try {
      const items = values.items.map((it) => {
        const base = it.unitPrice * it.quantity;
        const discountAmount = round2(base * (it.discountPercent / 100));
        const taxAmount = round2((base - discountAmount) * (it.taxRate / 100));
        const meta = productMeta.get(it.productId);
        const costUSD = meta?.costUSD ?? 0;
        return {
          productId: it.productId,
          quantity: it.quantity,
          unitPrice: it.unitPrice,
          discountPercent: it.discountPercent,
          discountAmount,
          taxRate: it.taxRate,
          taxAmount,
          costSnapshot: round2(costUSD * values.exchangeRate),
          description: meta?.label ?? '',
        };
      });
      await create.mutateAsync({
        customerId: values.customerId,
        date: values.date,
        dueDate: values.dueDate ?? '',
        exchangeRate: values.exchangeRate,
        notes: values.notes ?? '',
        items,
      });
      push({ title: 'Venta registrada', variant: 'success' });
      onOpenChange(false);
    } catch (err: unknown) {
      console.error('[SaleFormDialog] submit failed:', err);
      const msg = err instanceof Error ? err.message : String(err);
      push({
        title: 'No se pudo registrar la venta',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const [step, setStep] = useState(0);
  const steps = [
    { title: 'Cliente', description: 'Identifica al comprador y la fecha.' },
    { title: 'Carrito', description: 'Agrega los productos de la venta.' },
    { title: 'Pago', description: 'Revisa el tipo de cambio y confirma.' },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="xl">
        <DialogHeader>
          <DialogTitle>Nueva venta</DialogTitle>
          <DialogDescription>Registra un documento de venta para un cliente.</DialogDescription>
        </DialogHeader>
        <ol className="setup-steps stepper--sale" aria-label="Pasos de la venta">
          {steps.map((item, i) => (
            <li key={item.title} className={cx('setup-step', i === step && 'setup-step--active')}>
              <span className="setup-step__number">{i + 1}</span>
              <span><strong>{item.title}</strong><small>{item.description}</small></span>
            </li>
          ))}
        </ol>

        <Form schema={SaleFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {(form) => (
            <>
              <DialogBody>
                {step === 0 && (
                  <div className="form-grid">
                    <CustomerSelectField name="customerId" label="Cliente" required />
                    <DateField name="date" label="Fecha de venta" required />
                    <DateField
                      name="dueDate"
                      label="Fecha de vencimiento"
                      description="Obligatoria en ventas al crédito. No puede ser anterior a la fecha de venta."
                    />
                  </div>
                )}
                {step === 1 && (
                  <>
                    <LineItemsEditor products={productOptions} isSale currency="PEN" />
                  </>
                )}
                {step === 2 && (
                  <>
                    <NumberField
                      name="exchangeRate"
                      label="Tipo de cambio (USD→PEN)"
                      required
                      min={0.01}
                      step={0.001}
                      description={
                        isContingency && fallbackRate > 0
                          ? `Sin conexión a la API. Se aplica el tipo de cambio de respaldo (${fallbackRate}).`
                          : 'Valor de la moneda extranjera en soles.'
                      }
                    />
                    {isContingency && (
                      <div className="banner banner--warning" role="status">
                        <strong>Modo Contingencia / Sin conexión</strong> — no se pudo obtener el tipo
                        de cambio en tiempo real; se usa el valor de respaldo.
                      </div>
                    )}
                    <ExchangeRateSeed rate={rateQuery.data ?? (isContingency ? fallbackRate : undefined)} />
                    <SaleFinancialSummary productMeta={productMeta} />
                    <TextareaField name="notes" label="Notas" rows={2} />
                  </>
                )}
              </DialogBody>
              <DialogFooter>
                <Button variant="ghost" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
                  Cancelar
                </Button>
                {step > 0 && (
                  <Button variant="outline" type="button" onClick={() => setStep((s) => s - 1)} disabled={create.isPending}>
                    <ArrowLeft /> Atrás
                  </Button>
                )}
                {step < steps.length - 1 ? (
                  <Button
                    type="button"
                    onClick={async () => {
                      const fields: Array<keyof SaleFormValues> =
                        step === 0 ? ['customerId', 'date', 'dueDate'] : ['items'];
                      if (await form.trigger(fields)) setStep((s) => s + 1);
                    }}
                  >
                    Continuar <ArrowRight />
                  </Button>
                ) : (
                  <Button type="submit" loading={create.isPending}>
                    Guardar
                  </Button>
                )}
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
