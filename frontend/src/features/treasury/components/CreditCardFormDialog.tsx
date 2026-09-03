import { useMemo } from 'react';
import { z } from 'zod';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/dialog';
import { Form, TextField, NumberField, SelectField } from '@/components/form';
import { Button } from '@/components/button';
import { useCreateCreditCard, useUpdateCreditCard } from '../hooks/useTreasury';
import { useNotificationStore } from '@/stores/notification';

const schema = z.object({
  issuer: z.string().min(1, 'Selecciona el emisor.'),
  lastFour: z.string().length(4, 'Deben ser exactamente 4 dígitos.').regex(/^\d{4}$/, 'Solo dígitos.'),
  cardHolder: z.string().min(2, 'Ingresa el nombre del titular.'),
  creditLimit: z.number().positive('Debe ser positivo.'),
  cutOffDay: z.number().int().min(1).max(31),
  paymentDueDay: z.number().int().min(1).max(31),
  currencyCode: z.string().min(1, 'Selecciona la moneda.'),
});

type FormValues = z.infer<typeof schema>;

interface CreditCardFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editCard?: {
    id: string;
    issuer: string;
    lastFour: string;
    cardHolder: string;
    creditLimit: number;
    cutOffDay: number;
    paymentDueDay: number;
    currencyCode: string;
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
      issuer: editCard?.issuer ?? 'visa',
      lastFour: editCard?.lastFour ?? '',
      cardHolder: editCard?.cardHolder ?? '',
      creditLimit: editCard?.creditLimit ?? 1000,
      cutOffDay: editCard?.cutOffDay ?? 25,
      paymentDueDay: editCard?.paymentDueDay ?? 20,
      currencyCode: editCard?.currencyCode ?? 'USD',
    }),
    [editCard],
  );

  async function handleSubmit(values: FormValues) {
    try {
      if (isEditing && editCard) {
        await update.mutateAsync({
          id: editCard.id,
          ...values,
          isActive: editCard.isActive,
        });
        push({ title: 'Tarjeta actualizada', variant: 'success' });
      } else {
        await create.mutateAsync({
          ...values,
          expirationMonth: 12,
          expirationYear: 2030,
        });
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
          <div className="stack">
            <SelectField
              name="issuer"
              label="Emisor"
              options={[
                { value: 'visa', label: 'Visa' },
                { value: 'mastercard', label: 'Mastercard' },
                { value: 'amex', label: 'American Express' },
                { value: 'diners', label: 'Diners Club' },
                { value: 'other', label: 'Otro' },
              ]}
              required
            />
            <TextField name="lastFour" label="Últimos 4 dígitos" placeholder="1234" required />
            <TextField name="cardHolder" label="Nombre del titular" required />
            <NumberField name="creditLimit" label="Límite de crédito (USD)" min={0} step={100} required />
            <NumberField name="cutOffDay" label="Día de corte (1-31)" min={1} max={31} required />
            <NumberField name="paymentDueDay" label="Día de pago (1-31)" min={1} max={31} required />
            <SelectField
              name="currencyCode"
              label="Moneda"
              options={[
                { value: 'USD', label: 'Dólar estadounidense (USD)' },
                { value: 'PEN', label: 'Sol peruano (PEN)' },
              ]}
              required
            />
          </div>
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
