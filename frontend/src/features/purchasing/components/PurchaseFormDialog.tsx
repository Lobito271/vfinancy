import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import {
  Form,
  SupplierSelectField,
  DateField,
  TextareaField,
  LineItemsEditor,
  type LineItemFormValues,
  type ProductLineOption,
} from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useCreatePurchase } from '@/features/purchasing/hooks/usePurchases';
import { useProducts } from '@/features/products/hooks/useProducts';
import { settingsService } from '@/services/settings';
import { useNotificationStore } from '@/stores/notification';

const PurchaseFormSchema = z.object({
  supplierId: z.string().min(1, 'Seleccione un proveedor'),
  orderDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
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
});

type PurchaseFormValues = z.infer<typeof PurchaseFormSchema>;

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

interface PurchaseFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PurchaseFormDialog({ open, onOpenChange }: PurchaseFormDialogProps) {
  const create = useCreatePurchase();
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

  const labelByProduct = useMemo(
    () => new Map(productOptions.map((p) => [p.value, p.label])),
    [productOptions],
  );

  const defaults = useMemo<PurchaseFormValues>(
    () => ({ supplierId: '', orderDate: today(), notes: '', items: [emptyLine()] }),
    [],
  );

  const handleSubmit = (values: PurchaseFormValues) => {
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
    create.mutate(
      { supplierId: values.supplierId, orderDate: values.orderDate, notes: values.notes ?? '', items },
      {
        onSuccess: () => {
          push({ title: 'Orden de compra creada', variant: 'success' });
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
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <SupplierSelectField name="supplierId" label="Proveedor" required />
                  <DateField name="orderDate" label="Fecha de orden" required />
                </div>
                <LineItemsEditor products={productOptions} />
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
