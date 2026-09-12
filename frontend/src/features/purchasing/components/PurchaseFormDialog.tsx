import { useEffect, useMemo, useState } from 'react';
import { useFieldArray, useFormContext, useWatch, type Path } from 'react-hook-form';import { useQuery } from '@tanstack/react-query';
import { Trash2, Plus } from 'lucide-react';
import { z } from 'zod';
import {
  Form,
  DateField,
  TextareaField,
  SelectField,
  NumberField,
  TextField,
  type SelectOption,
} from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Badge } from '@/components/badge';
import { Input, Label } from '@/components/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useCreatePurchase } from '@/features/purchasing/hooks/usePurchases';
import { useCreditCards } from '@/features/treasury/hooks/useTreasury';
import { useProducts } from '@/features/products/hooks/useProducts';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { formatCurrency } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

type OrderType = 'general' | 'customer';

const lineSchema = z
  .object({
    productId: z.string(),
    description: z.string().trim(),
    quantity: z.number().int('Debe ser entero').positive('Cantidad debe ser mayor a 0'),
    unitPrice: z.number().positive('Costo USD debe ser mayor a 0'),
    salePricePen: z.number().min(0, 'El precio de venta no puede ser negativo'),
  })
  .refine((l) => l.productId !== '' || l.description !== '', {
    message: 'Indique el producto o su descripción',
    path: ['description'],
  });

const PurchaseFormSchema = z
  .object({
    orderType: z.enum(['general', 'customer']),
    customerId: z.string(),
    creditCardId: z.string().min(1, 'Seleccione la tarjeta de crédito'),
    exchangeRate: z.number().min(0.01, 'Tipo de cambio inválido'),
    orderDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
    expectedDate: z.string(),
    notes: z.string().optional(),
    items: z.array(lineSchema).min(1, 'Agregue al menos una línea'),
  })
  .refine((v) => v.orderType !== 'customer' || v.customerId !== '', {
    message: 'Seleccione el cliente',
    path: ['customerId'],
  });

type PurchaseFormValues = z.infer<typeof PurchaseFormSchema>;

const emptyLine = () => ({ productId: '', description: '', quantity: 1, unitPrice: 0, salePricePen: 0 });

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

interface PurchaseFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function ExchangeRateSeed({ rate }: { rate: number | undefined }) {
  const { setValue } = useFormContext<PurchaseFormValues>();
  useEffect(() => {
    if (rate != null && rate > 0) setValue('exchangeRate', rate, { shouldValidate: true });
  }, [rate, setValue]);
  return null;
}

function OrderTypePicker() {
  const { watch, setValue } = useFormContext<PurchaseFormValues>();
  const value = watch('orderType');
  const options: { value: OrderType; label: string; hint: string }[] = [
    { value: 'general', label: 'General (stock)', hint: 'Para vender en el almacén' },
    { value: 'customer', label: 'Cliente a pedido', hint: 'La mercadería va a un cliente' },
  ];
  const pick = (next: OrderType) => {
    if (next === value) return;
    if (next === 'general') {
      setValue('orderType', next);
      setValue('customerId', '', { shouldValidate: true });
    } else {
      setValue('orderType', next);
    }
  };
  return (
    <div className="field">
      <label className="label">Tipo de pedido</label>
      <div className="hstack hstack--sm" role="radiogroup" aria-label="Tipo de pedido">
        {options.map((o) => (
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={value === o.value}
            data-checked={value === o.value || undefined}
            className="segment-option"
            onClick={() => pick(o.value)}
            style={{ flex: 1 }}
          >
            <strong>{o.label}</strong>
            <small>{o.hint}</small>
          </button>
        ))}
      </div>
    </div>
  );
}

function CustomerField({ customers }: { customers: SelectOption[] }) {
  const orderType = useWatch<PurchaseFormValues, 'orderType'>({ name: 'orderType' });
  if (orderType !== 'customer') return null;
  return (
    <SelectField
      name="customerId"
      label="Cliente"
      required
      options={customers}
      placeholder="Seleccione el cliente…"
    />
  );
}

