import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import {
  Form,
  CustomerSelectField,
  DateField,
  TextareaField,
  LineItemsEditor,
  type SaleLineItemFormValues,
  type ProductLineOption,
} from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useCreateSale } from '@/features/sales/hooks/useSales';
import { useProducts } from '@/features/products/hooks/useProducts';
import { settingsService } from '@/services/settings';
import { useNotificationStore } from '@/stores/notification';

const SaleFormSchema = z.object({
  customerId: z.string().min(1, 'Seleccione un cliente'),
  date: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
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
        costSnapshot: z.number().min(0),
        description: z.string(),
      }),
    )
    .min(1, 'Agregue al menos una línea'),
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

  const taxRateById = useMemo(
    () => new Map((taxesQuery.data ?? []).map((t) => [t.id, t.defaultRate])),
    [taxesQuery.data],
  );

  const productOptions = useMemo<ProductLineOption[]>(
    () =>
      (productsQuery.data?.items ?? []).map((p) => ({
        value: p.id,
        label: `${p.sku} — ${p.description}`,
        unitCost: p.cost,
        salePrice: p.salePrice,
        taxRate: p.taxId ? (taxRateById.get(p.taxId) ?? 0) : 0,
      })),
    [productsQuery.data, taxRateById],
  );

  const productMeta = useMemo(
    () => new Map(productOptions.map((p) => [p.value, { label: p.label, cost: p.unitCost }])),
    [productOptions],
  );

  const defaults = useMemo<SaleFormValues>(
    () => ({ customerId: '', date: today(), notes: '', items: [emptyLine()] }),
    [],
  );

  const handleSubmit = (values: SaleFormValues) => {
    const items = values.items.map((it) => {
      const base = it.unitPrice * it.quantity;
      const discountAmount = round2(base * (it.discountPercent / 100));
      const taxAmount = round2((base - discountAmount) * (it.taxRate / 100));
      const meta = productMeta.get(it.productId);
      return {
        productId: it.productId,
        quantity: it.quantity,
        unitPrice: it.unitPrice,
        discountPercent: it.discountPercent,
        discountAmount,
        taxRate: it.taxRate,
        taxAmount,
        costSnapshot: meta?.cost ?? 0,
        description: meta?.label ?? '',
      };
    });
    create.mutate(
      { customerId: values.customerId, date: values.date, notes: values.notes ?? '', items },
      {
        onSuccess: () => {
          push({ title: 'Venta registrada', variant: 'success' });
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo registrar la venta',
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
          <DialogTitle>Nueva venta</DialogTitle>
          <DialogDescription>Registra un documento de venta para un cliente.</DialogDescription>
        </DialogHeader>

        <Form schema={SaleFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <CustomerSelectField name="customerId" label="Cliente" required />
                  <DateField name="date" label="Fecha de venta" required />
                </div>
                <LineItemsEditor products={productOptions} isSale />
                <TextareaField name="notes" label="Notas" rows={2} />
              </div>
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
