import { z } from 'zod';
import { Form, MoneyField } from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { formatCurrency } from '@/utils/format';

const CardPaymentSchema = z.object({
  amount: z.number().positive('El monto debe ser mayor a 0'),
});

type CardPaymentValues = z.infer<typeof CardPaymentSchema>;

interface CardPaymentDialogProps {
  card: { issuer: string; lastFour: string; currentBalance: number } | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  loading: boolean;
  onConfirm: (amount: number) => void;
}

export function CardPaymentDialog({ card, open, onOpenChange, loading, onConfirm }: CardPaymentDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>Registrar pago</DialogTitle>
          <DialogDescription>
            El pago liquida el ciclo activo de la tarjeta {card?.issuer} •••• {card?.lastFour}.
          </DialogDescription>
        </DialogHeader>
        <div className="doc-summary">
          <div className="doc-summary__row">
            <div className="doc-summary__meta">Saldo actual del ciclo</div>
            <div className="doc-summary__amount">{formatCurrency(card?.currentBalance ?? 0, 'USD')}</div>
          </div>
        </div>
        <Form<CardPaymentValues>
          key={card?.issuer ?? ''}
          schema={CardPaymentSchema}
          defaultValues={{ amount: card?.currentBalance ?? 0 }}
          onSubmit={(values) => onConfirm(values.amount)}
        >
          <DialogBody>
            <MoneyField
              name="amount"
              label="Monto del pago (USD)"
              currency="USD"
              required
              description="Se registra el pago y el saldo del ciclo vuelve a $0.00."
            />
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
              Cancelar
            </Button>
            <Button type="submit" loading={loading}>
              Confirmar pago
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
