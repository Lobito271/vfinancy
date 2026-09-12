import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Boxes, Lock, Plus, Trash2 } from 'lucide-react';
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
import { useNotificationStore } from '@/stores/notification';

interface ImportLotsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ImportLotsDialog({ open, onOpenChange }: ImportLotsDialogProps) {
  const qc = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const { data: purchases = [], isLoading: purchasesLoading } = usePurchases();
  const lotsQuery = useQuery({
    queryKey: ['import-lots'],
    queryFn: () => wailsClient.listImportLots({ page: 1, pageSize: 100 }, ''),
    enabled: open,
  });

  const [activeLotId, setActiveLotId] = useState('');
  const [description, setDescription] = useState('');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [warning, setWarning] = useState<{ lot: ImportLotDTO; purchaseIds: string[] } | null>(null);

  const membersQuery = useQuery({
    queryKey: ['import-lots', 'members', activeLotId],
    queryFn: () => wailsClient.listLotMembers(activeLotId),
    enabled: open && Boolean(activeLotId),
  });

  const invalidate = () => {
    void qc.invalidateQueries({ queryKey: ['import-lots'] });
  };

  const rollback = useMutation({
    mutationFn: (undo: { lotId: string; purchaseIds: string[] }) =>
      Promise.all(undo.purchaseIds.map((id) => wailsClient.removeFromImportLot(undo.lotId, id))),
    onSuccess: () => {
      invalidate();
      setWarning(null);
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo revertir la agrupación', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const group = useMutation({
    mutationFn: () => {
      const purchaseIds = Array.from(selected);
      if (activeLotId) return wailsClient.addToImportLot(activeLotId, purchaseIds);
      return wailsClient.createImportLot(description, purchaseIds);
    },
    onSuccess: (res) => {
      const purchaseIds = Array.from(selected);
      setSelected(new Set());
      setDescription('');
      invalidate();
      void qc.invalidateQueries({ queryKey: ['import-lots', 'members', activeLotId] });
      if (res.overLimit) setWarning({ lot: res, purchaseIds });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo agrupar', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const removeMember = useMutation({
    mutationFn: (purchaseId: string) => wailsClient.removeFromImportLot(activeLotId, purchaseId),
    onSuccess: () => {
      invalidate();
      void qc.invalidateQueries({ queryKey: ['import-lots', 'members', activeLotId] });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo quitar la orden', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const closeLot = useMutation({
    mutationFn: () => wailsClient.closeImportLot(activeLotId),
    onSuccess: () => {
      setActiveLotId('');
      invalidate();
      push({ title: 'Lote cerrado', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo cerrar el lote', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const memberIds = useMemo(
    () => new Set((membersQuery.data ?? []).map((p) => p.id)),
    [membersQuery.data],
  );

  const selectable = useMemo(
    () => purchases.filter((p) => p.status !== 'cancelled' && !selected.has(p.id) && !memberIds.has(p.id)),
    [purchases, selected, memberIds],
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
  const activeLot = lots.find((l) => l.id === activeLotId) ?? null;

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
                          <span className="tabular">{formatCurrency(p.costUsd, 'USD')}</span>
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
                      <button
                        key={l.id}
                        type="button"
                        className="card"
                        data-active={l.id === activeLotId || undefined}
                        style={{ padding: '0.75rem', textAlign: 'left', cursor: 'pointer' }}
                        onClick={() => setActiveLotId(l.id === activeLotId ? '' : l.id)}
                      >
                        <div className="hstack hstack--sm" style={{ justifyContent: 'space-between' }}>
                          <strong>{l.code}</strong>
                          <div className="hstack hstack--sm">
                            {l.status === 'closed' && <Badge variant="muted">Cerrado</Badge>}
                            {l.overLimit ? (
                              <Badge variant="destructive">Supera tope aduanero</Badge>
                            ) : (
                              <Badge variant="muted">Dentro del tope</Badge>
                            )}
                          </div>
                        </div>
                        {l.description && <p className="muted" style={{ fontSize: '0.8rem' }}>{l.description}</p>}
                        <p className="muted" style={{ fontSize: '0.8rem' }}>
                          Total{' '}
                          <span className="tabular">{formatCurrency(l.totalUsd, 'USD')}</span> de{' '}
                          {formatCurrency(l.customsLimitUsd, 'USD')}
                        </p>
                      </button>
                    ))}
                  </div>
                )}

                {activeLot && (
                  <div className="stack">
                    <Label>Órdenes del lote {activeLot.code}</Label>
                    {membersQuery.isLoading ? (
                      <Spinner />
                    ) : (membersQuery.data ?? []).length === 0 ? (
                      <EmptyState title="Sin órdenes" description="Este lote aún no tiene órdenes asignadas." />
                    ) : (
                      <div className="stack">
                        {(membersQuery.data ?? []).map((m) => (
                          <div
                            key={m.id}
                            className="hstack hstack--sm"
                            style={{ justifyContent: 'space-between', padding: '0.4rem 0', borderBottom: '1px solid var(--color-border)' }}
                          >
                            <span style={{ flex: 1 }}>{m.number}</span>
                            <span className="tabular muted">{formatCurrency(m.costUsd, 'USD')}</span>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon-sm"
                              aria-label={`Quitar ${m.number} del lote`}
                              onClick={() => removeMember.mutate(m.id)}
                              disabled={removeMember.isPending}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)} disabled={group.isPending}>
              Cerrar
            </Button>
            {activeLot && activeLot.status === 'active' && (
              <Button variant="outline" onClick={() => closeLot.mutate()} loading={closeLot.isPending}>
                <Lock /> Cerrar lote
              </Button>
            )}
            <Button
              onClick={() => group.mutate()}
              loading={group.isPending}
              disabled={selected.size === 0 || activeLot?.status === 'closed'}
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
          if (!open && warning) {
            rollback.mutate({ lotId: warning.lot.id, purchaseIds: warning.purchaseIds });
            setWarning(null);
          }
        }}
        title="El lote supera el tope aduanero"
        description={
          warning
            ? `El lote supera el tope aduanero de ${formatCurrency(warning.lot.customsLimitUsd, 'USD')}. Continuar requiere su confirmación explícita.`
            : ''
        }
        confirmLabel="Confirmar y agrupar"
        onConfirm={() => setWarning(null)}
      />
    </>
  );
}
