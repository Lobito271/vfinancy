import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField, NumberField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { useAdjustStock } from '@/features/inventory/hooks/useInventory';
import { useNotificationStore } from '@/stores/notification';
import type { InventoryItem } from '@/types/domain';

const AdjustSchema = z.object({
  delta: z.number().refine((v) => v !== 0, 'El ajuste no puede ser 0'),
  reason: z.string().min(1, 'Motivo requerido').max(200),
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

  const defaults = useMemo<AdjustFormValues>(() => ({ delta: 0, reason: '' }), []);

  const handleSubmit = (values: AdjustFormValues) => {
    if (!batch) return;
    adjust.mutate(
      { batchId: batch.id, delta: values.delta, reason: values.reason },
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
              <div className="stack stack--lg">
                <NumberField name="delta" label="Cantidad de ajuste" description="Usa valores negativos para reducir stock." required step={0.01} />
                <TextField name="reason" label="Motivo" required maxLength={200} />
              </div>
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
