import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { Form, TextField, NumberField, SelectField, MoneyField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { useCreateProduct, useUpdateProduct, useCategories, useBrands } from '@/features/products/hooks/useProducts';
import type { Product } from '@/types/domain';
import type { UnitDTO } from '@/services/wails-types';
import { catalogService } from '@/services/catalog';
import { settingsService } from '@/services/settings';
import { useNotificationStore } from '@/stores/notification';

const ProductSchema = z.object({
  sku: z.string().min(1, 'Requerido').max(50),
  barcode: z.string().optional(),
  description: z.string().min(1, 'Requerido'),
  categoryId: z.string().optional(),
  brandId: z.string().optional(),
  taxId: z.string().min(1, 'Seleccione un impuesto'),
  costUSD: z.number().min(0, 'Debe ser >= 0'),
  salePrice: z.number().min(0, 'Debe ser >= 0'),
  minStock: z.number().min(0, 'Debe ser >= 0'),
  maxStock: z.number().min(0, 'Debe ser >= 0'),
  status: z.enum(['active', 'inactive']),
});

type ProductFormValues = z.infer<typeof ProductSchema>;

const statusOptions = [
  { value: 'active', label: 'Activo' },
  { value: 'inactive', label: 'Inactivo' },
];

function defaultUnitId(units: UnitDTO[]): string {
  const byCode = (code: string) => units.find((u) => u.code.toUpperCase() === code);
  return (
    byCode('NIU')?.id ??
    byCode('UND')?.id ??
    units.find((u) => /unidad/i.test(u.name))?.id ??
    units[0]?.id ??
    ''
  );
}

interface ProductFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  product?: Product | null;
}

export function ProductFormDialog({ open, onOpenChange, product }: ProductFormDialogProps) {
  const create = useCreateProduct();
  const update = useUpdateProduct();
  const push = useNotificationStore((s) => s.push);

  const unitsQuery = useQuery({ queryKey: ['catalog', 'units'], queryFn: () => catalogService.listUnits(), staleTime: 5 * 60 * 1000 });
  const taxesQuery = useQuery({ queryKey: ['catalog', 'taxes'], queryFn: () => settingsService.getTaxes(), staleTime: 5 * 60 * 1000 });
  const categoriesQuery = useCategories();
  const brandsQuery = useBrands();

  const defaults = useMemo<ProductFormValues>(
    () => ({
      sku: product?.sku ?? '',
      barcode: product?.barcode ?? '',
      description: product?.description ?? '',
      categoryId: product?.categoryId ?? '',
      brandId: product?.brandId ?? '',
      taxId: product?.taxId ?? '',
      costUSD: product?.costUSD ?? 0,
      salePrice: product?.salePrice ?? 0,
      minStock: product?.minStock ?? 0,
      maxStock: product?.maxStock ?? 0,
      status: product ? (product.isActive ? 'active' : 'inactive') : 'active',
    }),
    [product],
  );

  const handleSubmit = (values: ProductFormValues) => {
    const onSuccess = () => {
      push({ title: product ? 'Producto actualizado' : 'Producto creado', variant: 'success' });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({ title: 'No se pudo guardar el producto', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    };

    if (product) {
      update.mutate(
        {
          id: product.id,
          description: values.description,
          categoryId: values.categoryId || undefined,
          brandId: values.brandId || undefined,
          costUSD: values.costUSD,
          salePrice: values.salePrice,
          minStock: values.minStock,
          maxStock: values.maxStock,
          isActive: values.status === 'active',
        },
        { onSuccess, onError },
      );
      return;
    }

    create.mutate(
      {
        sku: values.sku,
        barcode: values.barcode || undefined,
        description: values.description,
        categoryId: values.categoryId || undefined,
        brandId: values.brandId || undefined,
        unitId: defaultUnitId(unitsQuery.data ?? []),
        taxId: values.taxId,
        costUSD: values.costUSD,
        salePrice: values.salePrice,
        minStock: values.minStock,
        maxStock: values.maxStock,
      },
      { onSuccess, onError },
    );
  };

  const loading = create.isPending || update.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{product ? 'Editar producto' : 'Nuevo producto'}</DialogTitle>
          <DialogDescription>
            {product ? 'Actualiza la descripción, precios y niveles de stock.' : 'Registra un nuevo producto o servicio.'}
          </DialogDescription>
        </DialogHeader>

        <Form<ProductFormValues> schema={ProductSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="dialog-body-scroll">
                <Grid cols={2}>
                  <TextField name="sku" label="SKU" required description="Código único interno. Letras, números, - _ ." />
                  <TextField name="barcode" label="Código de barras" description="Opcional: código EAN/UPC." />
                </Grid>
                <TextField name="description" label="Descripción" required />
                <Grid cols={2}>
                  <SelectField name="categoryId" label="Categoría" options={categoriesQuery.data ?? []} loading={categoriesQuery.isLoading} />
                  <SelectField name="brandId" label="Marca" options={brandsQuery.data ?? []} loading={brandsQuery.isLoading} />
                </Grid>
                {!product && (
                  <SelectField
                    name="taxId"
                    label="Impuesto"
                    required
                    options={(taxesQuery.data ?? []).map((t) => ({ value: t.id, label: `${t.code} — ${t.name} (${t.defaultRate}%)` }))}
                    loading={taxesQuery.isLoading}
                  />
                )}
                <Grid cols={2}>
                  <MoneyField name="costUSD" label="Costo Base (USD)" currency="USD" description="Costo de adquisición en dólares." />
                  <MoneyField name="salePrice" label="Precio de venta" description="Precio unitario de venta (PEN)." />
                </Grid>
                <Grid cols={2}>
                  <NumberField name="minStock" label="Stock mínimo" min={0} step={0.0001} description="Alerta cuando el stock baja de este nivel." />
                  <NumberField name="maxStock" label="Stock máximo" min={0} step={0.0001} description="Nivel tope de reposición." />
                </Grid>
                {product && (
                  <Grid cols={2}>
                    <SelectField name="status" label="Estado" options={statusOptions} clearable={false} />
                  </Grid>
                )}
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
                  Cancelar
                </Button>
                <Button type="submit" loading={loading} disabled={!formState.isValid}>
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