interface ProductCostOption extends SelectOption {
  unitCost: number;
  salePrice: number;
}
function PurchaseLines({ products }: { products: ProductCostOption[] }) {
  const { control, setValue } = useFormContext<PurchaseFormValues>();
  const { fields, append, remove } = useFieldArray<PurchaseFormValues, 'items'>({ control, name: 'items' });
  const rows = useWatch<PurchaseFormValues, 'items'>({ control, name: 'items' });

  const handleProduct = (index: number, productId: string) => {
    if (!productId) return;
    const p = products.find((o) => o.value === productId);
    if (!p) return;
    setValue(`items.${index}.description` as Path<PurchaseFormValues>, p.label);
    setValue(`items.${index}.unitPrice` as Path<PurchaseFormValues>, p.unitCost);
    setValue(`items.${index}.salePricePen` as Path<PurchaseFormValues>, p.salePrice);
  };

  return (
    <div className="stack">
      <div className="field">
        <label className="label">Ítems</label>
        <div className="stack">
          {fields.map((field, index) => (
            <div key={field.id} className="card" style={{ padding: '0.75rem' }}>
              <div className="hstack hstack--sm" style={{ justifyContent: 'space-between' }}>
                <strong>Ítem {index + 1}</strong>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Quitar ítem ${index + 1}`}
                  onClick={() => remove(index)}
                >
                  <Trash2 />
                </Button>
              </div>
              <div className="form-grid">
                <SelectField
                  name={`items.${index}.productId` as Path<PurchaseFormValues>}
                  label="Producto"
                  options={products}
                  placeholder="Nuevo (por descripción)…"
                  clearable
                  onChange={(v) => handleProduct(index, v)}
                />
                <TextField
                  name={`items.${index}.description` as Path<PurchaseFormValues>}
                  label="Descripción"
                  description="Si no eliges producto, se creará con esta descripción."
                />
              </div>
              <div className="form-grid">
                <NumberField name={`items.${index}.quantity` as Path<PurchaseFormValues>} label="Cantidad" required min={1} step={1} />
                <NumberField name={`items.${index}.unitPrice` as Path<PurchaseFormValues>} label="Costo (USD)" required min={0} step={0.01} />
                <NumberField
                  name={`items.${index}.salePricePen` as Path<PurchaseFormValues>}
                  label="Precio de venta (PEN)"
                  description="Precio sugerido al vender en soles."
                  min={0}
                  step={0.01}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
      <Button type="button" variant="outline" size="sm" onClick={() => append(emptyLine())}>
        <Plus /> Agregar ítem
      </Button>
      <div className="line-items-totals">
        <div className="totals-row totals-row--total">
          <span>Costo total (USD)</span>
          <span className="tabular">
            {rows != null && formatUsd(rows.reduce((s, r) => s + (r.unitPrice ?? 0) * (r.quantity ?? 0), 0))}
          </span>
        </div>
      </div>
    </div>
  );
}

function formatUsd(value: number): string {
  return formatCurrency(value, 'USD');
}

export function PurchaseFormDialog({ open, onOpenChange }: PurchaseFormDialogProps) {
  const create = useCreatePurchase();
  const push = useNotificationStore((s) => s.push);
  const [lotId, setLotId] = useState('');
  const productsQuery = useProducts();
  const cardsQuery = useCreditCards();
  const customersQuery = useQuery({
    queryKey: queryKeys.customers.options,
    queryFn: () => wailsClient.customerOptions(),
    enabled: open,
  });
  const rateQuery = useQuery({
    queryKey: queryKeys.treasury.exchangeRate('USD', 'PEN'),
    queryFn: () => wailsClient.latestExchangeRate(),
    staleTime: 60 * 1000,
  });
  const lotsQuery = useQuery({
    queryKey: queryKeys.importLots.list({ page: 1, pageSize: 100, search: '' }),
    queryFn: () => wailsClient.listImportLots({ page: 1, pageSize: 100 }, ''),
    enabled: open,
  });

  const cardOptions = useMemo<SelectOption[]>(
    () => (cardsQuery.data ?? []).map((c) => ({ value: c.id, label: `${c.issuer} •••• ${c.lastFour}` })),
    [cardsQuery.data],
  );

  const customerOptions = useMemo<SelectOption[]>(
    () => (customersQuery.data ?? []).map((c) => ({ value: c.id, label: c.businessName })),
    [customersQuery.data],
  );

  const productOptions = useMemo<ProductCostOption[]>(
    () => (productsQuery.data?.items ?? []).map((p) => ({ value: p.id, label: `${p.sku} — ${p.description}`, unitCost: p.costUsd, salePrice: p.salePrice })),
    [productsQuery.data],
  );

  const defaults = useMemo<PurchaseFormValues>(
    () => ({
      orderType: 'general',
      customerId: '',
      creditCardId: '',
      exchangeRate: 0,
      orderDate: today(),
      expectedDate: '',
      notes: '',
      items: [emptyLine()],
    }),
    [],
  );

  useEffect(() => {
    if (!open) setLotId('');
  }, [open]);

  const handleSubmit = (values: PurchaseFormValues) => {
    create.mutate(
      {
        orderType: values.orderType,
        customerId: values.customerId,
        creditCardId: values.creditCardId,
        orderDate: values.orderDate,
        expectedDate: values.expectedDate,
        exchangeRate: values.exchangeRate,
        notes: values.notes ?? '',
        items: values.items.map((it) => ({
          productId: it.productId,
          quantity: it.quantity,
          unitPrice: it.unitPrice,
          salePricePen: it.salePricePen,
          description: it.description,
        })),
      },
      {
        onSuccess: async (purchase) => {
          if (lotId) {
            try {
              const lot = await wailsClient.addToImportLot(lotId, [purchase.id]);
              if (lot.overLimit) {
                push({
                  title: 'El lote supera el tope aduanero',
                  description: `El lote ${lot.code} supera ${formatCurrency(lot.customsLimitUsd, 'USD')}.`,
                  variant: 'warning',
                });
              }
            } catch (err) {
              push({
                title: 'La orden se creó, pero no se pudo asignar al lote',
                description: err instanceof Error ? err.message : undefined,
                variant: 'destructive',
              });
            }
          }
          push({ title: 'Orden de compra creada', description: purchase.number, variant: 'success' });
          setLotId('');
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo crear la orden de compra',
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
          <DialogTitle>Nueva orden de compra</DialogTitle>
          <DialogDescription>Registra una orden de compra para un proveedor.</DialogDescription>
        </DialogHeader>

        <Form schema={PurchaseFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <DialogBody>
                <ExchangeRateSeed rate={rateQuery.data?.rate} />
                <OrderTypePicker />
                <CustomerField customers={customerOptions} />
                <div className="form-grid">
                  <SelectField
                    name="creditCardId"
                    label="Tarjeta de crédito (pago en USD)"
                    required
                    description="Pago obligatorio con tarjeta"
                    placeholder={cardsQuery.isLoading ? 'Cargando tarjetas…' : 'Seleccione la tarjeta…'}
                    options={cardOptions}
                    loading={cardsQuery.isLoading}
                  />
                </div>
                {cardOptions.length === 0 && !cardsQuery.isLoading && (
                  <p className="field__error" role="alert">Cree una tarjeta en Tesorería</p>
                )}
                <div className="form-grid">
                  <DateField name="orderDate" label="Fecha de pedido" required />
                  <DateField name="expectedDate" label="Fecha estimada" description="Opcional" />
                </div>
                <div className="form-grid">
                  <div className="field">
                    <Label htmlFor="purchase-unit-code">Unidad de medida</Label>
                    <Input id="purchase-unit-code" value="Unidad" disabled readOnly aria-label="Unidad de medida" />
                  </div>
                  <div className="field">
                    <Label htmlFor="purchase-lot">Lote de importación</Label>
                    <Select
                      items={[
                        { value: '', label: 'Sin lote' },
                        ...(lotsQuery.data?.items ?? [])
                          .filter((l) => l.status === 'active')
                          .map((l) => ({ value: l.id, label: `${l.code} · ${l.description || 'sin descripción'}` })),
                      ]}
                      value={lotId}
                      onValueChange={(v) => setLotId(v ?? '')}
                    >
                      <SelectTrigger id="purchase-lot" aria-label="Lote de importación">
                        <SelectValue placeholder="Sin lote" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">Sin lote</SelectItem>
                        {(lotsQuery.data?.items ?? [])
                          .filter((l) => l.status === 'active')
                          .map((l) => (
                            <SelectItem key={l.id} value={l.id}>
                              {l.code} · {l.description || 'sin descripción'}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="form-grid form-grid--wide">
                  <div className="stack stack--tight">
                    <div className="hstack hstack--sm">
                      <label className="input-label">Tipo de cambio (USD→PEN)</label>
                      {rateQuery.data?.isFallback && <Badge variant="warning">Modo contingencia</Badge>}
                    </div>
                    <NumberField name="exchangeRate" min={0.01} step={0.01} description={rateQuery.isLoading ? 'Cargando tipo de cambio…' : 'T.C. de referencia editable'} required />
                  </div>
                </div>
                <PurchaseLines products={productOptions} />
                <TextareaField name="notes" label="Notas" rows={2} />
              </DialogBody>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={create.isPending} disabled={!formState.isValid}>
                  Guardar
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
