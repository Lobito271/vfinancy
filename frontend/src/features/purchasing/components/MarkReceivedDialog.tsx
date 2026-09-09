import { useMemo } from 'react';
import { z } from 'zod';
import { Form, DateField } from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';

const ReceivedSchema = z.object({
  arrivalDate: z
    .string()
    .min(1, 'Fecha requerida')
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido')
    .refine((value) => value <= today(), 'La fecha de recepción no puede ser posterior a hoy'),
});

type ReceivedValues = z.infer<typeof ReceivedSchema>;

interface MarkReceivedInput {
  arrivalDate: string;
}

interface MarkReceivedDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  documentNumber: string;
  loading: boolean;
  onConfirm: (input: MarkReceivedInput) => void;
}

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

const todayString = today();

export function MarkReceivedDialog({ open, onOpenChange, documentNumber, loading, onConfirm }: MarkReceivedDialogProps) {
  const defaults = useMemo<ReceivedValues>(() => ({ arrivalDate: todayString }), []);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>Marcar como Recibido</DialogTitle>
          <DialogDescription>
            Confirma la llegada del pedido <span className="fw-medium">{documentNumber}</span>. La mercadería se
            ingresará al inventario y comenzará a contar el plazo de liquidación (25 días).
          </DialogDescription>
        </DialogHeader>

        <Form key={documentNumber} schema={ReceivedSchema} defaultValues={defaults} onSubmit={onConfirm}>
          {({ formState }) => (
            <>
              <DialogBody>
                <DateField name="arrivalDate" label="Fecha de llegada" required />
              </DialogBody>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
                  Volver
                </Button>
                <Button type="submit" loading={loading} disabled={!formState.isValid}>
                  Confirmar recepción
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
