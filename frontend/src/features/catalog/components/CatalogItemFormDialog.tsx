import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import {
  useCreateBrand,
  useCreateCategory,
  useUpdateBrand,
  useUpdateCategory,
} from '@/features/catalog/hooks/useCatalog';
import { useNotificationStore } from '@/stores/notification';

export type CatalogKind = 'category' | 'brand';

export interface CatalogItem {
  id: string;
  code: string;
  name: string;
}

const codeSchema = z
  .string()
  .min(1, 'Requerido')
  .max(20, 'Máximo 20 caracteres')
  .regex(/^[A-Za-z0-9._-]+$/, 'Solo letras, números, -, _ y .')
  .transform((v) => v.toUpperCase());

const CatalogItemSchema = z.object({
  code: codeSchema,
  name: z.string().min(1, 'Requerido').max(200, 'Máximo 200 caracteres'),
});

type CatalogItemValues = z.infer<typeof CatalogItemSchema>;

interface CatalogItemFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  kind: CatalogKind;
  item?: CatalogItem | null;
}

export function CatalogItemFormDialog({ open, onOpenChange, kind, item }: CatalogItemFormDialogProps) {
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory();
  const createBrand = useCreateBrand();
  const updateBrand = useUpdateBrand();
  const push = useNotificationStore((s) => s.push);

  const isCategory = kind === 'category';
  const loading = isCategory
    ? createCategory.isPending || updateCategory.isPending
    : createBrand.isPending || updateBrand.isPending;

  const defaults = useMemo<CatalogItemValues>(
    () => ({ code: item?.code ?? '', name: item?.name ?? '' }),
    [item],
  );

  const handleSubmit = (values: CatalogItemValues) => {
    const onSuccess = () => {
      push({
        title: item
          ? `${isCategory ? 'Categoría' : 'Marca'} actualizada`
          : `${isCategory ? 'Categoría' : 'Marca'} creada`,
        variant: 'success',
      });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({
        title: `No se pudo guardar ${isCategory ? 'la categoría' : 'la marca'}`,
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    };

    if (isCategory) {
      if (item) {
        updateCategory.mutate({ id: item.id, code: values.code, name: values.name }, { onSuccess, onError });
      } else {
        createCategory.mutate({ code: values.code, name: values.name }, { onSuccess, onError });
      }
      return;
    }
    if (item) {
      updateBrand.mutate({ id: item.id, code: values.code, name: values.name }, { onSuccess, onError });
    } else {
      createBrand.mutate({ code: values.code, name: values.name }, { onSuccess, onError });
    }
  };

  const title = item
    ? `Editar ${isCategory ? 'categoría' : 'marca'}`
    : `Nueva ${isCategory ? 'categoría' : 'marca'}`;
  const description = item
    ? `Actualiza los datos de la ${isCategory ? 'categoría' : 'marca'}.`
    : `Registra una nueva ${isCategory ? 'categoría' : 'marca'} para el catálogo de productos.`;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        <Form<CatalogItemValues> schema={CatalogItemSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="stack dialog-body-scroll">
                <Grid cols={2}>
                  <TextField name="code" label="Código" description="Ej.: ABR" required />
                  <TextField name="name" label="Nombre" description="Ej.: Abarrotes" required />
                </Grid>
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
