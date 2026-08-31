import { useEffect, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useFormContext } from 'react-hook-form';
import { z } from 'zod';
import {
  Form,
  CustomerSelectField,
  SupplierSelectField,
  DateField,
  MoneyField,
  TextareaField,
  SelectField,
  NumberField,
  type SelectOption,
  LineItemsEditor,
  type LineItemFormValues,
  type ProductLineOption,
} from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useCreateCustomerOrder } from '@/features/purchasing/hooks/usePurchases';
import { useCreditCards } from '@/features/treasury/hooks/useTreasury';
import { OrderFinancialPreview } from '@/features/purchasing/components/OrderFinancialPreview';
import { useProducts } from '@/features/products/hooks/useProducts';
import { settingsService } from '@/services/settings';
import { treasuryService } from '@/services/treasury';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';

const CustomerOrderFormSchema = z
  .object({
    customerId: z.string().min(1, 'Seleccione un cliente'),
    supplierId: z.string().min(1, 'Seleccione un proveedor'),
    creditCardId: z.string().min(1, 'Seleccione la tarjeta de crédito'),
    orderDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
    supplierOrderNumber: z.string().max(100, 'Máximo 100 caracteres').optional(),
    exchangeRate: z.number().min(0.01, 'Tipo de cambio requerido'),
    salePricePEN: z.number().min(0.01, 'El precio de venta debe ser mayor a 0'),
    anticipo: z.number().min(0, 'Debe ser >= 0'),
    anticipoDate: z.string().optional(),
    notes: z.string().optional(),
    items: z
      .array(
        z.object({
          productId: z.string().min(1, 'Seleccione un producto'),
          quantity: z.number().positive('Cantidad debe ser mayor a 0'),
          unitPrice: z.number().min(0, 'Debe ser >= 0'),
          discountPercent: z.number().min(0).max(100),
          discountAmount: z.number().min(0),
          taxRate: z.number().min(0).max(100),
          taxAmount: z.number().min(0),
          description: z.string(),
        }),
      )
      .min(1, 'Agregue al menos una línea'),
  })
  .refine((v) => !(v.anticipo > 0 && !v.anticipoDate), { message: 'Indique la fecha del anticipo', path: ['anticipoDate'] })
  .refine((v) => v.anticipo <= v.salePricePEN, {
    message: 'El anticipo no puede superar el precio de venta',
    path: ['anticipo'],
  });

type CustomerOrderFormValues = z.infer<typeof CustomerOrderFormSchema>;

const emptyLine = (): LineItemFormValues => ({
  productId: '',
  quantity: 1,
  unitPrice: 0,
  discountPercent: 0,
  discountAmount: 0,
  taxRate: 0,
  taxAmount: 0,
  description: '',
});

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

interface CreateCustomerOrderDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
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

