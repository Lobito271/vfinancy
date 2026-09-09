import { useMemo } from 'react';
import { z } from 'zod';
import { Form, NumberField, TextField } from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useAdjustStock } from '@/features/inventory/hooks/useInventory';
import { useNotificationStore } from '@/stores/notification';
import type { InventoryItem } from '@/types/domain';

const AdjustSchema = z.object({
  newQuantity: z.number().positive('La existencia debe ser mayor a 0'),
  notes: z.string().min(1, 'Motivo requerido').max(200),
});

type AdjustFormValues = z.infer<typeof AdjustSchema>;

interface InventoryAdjustDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  batch?: InventoryItem | null;
}

export function InventoryAdjustDialog({ open, onOpenChange, batch }: InventoryAdjustDialogProps) {
  const adjust = useAdjustStock();
  const push = useNotificationStore((s) => s.push);

  const defaults = useMemo<AdjustFormValues>(() => ({ newQuantity: batch?.quantity ?? 0, notes: '' }), [batch]);

  const handleSubmit = (values: AdjustFormValues) => {
    if (!batch) return;
    adjust.mutate(
      { batchId: batch.id, newQuantity: values.newQuantity, notes: values.notes },
      {
        onSuccess: () => {
          push({ title: 'Ajuste aplicado', variant: 'success' });
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo aplicar el ajuste',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          });
        },
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>Ajustar lote</DialogTitle>
          <DialogDescription>
            {batch
              ? `${batch.productSku} — ${batch.productDescription} · Existencia actual: ${batch.quantity}`
              : 'Corrige la existencia de un lote.'}
          </DialogDescription>
        </DialogHeader>

        <Form schema={AdjustSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <DialogBody>
                <NumberField name="newQuantity" label="Nueva existencia" description="Cantidad total que quedará en el lote." required step={0.01} />
                <TextField name="notes" label="Motivo" required maxLength={200} />
              </DialogBody>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={adjust.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={adjust.isPending} disabled={!formState.isValid}>
                  Aplicar
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
