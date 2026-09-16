import { useEffect, useMemo, useRef, useState } from 'react';
import { useFieldArray, useFormContext, useWatch, type Path } from 'react-hook-form';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, ArrowRight, Trash2, Plus } from 'lucide-react';
import { z } from 'zod';
import {
  Form,
  DateField,
  TextareaField,
  SelectField,
  NumberField,
  TextField,
  type CreateSelectOption,
  type SelectOption,
} from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
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
import { CreditCardFormDialog } from '@/features/treasury/components/CreditCardFormDialog';
import { CreateCustomerDialog } from '@/features/customers/components/CreateCustomerDialog';
import { useProducts } from '@/features/products/hooks/useProducts';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { formatCurrency } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

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
    number: z.string().trim().optional(),
    customerId: z.string(),
    creditCardId: z.string().min(1, 'Seleccione la tarjeta de crédito'),
    exchangeRate: z.number().min(0.01, 'Tipo de cambio inválido'),
    orderDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
    expectedDate: z.string(),
    notes: z.string().optional(),
    items: z.array(lineSchema).min(1, 'Agregue al menos una línea'),
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

function CustomerField({ customers, createCustomer }: { customers: SelectOption[]; createCustomer?: CreateSelectOption }) {
  return (
    <SelectField
      name="customerId"
      label="Cliente"
      description="Opcional: pedido para un cliente."
      clearable
      options={customers}
      createOption={createCustomer}
      placeholder="Sin cliente…"
    />
  );
}

interface ProductCostOption extends SelectOption {
  unitCost: number;
  salePrice: number;
}

function OrderDataStep({ cardOptions, customerOptions, cardsQuery, rateQuery, lotsQuery, cardCreateOption, customerCreateOption }: {
  cardOptions: SelectOption[];
  customerOptions: SelectOption[];
  cardsQuery: { isLoading: boolean };
  rateQuery: { data?: { rate?: number; isFallback?: boolean }; isLoading: boolean };
  lotsQuery: { data?: { items: { id: string; status: string; code: string; description: string }[] } };
  cardCreateOption: CreateSelectOption;
  customerCreateOption: CreateSelectOption;
}) {
  return (
    <div className="stack">
      <ExchangeRateSeed rate={rateQuery.data?.rate} />
      <CustomerField customers={customerOptions} createCustomer={customerCreateOption} />
      <div className="form-grid">
        <SelectField
          name="creditCardId"
          label="Tarjeta de crédito (pago en USD)"
          required
          description="Pago obligatorio con tarjeta"
          placeholder={cardsQuery.isLoading ? 'Cargando tarjetas…' : 'Seleccione la tarjeta…'}
          options={cardOptions}
          loading={cardsQuery.isLoading}
          createOption={cardCreateOption}
        />
      </div>
      {cardOptions.length === 0 && !cardsQuery.isLoading && (
        <p className="field__error" role="alert">Cree una tarjeta en Tesorería</p>
      )}
      <div className="form-grid">
        <TextField
          name="number"
          label="Número de orden"
          description="Opcional: déjalo vacío para generarlo automáticamente. No se puede cambiar después."
        />
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
    </div>
  );
}

function ItemsStep({ products }: { products: ProductCostOption[] }) {
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

function ReviewStep() {
  return (
    <div className="stack">
      <TextareaField name="notes" label="Notas" rows={2} />
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
  const [cardCreateOpen, setCardCreateOpen] = useState(false);
  const assignCard = useRef<(id: string) => void>(() => {});
  const [customerCreateOpen, setCustomerCreateOpen] = useState(false);
  const assignCustomer = useRef<(id: string) => void>(() => {});
  const cardCreateOption: CreateSelectOption = {
    label: 'Crear nueva tarjeta…',
    onSelect: (assign) => {
      assignCard.current = assign;
      setCardCreateOpen(true);
    },
  };
  const customerCreateOption: CreateSelectOption = {
    label: 'Crear nuevo cliente…',
    onSelect: (assign) => {
      assignCustomer.current = assign;
      setCustomerCreateOpen(true);
    },
  };
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
      number: '',
      customerId: '',
      creditCardId: '',
      exchangeRate: 0,
      orderDate: today(),
      expectedDate: today(),
      notes: '',
      items: [emptyLine()],
    }),
    [],
  );

  useEffect(() => {
    if (!open) setLotId('');
  }, [open]);

  const steps = [
    { title: 'Orden', description: 'Datos generales de la compra.' },
    { title: 'Ítems', description: 'Productos, cantidades y precios.' },
    { title: 'Revisar', description: 'Notas y confirmación.' },
  ];

  const [step, setStep] = useState(0);

  const handleSubmit = (values: PurchaseFormValues) => {
    create.mutate(
      {
        number: values.number,
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
          <DialogTitle>{steps[step].title}</DialogTitle>
        </DialogHeader>

        <Form schema={PurchaseFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {(form) => (
            <>
              <DialogBody>
                {step === 0 && (
                  <OrderDataStep
                    cardOptions={cardOptions}
                    customerOptions={customerOptions}
                    cardsQuery={cardsQuery}
                    rateQuery={rateQuery}
                    lotsQuery={lotsQuery}
                    cardCreateOption={cardCreateOption}
                    customerCreateOption={customerCreateOption}
                  />
                )}
                {step === 1 && <ItemsStep products={productOptions} />}
                {step === 2 && <ReviewStep />}
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
                      const fields: Array<Path<PurchaseFormValues>> =
                        step === 0
                          ? ['creditCardId', 'customerId', 'orderDate', 'exchangeRate']
                          : ['items'];
                      if (await form.trigger(fields)) setStep((s) => s + 1);
                    }}
                  >
                    Continuar <ArrowRight />
                  </Button>
                ) : (
                  <Button type="submit" loading={create.isPending} disabled={!form.formState.isValid}>
                    Guardar
                  </Button>
                )}
              </DialogFooter>
            </>
          )}
        </Form>

        <CreditCardFormDialog
          open={cardCreateOpen}
          onOpenChange={setCardCreateOpen}
          editCard={null}
          onCreated={(card) => assignCard.current(card.id)}
        />
        <CreateCustomerDialog
          open={customerCreateOpen}
          onOpenChange={setCustomerCreateOpen}
          onCreated={(id) => assignCustomer.current(id)}
        />
      </DialogContent>
    </Dialog>
  );
}
