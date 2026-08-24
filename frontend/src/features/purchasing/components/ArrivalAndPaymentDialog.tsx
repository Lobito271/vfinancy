import { z } from 'zod';
import { Form, DateField, SelectField, TextField, TextareaField } from '@/components/form';
import type { SelectOption } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { formatCurrency } from '@/utils/format';

const ArrivalPaymentSchema = z.object({
  arrivalDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
  amount: z.coerce.number().min(0.01, 'El monto debe ser mayor a 0'),
  creditCardId: z.string().min(1, 'Seleccione una tarjeta'),
  reference: z.string().optional(),
  notes: z.string().optional(),
});

type ArrivalPaymentValues = z.infer<typeof ArrivalPaymentSchema>;

export interface ArrivalAndPaymentInput {
  arrivalDate: string;
  amount: number;
  creditCardId: string;
  reference: string;
  notes: string;
}

interface ArrivalAndPaymentDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  documentNumber: string;
  outstandingAmount: number;
  loading: boolean;
  creditCardOptions: SelectOption[];
  creditCardLoading: boolean;
  onConfirm: (input: ArrivalAndPaymentInput) => void;
}

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

export function ArrivalAndPaymentDialog({
  open,
  onOpenChange,
  documentNumber,
  outstandingAmount,
  loading,
  creditCardOptions,
  creditCardLoading,
  onConfirm,
}: ArrivalAndPaymentDialogProps) {
  const defaults: ArrivalPaymentValues = {
    arrivalDate: today(),
    amount: outstandingAmount,
    creditCardId: creditCardOptions[0]?.value ?? '',
    reference: '',
    notes: '',
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>Llegada y Cobro</DialogTitle>
          <DialogDescription>
            Registra la llegada del pedido{' '}
            <span className="font-medium">{documentNumber}</span> y cobra el saldo
            pendiente en un solo paso.
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-lg border bg-muted/40 p-4">
          <div className="flex items-baseline justify-between gap-4">
            <div className="text-sm text-muted-foreground">
              Saldo por cobrar ·{' '}
              <span className="font-medium text-foreground">{documentNumber}</span>
            </div>
            <div className="text-lg font-semibold tabular-nums">
              {formatCurrency(outstandingAmount)}
            </div>
          </div>
        </div>

        <Form
          key={documentNumber}
          schema={ArrivalPaymentSchema}
          defaultValues={defaults}
          onSubmit={(values) =>
            onConfirm({
              arrivalDate: values.arrivalDate,
              amount: values.amount,
              creditCardId: values.creditCardId,
              reference: values.reference ?? '',
              notes: values.notes ?? '',
            })
          }
        >
          {({ formState }) => (
            <>
              <div className="space-y-4">
                <DateField name="arrivalDate" label="Fecha de llegada" required />
                <TextField name="amount" label="Monto a cobrar (PEN)" required />
                <SelectField
                  name="creditCardId"
                  label="Tarjeta de crédito"
                  required
                  placeholder={
                    creditCardLoading ? 'Cargando tarjetas...' : 'Seleccione la tarjeta...'
                  }
                  options={creditCardOptions}
                  loading={creditCardLoading}
                />
                <TextareaField name="reference" label="Referencia" rows={1} />
                <TextareaField name="notes" label="Notas" rows={2} />
              </div>
              <DialogFooter>
                <Button
                  variant="outline"
                  type="button"
                  onClick={() => onOpenChange(false)}
                  disabled={loading}
                >
                  Cancelar
                </Button>
                <Button type="submit" loading={loading} disabled={!formState.isValid}>
                  Confirmar llegada y cobro
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
