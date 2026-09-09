import { useMemo } from 'react';
import { z } from 'zod';
import { DialogBody, Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/dialog';
import { Form, TextField, NumberField } from '@/components/form';
import { Button } from '@/components/button';
import { useCreateCreditCard, useUpdateCreditCard } from '../hooks/useTreasury';
import { useNotificationStore } from '@/stores/notification';

const schema = z.object({
  issuer: z.string().min(1, 'Selecciona el banco o entidad.'),
  lastFour: z.string().length(4, 'Deben ser exactamente 4 dígitos.').regex(/^\d{4}$/, 'Solo dígitos.'),
  creditLimit: z.number().positive('Debe ser positivo.'),
  cutOffDay: z.number().int().min(1).max(31),
  paymentDueDay: z.number().int().min(1).max(31),
});

type FormValues = z.infer<typeof schema>;

interface CreditCardFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editCard?: {
    id: string;
    issuer: string;
    lastFour: string;
    creditLimit: number;
    cutOffDay: number;
    paymentDueDay: number;
    isActive: boolean;
  } | null;
}

export function CreditCardFormDialog({ open, onOpenChange, editCard }: CreditCardFormDialogProps) {
  const create = useCreateCreditCard();
  const update = useUpdateCreditCard();
  const push = useNotificationStore((s) => s.push);

  const isEditing = !!editCard;
  const isPending = create.isPending || update.isPending;

  const defaultValues = useMemo<FormValues>(
    () => ({
      issuer: editCard?.issuer ?? '',
      lastFour: editCard?.lastFour ?? '',
      creditLimit: editCard?.creditLimit ?? 1000,
      cutOffDay: editCard?.cutOffDay ?? 25,
      paymentDueDay: editCard?.paymentDueDay ?? 20,
    }),
    [editCard],
  );

  async function handleSubmit(values: FormValues) {
    try {
      if (isEditing && editCard) {
        await update.mutateAsync({
          id: editCard.id,
          ...values,
        });
        push({ title: 'Tarjeta actualizada', variant: 'success' });
      } else {
        await create.mutateAsync(values);
        push({ title: 'Tarjeta creada', variant: 'success' });
      }
      onOpenChange(false);
    } catch (err: unknown) {
      push({
        title: isEditing ? 'No se pudo actualizar' : 'No se pudo crear',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEditing ? 'Editar tarjeta' : 'Nueva tarjeta de crédito'}</DialogTitle>
          <DialogDescription>
            {isEditing ? 'Actualiza los datos de la tarjeta.' : 'Registra una nueva tarjeta para compras y proyecciones.'}
          </DialogDescription>
        </DialogHeader>
        <Form<FormValues> schema={schema} defaultValues={defaultValues} onSubmit={handleSubmit}>
          <DialogBody>
            <TextField name="issuer" label="Banco / Entidad" placeholder="Banco o entidad" required />
            <TextField name="lastFour" label="Últimos 4 dígitos" placeholder="1234" required />
            <NumberField name="creditLimit" label="Límite de crédito (USD)" min={0} step={100} required />
            <NumberField name="cutOffDay" label="Día de corte (1-31)" min={1} max={31} required />
            <NumberField name="paymentDueDay" label="Día de pago (1-31)" min={1} max={31} required />
          </DialogBody>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
              Cancelar
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? 'Actualizar' : 'Crear'}
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
