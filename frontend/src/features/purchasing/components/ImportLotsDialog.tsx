import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Boxes, Plus } from 'lucide-react';
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/dialog';
import { AlertDialog } from '@/components/dialog';
import { Button } from '@/components/button';
import { Badge } from '@/components/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { Input, Label } from '@/components/input';
import { Spinner, EmptyState } from '@/components/feedback';
import { wailsClient } from '@/services/bindings';
import { usePurchases } from '@/features/purchasing/hooks/usePurchases';
import type { ImportLotDTO } from '@/services/wails-types';
import { formatCurrency } from '@/utils/format';

interface ImportLotsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ImportLotsDialog({ open, onOpenChange }: ImportLotsDialogProps) {
  const qc = useQueryClient();
  const { data: purchases = [], isLoading: purchasesLoading } = usePurchases();
  const lotsQuery = useQuery({
    queryKey: ['import-lots'],
    queryFn: () => wailsClient.listImportLots({ status: '', search: '', page: 1, pageSize: 100 }),
    enabled: open,
  });

  const [activeLotId, setActiveLotId] = useState('');
  const [description, setDescription] = useState('');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [warning, setWarning] = useState<ImportLotDTO | null>(null);

  const invalidate = () => void qc.invalidateQueries({ queryKey: ['import-lots'] });

  const group = useMutation({
    mutationFn: () => {
      const purchaseIds = Array.from(selected);
      if (activeLotId) return wailsClient.addToImportLot({ id: activeLotId, purchaseIds });
      return wailsClient.createImportLot({ description, purchaseIds });
    },
    onSuccess: (res) => {
      setSelected(new Set());
      setDescription('');
      invalidate();
      if (res.overLimit) setWarning(res);
    },
  });

  const selectable = useMemo(
    () => purchases.filter((p) => p.status !== 'cancelled' && !selected.has(p.id)),
    [purchases, selected],
  );

  const toggle = (id: string, checked: boolean) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (checked) next.add(id);
      else next.delete(id);
      return next;
    });
  };

  const lots = lotsQuery.data?.items ?? [];

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent size="xl">
          <DialogHeader>
            <DialogTitle>
              <Boxes /> Lotes de importación
            </DialogTitle>
            <DialogDescription>
              Agrupa órdenes de compra para monitorear el tope aduanero simplificado.
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <div className="grid-2" style={{ alignItems: 'start' }}>
              <div className="stack">
                <div className="field">
                  <Label htmlFor="lot-select">Lote</Label>
                  <Select
                    items={[
                      { value: '', label: 'Crear un lote nuevo' },
                      ...lots.map((l) => ({ value: l.id, label: `${l.code} · ${l.description || 'sin descripción'}` })),
                    ]}
                    value={activeLotId}
                    onValueChange={(v) => setActiveLotId(v ?? '')}
                  >
                    <SelectTrigger id="lot-select" aria-label="Seleccionar lote">
                      <SelectValue placeholder="Elige un lote" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">Crear un lote nuevo</SelectItem>
                      {lots.map((l) => (
                        <SelectItem key={l.id} value={l.id}>
                          {l.code} · {l.description || 'sin descripción'}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                {!activeLotId && (
                  <div className="field">
                    <Label htmlFor="lot-desc">Descripción del lote</Label>
                    <Input id="lot-desc" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Ej. Envío 3 / proveedor X" />
                  </div>
                )}
                <div className="field">
                  <Label>Órdenes a agrupar (USD)</Label>
                  {purchasesLoading ? (
                    <Spinner />
                  ) : selectable.length === 0 ? (
                    <EmptyState title="Sin órdenes" description="No hay órdenes de compra disponibles para agrupar." />
                  ) : (
                    <div className="stack" style={{ maxHeight: '16rem', overflow: 'auto' }}>
                      {selectable.map((p) => (
                        <label
                          key={p.id}
                          className="hstack hstack--sm"
                          style={{ padding: '0.4rem 0', borderBottom: '1px solid var(--color-border)' }}
                        >
                          <input
                            type="checkbox"
                            checked={selected.has(p.id)}
                            onChange={(e) => toggle(p.id, e.target.checked)}
                          />
                          <span style={{ flex: 1 }}>{p.number}</span>
                          <span className="tabular">{formatCurrency(p.costUSD, 'USD')}</span>
                        </label>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              <div className="stack">
                <Label>Lotes existentes</Label>
                {lotsQuery.isLoading ? (
                  <Spinner />
                ) : lots.length === 0 ? (
                  <EmptyState title="Sin lotes" description="Crea el primer lote de importación." />
                ) : (
                  <div className="stack">
                    {lots.map((l) => (
                      <div key={l.id} className="card" style={{ padding: '0.75rem' }}>
                        <div className="hstack hstack--sm" style={{ justifyContent: 'space-between' }}>
                          <strong>{l.code}</strong>
                          {l.overLimit ? (
                            <Badge variant="destructive">Supera tope aduanero</Badge>
                          ) : (
                            <Badge variant="muted">Dentro del tope</Badge>
                          )}
                        </div>
                        {l.description && <p className="muted" style={{ fontSize: '0.8rem' }}>{l.description}</p>}
                        <p className="muted" style={{ fontSize: '0.8rem' }}>
                          {l.memberCount} pedido(s) · Total{' '}
                          <span className="tabular">{formatCurrency(Number(l.totalUSD), 'USD')}</span> de{' '}
                          {formatCurrency(Number(l.customsLimitUSD), 'USD')}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)} disabled={group.isPending}>
              Cerrar
            </Button>
            <Button
              onClick={() => group.mutate()}
              loading={group.isPending}
              disabled={selected.size === 0}
            >
              <Plus /> Agrupar
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        variant="destructive"
        open={!!warning}
        onOpenChange={(open) => {
          if (!open) setWarning(null);
        }}
        title="El lote supera el tope aduanero"
        description={
          warning
            ? `El total del lote ${warning.code} es ${formatCurrency(Number(warning.totalUSD), 'USD')}, superando el límite de ${formatCurrency(Number(warning.customsLimitUSD), 'USD')}. La agrupación ya se aplicó; revisa el lote por si necesitas dividirlo.`
            : ''
        }
        confirmLabel="Entendido"
        onConfirm={() => setWarning(null)}
      />
    </>
  );
}
