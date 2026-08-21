import { useMemo } from 'react';
import { z } from 'zod';
import {
  Form,
  ProductSelectField,
  WarehouseSelectField,
  DateField,
  TextField,
  NumberField,
  MoneyField,
} from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useReceiveStock } from '@/features/inventory/hooks/useInventory';
import { useNotificationStore } from '@/stores/notification';

const ReceiveSchema = z.object({
  productId: z.string().min(1, 'Seleccione un producto'),
  warehouseId: z.string().min(1, 'Seleccione un almacén'),
  lotNumber: z.string().min(1, 'Número de lote requerido').max(50),
  arrivalDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
  quantity: z.number().positive('Cantidad debe ser mayor a 0'),
  unitCost: z.number().min(0, 'Debe ser >= 0'),
});

type ReceiveFormValues = z.infer<typeof ReceiveSchema>;

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

interface InventoryReceiveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function InventoryReceiveDialog({ open, onOpenChange }: InventoryReceiveDialogProps) {
  const receive = useReceiveStock();
  const push = useNotificationStore((s) => s.push);

  const defaults = useMemo<ReceiveFormValues>(
    () => ({
      productId: '',
      warehouseId: '',
      lotNumber: '',
      arrivalDate: today(),
      quantity: 1,
      unitCost: 0,
    }),
    [],
  );

  const handleSubmit = (values: ReceiveFormValues) => {
    receive.mutate(values, {
      onSuccess: () => {
        push({ title: 'Stock registrado', variant: 'success' });
        onOpenChange(false);
      },
      onError: (err: unknown) => {
        push({
          title: 'No se pudo registrar el ingreso',
          description: err instanceof Error ? err.message : undefined,
          variant: 'destructive',
        });
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>Ingreso de stock</DialogTitle>
          <DialogDescription>Registra un nuevo lote de mercadería en el almacén.</DialogDescription>
        </DialogHeader>

        <Form schema={ReceiveSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="dialog-body-scroll">
                <ProductSelectField name="productId" label="Producto" required />
                <WarehouseSelectField name="warehouseId" label="Almacén" required />
                <div className="form-grid">
                  <TextField name="lotNumber" label="Número de lote" required />
                  <DateField name="arrivalDate" label="Fecha de ingreso" required />
                </div>
                <div className="form-grid">
                  <NumberField name="quantity" label="Cantidad" required min={0} step={0.01} />
                  <MoneyField name="unitCost" label="Costo unitario" />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={receive.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={receive.isPending} disabled={!formState.isValid}>
                  Registrar
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
