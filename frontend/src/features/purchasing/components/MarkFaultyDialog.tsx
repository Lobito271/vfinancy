import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextareaField } from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';

const FaultySchema = z.object({
  reason: z.string().min(1, 'Indique el motivo'),
});

type FaultyValues = z.infer<typeof FaultySchema>;

export interface MarkFaultyInput {
  reason: string;
}

interface MarkFaultyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  documentNumber: string;
  loading: boolean;
  onConfirm: (input: MarkFaultyInput) => void;
}

export function MarkFaultyDialog({ open, onOpenChange, documentNumber, loading, onConfirm }: MarkFaultyDialogProps) {
  const defaults = useMemo<FaultyValues>(() => ({ reason: '' }), []);

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
              <DialogBody>
                <TextareaField name="reason" label="Motivo del daño" rows={3} required placeholder="Describa el estado de la mercadería…" />
              </DialogBody>
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