import { useMemo } from 'react';
import { z } from 'zod';
import { Form, DateField, TextareaField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';

const FaultySchema = z.object({
  arrivalDate: z.string().min(1, 'Fecha requerida').regex(/^\d{4}-\d{2}-\d{2}$/, 'Formato de fecha inválido'),
  reason: z.string().min(1, 'Indique el motivo'),
});

type FaultyValues = z.infer<typeof FaultySchema>;

export interface MarkFaultyInput {
  arrivalDate: string;
  reason: string;
}

interface MarkFaultyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  documentNumber: string;
  loading: boolean;
  onConfirm: (input: MarkFaultyInput) => void;
}

function today(): string {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
}

const todayString = today();

export function MarkFaultyDialog({ open, onOpenChange, documentNumber, loading, onConfirm }: MarkFaultyDialogProps) {
  const defaults = useMemo<FaultyValues>(() => ({ arrivalDate: todayString, reason: '' }), []);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>Llegó en mal estado</DialogTitle>
          <DialogDescription>
            Se anulará el pedido <span className="fw-medium">{documentNumber}</span>, se restituirá el inventario y se
            reembolsarán automáticamente todos los anticipos registrados.
          </DialogDescription>
        </DialogHeader>

        <Form key={documentNumber} schema={FaultySchema} defaultValues={defaults} onSubmit={onConfirm}>
          {({ formState }) => (
            <>
              <div className="stack">
                <DateField name="arrivalDate" label="Fecha de llegada" required />
                <TextareaField name="reason" label="Motivo del daño" rows={3} required placeholder="Describa el estado de la mercadería…" />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
                  Volver
                </Button>
                <Button type="submit" variant="destructive" loading={loading} disabled={!formState.isValid}>
                  Marcar como defectuoso
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
