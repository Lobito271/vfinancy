import { z } from 'zod';
import { Form, DateField, SelectField, TextareaField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { PaymentMethodOptions } from '@/constants/paymentMethods';
import { formatCurrency } from '@/utils/format';

const RegisterPaymentSchema = z.object({
  paymentDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
  method: z.string().min(1, 'Seleccione un método'),
  reference: z.string().optional(),
  notes: z.string().optional(),
});

type RegisterPaymentValues = z.infer<typeof RegisterPaymentSchema>;

export interface RegisterPaymentInput {
  paymentDate: string;
  method: string;
  reference: string;
  notes: string;
}

interface RegisterPaymentDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  documentNumber: string;
  amount: number;
  amountLabel: string;
  confirmLabel: string;
  loading: boolean;
  onConfirm: (input: RegisterPaymentInput) => void;
}

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

export function RegisterPaymentDialog({
  open,
  onOpenChange,
  title,
  description,
  documentNumber,
  amount,
  amountLabel,
  confirmLabel,
  loading,
  onConfirm,
}: RegisterPaymentDialogProps) {
  const defaults: RegisterPaymentValues = { paymentDate: today(), method: 'cash', reference: '', notes: '' };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        <div className="rounded-lg border bg-muted/40 p-4">
          <div className="flex items-baseline justify-between gap-4">
            <div className="text-sm text-muted-foreground">
              {amountLabel} · <span className="font-medium text-foreground">{documentNumber}</span>
            </div>
            <div className="text-lg font-semibold tabular-nums">{formatCurrency(amount)}</div>
          </div>
        </div>

        <Form
          key={documentNumber}
          schema={RegisterPaymentSchema}
          defaultValues={defaults}
          onSubmit={(values) =>
            onConfirm({
              ...values,
              reference: values.reference ?? '',
              notes: values.notes ?? '',
            })
          }
        >
          {({ formState }) => (
            <>
              <div className="space-y-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <DateField name="paymentDate" label="Fecha de pago" required />
                  <SelectField name="method" label="Método de pago" required placeholder="Seleccione…" options={PaymentMethodOptions} />
                </div>
                <TextareaField name="reference" label="Referencia" rows={1} />
                <TextareaField name="notes" label="Notas" rows={2} />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
                  Cancelar
                </Button>
                <Button type="submit" loading={loading} disabled={!formState.isValid}>
                  {confirmLabel}
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