export function CreateCustomerOrderDialog({ open, onOpenChange }: CreateCustomerOrderDialogProps) {
  const create = useCreateCustomerOrder();
  const push = useNotificationStore((s) => s.push);
  const productsQuery = useProducts();
  const cardsQuery = useCreditCards();
  const rateQuery = useQuery({
    queryKey: queryKeys.treasury.exchangeRate('USD', 'PEN'),
    queryFn: () => treasuryService.getExchangeRate('USD', 'PEN'),
    staleTime: 60 * 1000,
  });
  const taxesQuery = useQuery({
    queryKey: ['catalog', 'taxes'],
    queryFn: () => settingsService.getTaxes(),
    staleTime: 5 * 60 * 1000,
  });

  const cardOptions = useMemo<SelectOption[]>(
    () =>
      (cardsQuery.data ?? []).map((c) => ({
        value: c.id,
        label: `${c.issuer} •••• ${c.lastFour} (${c.currencyCode})`,
      })),
    [cardsQuery.data],
  );

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
        salePrice: p.salePrice,
        taxRate: p.taxId ? (taxRateById.get(p.taxId) ?? 0) : 0,
      })),
    [productsQuery.data, taxRateById],
  );

  const labelByProduct = useMemo(() => new Map(productOptions.map((p) => [p.value, p.label])), [productOptions]);

  const defaults = useMemo<CustomerOrderFormValues>(
    () => ({
      customerId: '',
      supplierId: '',
      creditCardId: '',
      orderDate: today(),
      supplierOrderNumber: '',
      exchangeRate: 0,
      salePricePEN: 0,
      anticipo: 0,
      anticipoDate: today(),
      notes: '',
      items: [emptyLine()],
    }),
    [],
  );

  const handleSubmit = (values: CustomerOrderFormValues) => {
    const items = values.items.map((it) => {
      const base = it.unitPrice * it.quantity;
      const discountAmount = round2(base * (it.discountPercent / 100));
      const taxAmount = round2((base - discountAmount) * (it.taxRate / 100));
      return {
        productId: it.productId,
        quantity: it.quantity,
        unitPrice: it.unitPrice,
        discountPercent: it.discountPercent,
        discountAmount,
        taxRate: it.taxRate,
        taxAmount,
        description: labelByProduct.get(it.productId) ?? '',
      };
    });
    const costUSD = round2(items.reduce((sum, it) => sum + it.unitPrice * it.quantity - it.discountAmount, 0));
    create.mutate(
      {
        supplierId: values.supplierId,
        customerId: values.customerId,
        creditCardId: values.creditCardId,
        orderDate: values.orderDate,
        supplierOrderNumber: values.supplierOrderNumber ?? '',
        costUSD,
        salePricePEN: round2(values.salePricePEN),
        anticipo: round2(values.anticipo),
        anticipoDate: values.anticipo > 0 ? (values.anticipoDate ?? today()) : today(),
        exchangeRate: values.exchangeRate,
        notes: values.notes ?? '',
        items,
      },
      {
        onSuccess: () => {
          push({ title: 'Pedido de cliente creado', variant: 'success' });
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo crear el pedido',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          });
        },
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="xl">
        <DialogHeader>
          <DialogTitle>Nuevo pedido de cliente</DialogTitle>
          <DialogDescription>
            Pedido de compra para un cliente específico: costo en PEN, precio de venta proyectado y anticipo inicial.
          </DialogDescription>
        </DialogHeader>

        <Form schema={CustomerOrderFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <ExchangeRateSeed rate={rateQuery.data} />
              <div className="stack dialog-scroll">
                <div className="form-grid form-grid--3">
                  <CustomerSelectField name="customerId" label="Cliente" required />
                  <SupplierSelectField name="supplierId" label="Proveedor / Importador" required />
                  <DateField name="orderDate" label="Fecha de pedido" required />
                </div>
                <TextareaField name="supplierOrderNumber" label="N° Orden del Proveedor" rows={1} placeholder="Opcional — número de orden del proveedor" />
                <div className="form-grid">
                  <SelectField
                    name="creditCardId"
                    label="Tarjeta de crédito (pago en USD)"
                    description="La tarjeta usada para pagar al proveedor en dólares"
                    placeholder={cardsQuery.isLoading ? 'Cargando tarjetas…' : 'Seleccione la tarjeta…'}
                    options={cardOptions}
                    loading={cardsQuery.isLoading}
                    required
                  />
                  <NumberField
                    name="exchangeRate"
                    label="Tipo de cambio (USD→PEN)"
                    description={rateQuery.isLoading ? 'Cargando…' : 'T.C. de referencia (editable en el panel de costos)'}
                    min={0.01}
                    step={0.01}
                    readOnly
                  />
                </div>
                <OrderFinancialPreview />
                <LineItemsEditor products={productOptions} currency="USD" />
                <div className="form-grid form-grid--3">
                  <MoneyField
                    name="salePricePEN"
                    label="Precio de venta (PEN)"
                    description="Precio de venta esperado del pedido"
                    required
                  />
                  <MoneyField name="anticipo" label="Anticipo inicial" description="Abono inicial del cliente" />
                  <DateField name="anticipoDate" label="Fecha del anticipo" />
                </div>
                <TextareaField name="notes" label="Notas" rows={2} />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={create.isPending} disabled={!formState.isValid}>
                  Guardar pedido
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
