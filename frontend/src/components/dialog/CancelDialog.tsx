import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './Dialog';
import { Button } from '@/components/button';
import { Label, Textarea } from '@/components/input';

interface CancelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  description?: string;
  confirmLabel?: string;
  loading?: boolean;
  onConfirm: (reason: string) => void;
}

export function CancelDialog({
  open,
  onOpenChange,
  title = 'Cancelar documento',
  description = 'Indique el motivo de la cancelación. Esta acción no se puede deshacer.',
  confirmLabel = 'Cancelar documento',
  loading = false,
  onConfirm,
}: CancelDialogProps) {
  const [reason, setReason] = useState('');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="stack stack--sm dialog-body-scroll">
          <Label htmlFor="cancel-reason">Motivo</Label>
          <Textarea
            id="cancel-reason"
            rows={3}
            placeholder="Motivo de la cancelación…"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
            Volver
          </Button>
          <Button variant="destructive" onClick={() => onConfirm(reason.trim())} loading={loading} disabled={!reason.trim()}>
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
