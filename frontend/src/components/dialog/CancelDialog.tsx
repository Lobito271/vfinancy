import { useState } from 'react';
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './Dialog';
import { Button } from '@/components/button';
import { Label, Textarea } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';

interface CancelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  confirmLabel?: string;
  reasons?: string[];
  loading?: boolean;
  onConfirm: (reason: string, preset?: string) => void;
}

export function CancelDialog({
  open,
  onOpenChange,
  title = 'Cancelar documento',
  confirmLabel = 'Cancelar documento',
  reasons,
  loading = false,
  onConfirm,
}: CancelDialogProps) {
  const [reason, setReason] = useState('');
  const [preset, setPreset] = useState<string | null>(null);

  const pickPreset = (value: string | null) => {
    setPreset(value);
    if (value) setReason(value);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          {reasons && reasons.length > 0 && (
            <div className="field">
              <Label htmlFor="cancel-reason-preset">Motivo común</Label>
              <Select value={preset ?? ''} onValueChange={pickPreset}>
                <SelectTrigger id="cancel-reason-preset" aria-label="Seleccionar un motivo común">
                  <SelectValue placeholder="Selecciona un motivo…" />
                </SelectTrigger>
                <SelectContent>
                  {reasons.map((r) => (
                    <SelectItem key={r} value={r}>
                      {r}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
          <div className="field">
            <Label htmlFor="cancel-reason">Motivo</Label>
            <Textarea
              id="cancel-reason"
              rows={3}
              placeholder="Motivo de la cancelación…"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
            Volver
          </Button>
          <Button
            variant="destructive"
            onClick={() => onConfirm(reason.trim(), preset ?? undefined)}
            loading={loading}
            disabled={!reason.trim()}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}