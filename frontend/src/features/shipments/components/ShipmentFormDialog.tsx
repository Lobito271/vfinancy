import { useCallback, useMemo } from 'react';
import { z } from 'zod';
import {
  DialogBody,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/dialog';
import {
  Form,
  TextField,
  TextareaField,
  AsyncSelectField,
  CustomerSelectField,
  type CreateSelectOption,
} from '@/components/form';
import { Button } from '@/components/button';
import { salesService } from '@/services/sales';
import { useCreateShipment, useUpdateShipment } from '../hooks/useShipments';
import { SHIPMENT_STATUSES } from '@/services/shipments';
import type { ShipmentDTO } from '@/services/wails-types';
import { useNotificationStore } from '@/stores/notification';

const NEXT_STATUS: Record<ShipmentDTO['status'], ShipmentDTO['status'] | null> = {
  pending: 'shipped',
  shipped: 'delivered',
  delivered: null,
};

function SaleSelectField({
  name, label, required,
}: { name: 'saleId'; label: string; required?: boolean }) {
  const load = useCallback(async () => {
    const sales = await salesService.list({ pageSize: 200 });
    return sales.map((s) => ({ value: s.id, label: `#${s.number} — ${s.customerName}` }));
  }, []);
  return <AsyncSelectField name={name} label={label} required={required} loadOptions={load} />;
}

const schema = z.object({
  customerId: z.string().optional(),
  saleId: z.string().optional(),
  description: z.string().min(1, 'Ingresa una descripción corta del envío.'),
  notes: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

interface ShipmentFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  edit?: ShipmentDTO | null;
  customerCreateOption?: CreateSelectOption;
}

export function ShipmentFormDialog({ open, onOpenChange, edit, customerCreateOption }: ShipmentFormDialogProps) {
  const create = useCreateShipment();
  const update = useUpdateShipment();
  const push = useNotificationStore((s) => s.push);

  const isEditing = !!edit;
  const isPending = create.isPending || update.isPending;

  const defaultValues = useMemo<FormValues>(
    () => ({
      customerId: edit?.customerId || undefined,
      saleId: edit?.saleId || undefined,
      description: edit?.description ?? '',
      notes: edit?.notes ?? '',
    }),
    [edit],
  );

  async function handleSubmit(values: FormValues) {
    try {
      if (isEditing && edit) {
        await update.mutateAsync({ id: edit.id, input: values, status: edit.status });
        push({ title: 'Envío actualizado', variant: 'success' });
      } else {
        await create.mutateAsync(values);
        push({ title: 'Envío creado', variant: 'success' });
      }
      onOpenChange(false);
    } catch (err: unknown) {
      push({
        title: isEditing ? 'No se pudo actualizar el envío' : 'No se pudo crear el envío',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    }
  }

  function advanceStatus() {
    if (!edit) return;
    const next = NEXT_STATUS[edit.status];
    if (!next) return;
    update.mutate(
      { id: edit.id, input: { customerId: edit.customerId, saleId: edit.saleId, description: edit.description, notes: edit.notes }, status: next },
      {
        onSuccess: () => push({ title: `Envío marcado como ${SHIPMENT_STATUSES.find((s) => s.value === next)?.label.toLowerCase()}`, variant: 'success' }),
        onError: (err: unknown) =>
          push({ title: 'No se pudo actualizar el estado', description: err instanceof Error ? err.message : undefined, variant: 'destructive' }),
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEditing ? `Editar envío #${edit.code}` : 'Nuevo envío'}</DialogTitle>
          <DialogDescription>
            {isEditing
              ? 'El código de envío no se puede cambiar.'
              : 'Se generará un código de 4 dígitos para compartir con el transportista.'}
          </DialogDescription>
        </DialogHeader>
        <Form<FormValues> schema={schema} defaultValues={defaultValues} onSubmit={handleSubmit}>
          <DialogBody>
            <CustomerSelectField
              name="customerId"
              label="Cliente"
              placeholder="Opcional"
              createOption={customerCreateOption}
            />
            <SaleSelectField name="saleId" label="Venta" />
            {isEditing && edit && (
              <div className="hstack" style={{ gap: '1rem' }}>
                <span className="muted" style={{ flex: 1 }}>
                  Estado: {SHIPMENT_STATUSES.find((s) => s.value === edit.status)?.label}
                </span>
                {NEXT_STATUS[edit.status] && (
                  <Button type="button" variant="outline" size="sm" onClick={advanceStatus} disabled={isPending}>
                    Marcar como {SHIPMENT_STATUSES.find((s) => s.value === NEXT_STATUS[edit.status])?.label}
                  </Button>
                )}
              </div>
            )}
            <TextField name="description" label="Descripción" placeholder="Ej. Caja de mercadería" required />
            <TextareaField name="notes" label="Notas" placeholder="Instrucciones de entrega" />
          </DialogBody>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              Cancelar
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? 'Guardar' : 'Crear envío'}
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}