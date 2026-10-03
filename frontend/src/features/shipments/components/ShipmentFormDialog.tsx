import { useCallback, useMemo } from 'react';
import { useWatch } from 'react-hook-form';
import { z } from 'zod';
import { Copy } from 'lucide-react';
import {
  DialogBody,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/dialog';
import {
  Form,
  TextField,
  TextareaField,
  AsyncSelectField,
  DateField,
} from '@/components/form';
import { Button } from '@/components/button';
import { salesService } from '@/services/sales';
import { useCreateShipment, useUpdateShipment } from '../hooks/useShipments';
import { SHIPMENT_STATUSES } from '@/services/shipments';
import type { ShipmentDTO } from '@/services/wails-types';
import { useNotificationStore } from '@/stores/notification';
import { copyText } from '@/utils/clipboard';

const NEXT_STATUS: Record<ShipmentDTO['status'], ShipmentDTO['status'] | null> = {
  pending: 'shipped',
  shipped: 'delivered',
  delivered: null,
};

/** Four digits the recipient reads out on delivery. */
function randomSecurityCode(): string {
  return String(Math.floor(Math.random() * 9000) + 1000);
}

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

function SaleSelectField({
  name,
  label,
  required,
  disabled,
}: {
  name: 'saleId';
  label: string;
  required?: boolean;
  disabled?: boolean;
}) {
  const load = useCallback(async () => {
    const sales = await salesService.list({ pageSize: 200 });
    return sales.map((s) => ({ value: s.id, label: `#${s.number} — ${s.customerName}` }));
  }, []);
  return (
    <AsyncSelectField name={name} label={label} required={required} disabled={disabled} loadOptions={load} />
  );
}

const schema = z
  .object({
    saleId: z.string().min(1, 'Selecciona la venta que originó el envío'),
    securityCode: z.string().regex(/^\d{4}$/, 'La clave debe tener 4 dígitos'),
    location: z.string().optional(),
    shipmentDate: z.string().optional(),
    deliveredAt: z.string().optional(),
    description: z.string().min(1, 'Ingresa una descripción corta del envío'),
    notes: z.string().optional(),
  })
  .refine((v) => !v.shipmentDate || !v.deliveredAt || v.deliveredAt >= v.shipmentDate, {
    message: 'La entrega no puede ser anterior al envío',
    path: ['deliveredAt'],
  });

type FormValues = z.infer<typeof schema>;

interface ShipmentFieldsProps {
  isEditing: boolean;
  onCopySecurityCode: (code: string) => void;
}

/** Renders inside <Form>: it needs the form context to watch values. */
function ShipmentFields({ isEditing, onCopySecurityCode }: ShipmentFieldsProps) {
  const shipmentDate = useWatch<FormValues>({ name: 'shipmentDate' }) ?? '';
  const securityCode = useWatch<FormValues>({ name: 'securityCode' }) ?? '';
  return (
    <>
      <SaleSelectField name="saleId" label="Venta" required disabled={isEditing} />
      <div className="field">
        <TextField
          name="securityCode"
          label="Clave de seguridad"
          description={
            isEditing
              ? 'Inmutable: ya fue emitida al cliente.'
              : 'El cliente la muestra al recibir. Puedes cambiarla antes de guardar.'
          }
          inputMode="numeric"
          maxLength={4}
          disabled={isEditing}
          required
        />
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={isEditing || !securityCode}
          onClick={() => onCopySecurityCode(securityCode)}
        >
          <Copy size={14} /> Copiar clave
        </Button>
      </div>
      <TextField name="location" label="Ubicación de envío" placeholder="Ej. Oficina de Lima" />
      <div className="grid-2">
        <DateField name="shipmentDate" label="Fecha de envío" />
        <DateField name="deliveredAt" label="Fecha de entrega" min={shipmentDate || undefined} />
      </div>
    </>
  );
}

export interface ShipmentDefaults {
  saleId?: string;
}

interface ShipmentFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  edit?: ShipmentDTO | null;
  defaults?: ShipmentDefaults;
}

export function ShipmentFormDialog({ open, onOpenChange, edit, defaults }: ShipmentFormDialogProps) {
  const create = useCreateShipment();
  const update = useUpdateShipment();
  const push = useNotificationStore((s) => s.push);

  const isEditing = !!edit;
  const isPending = create.isPending || update.isPending;

  const defaultValues = useMemo<FormValues>(
    () => ({
      saleId: edit?.saleId || defaults?.saleId || '',
      securityCode: edit?.securityCode || randomSecurityCode(),
      location: edit?.location ?? '',
      shipmentDate: edit?.shipmentDate ?? '',
      deliveredAt: edit?.deliveredAt ?? '',
      description: edit?.description ?? '',
      notes: edit?.notes ?? '',
    }),
    [edit, defaults],
  );

  function copySecurityCode(code: string) {
    void copyText(code).then(() =>
      push({ title: `Clave ${code} copiada`, variant: 'success' }),
    );
  }

  async function handleSubmit(values: FormValues) {
    try {
      if (isEditing && edit) {
        await update.mutateAsync({ id: edit.id, input: values, status: edit.status });
        push({ title: 'Envío actualizado', variant: 'success' });
      } else {
        await create.mutateAsync(values);
        push({
          title: 'Envío creado',
          description: `Clave de seguridad ${values.securityCode}`,
          variant: 'success',
        });
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

  /** Marks the shipment one step further and stamps the matching date. */
  function advanceStatus() {
    if (!edit) return;
    const next = NEXT_STATUS[edit.status];
    if (!next) return;
    update.mutate(
      {
        id: edit.id,
        input: {
          saleId: edit.saleId,
          location: edit.location,
          shipmentDate: edit.shipmentDate || (next === 'shipped' ? today() : undefined),
          deliveredAt: edit.deliveredAt || (next === 'delivered' ? today() : undefined),
          description: edit.description,
          notes: edit.notes,
        },
        status: next,
      },
      {
        onSuccess: () =>
          push({
            title: `Envío marcado como ${SHIPMENT_STATUSES.find((s) => s.value === next)?.label.toLowerCase()}`,
            variant: 'success',
          }),
        onError: (err: unknown) =>
          push({
            title: 'No se pudo actualizar el estado',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          }),
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEditing ? `Editar envío ${edit.code}` : 'Nuevo envío'}</DialogTitle>
        </DialogHeader>
        <Form<FormValues> schema={schema} defaultValues={defaultValues} onSubmit={handleSubmit}>
          <DialogBody>
            <ShipmentFields isEditing={isEditing} onCopySecurityCode={copySecurityCode} />
            <TextField name="description" label="Descripción" placeholder="Ej. Caja de mercadería" required />
            <TextareaField name="notes" label="Notas" placeholder="Instrucciones de entrega" />
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