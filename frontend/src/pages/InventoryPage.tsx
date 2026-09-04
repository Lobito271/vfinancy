import { useMemo, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Boxes, Pencil, Ban, Plus, Settings2, AlertTriangle } from 'lucide-react';
import { z } from 'zod';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { ConfirmDialog } from '@/components/dialog';
import { Drawer, RowActions } from '@/components/misc';
import { Form, NumberField } from '@/components/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useInventory, useVoidStock } from '@/features/inventory/hooks/useInventory';
import { InventoryReceiveDialog } from '@/features/inventory/components/InventoryReceiveDialog';
import { InventoryAdjustDialog } from '@/features/inventory/components/InventoryAdjustDialog';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import type { InventoryItem } from '@/types/domain';
import { formatCurrency, formatDate, formatNumber } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<InventoryItem>[] = [
  {
    id: 'productSku',
    header: 'SKU',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="fw-medium tabular">{row.productSku}</span>,
  },
  {
    id: 'productDescription',
    header: 'Producto',
    sortable: true,
    cell: (row) => row.productDescription,
  },
  { id: 'warehouse', header: 'Almacén', cell: (row) => row.warehouse || '—' },
  {
    id: 'quantity',
    header: 'Cantidad',
    align: 'numeric',
    sortable: true,
    cell: (row) => <span className="tabular">{formatNumber(row.quantity)}</span>,
  },
  {
    id: 'unitCost',
    header: 'Costo unitario',
    align: 'numeric',
    sortable: true,
    cell: (row) => <span className="tabular">{formatCurrency(row.unitCost, row.currencyCode)}</span>,
  },
  {
    id: 'totalCost',
    header: 'Costo total',
    align: 'numeric',
    sortable: true,
    accessor: (row) => row.quantity * row.unitCost,
    cell: (row) => <span className="tabular">{formatCurrency(row.quantity * row.unitCost, row.currencyCode)}</span>,
  },
  {
    id: 'arrivalDate',
    header: 'Ingreso',
    cell: (row) => <span className="muted">{row.arrivalDate ? formatDate(row.arrivalDate) : '—'}</span>,
  },
  {
    id: 'maxSaleDate',
    header: 'Venta máxima',
    cell: (row) => <span className="muted">{row.maxSaleDate ? formatDate(row.maxSaleDate) : '—'}</span>,
  },
  {
    id: 'daysRemaining',
    header: 'Días restantes',
    align: 'numeric',
    sortable: true,
    cell: (row) => <span className="tabular">{row.daysRemaining}</span>,
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => {
      if (row.status === 'voided') return <Badge variant="destructive">Anulado</Badge>;
      if (row.status === 'written_off') return <Badge variant="muted">Baja</Badge>;
      if (row.status === 'depleted') return <Badge variant="muted">Agotado</Badge>;
      if (row.isClearance) return <Badge variant="destructive" className="fw-bold">REMATE</Badge>;
      return <Badge variant="success">Normal</Badge>;
    },
  },
];

const clearanceSchema = z.object({
  clearanceDays: z.number().int().min(1, 'Usa al menos 1 día.').max(365),
  clearanceWarningDays: z.number().int().min(0).max(90),
});

type ClearanceValues = z.infer<typeof clearanceSchema>;

function InventorySettingsDrawer({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);

  const preferences = useQuery({
    queryKey: queryKeys.settings.preferences,
    queryFn: () => wailsClient.getPreferences(),
    enabled: open,
  });

  const [saving, setSaving] = useState(false);

  const defaults: ClearanceValues = {
    clearanceDays: preferences.data?.clearanceDays ?? 25,
    clearanceWarningDays: preferences.data?.clearanceWarningDays ?? 3,
  };

  async function save(values: ClearanceValues) {
    setSaving(true);
    try {
      await wailsClient.updatePreference('clearance_days', String(values.clearanceDays));
      await wailsClient.updatePreference('clearance_warning_days', String(values.clearanceWarningDays));
      await queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences });
      push({ title: 'Reglas de remate guardadas', variant: 'success' });
      onOpenChange(false);
    } catch (cause) {
      push({
        title: 'No se pudo guardar la configuración',
        description: cause instanceof Error ? cause.message : undefined,
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  }

  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      title="Reglas de inventario"
      description="Controla cuándo un lote pasa a remate y con cuánta anticipación se avisa."
      footer={
        <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={saving}>
          Cancelar
        </Button>
      }
    >
      <Form<ClearanceValues> key={`${defaults.clearanceDays}-${defaults.clearanceWarningDays}`} schema={clearanceSchema} defaultValues={defaults} onSubmit={save}>
        {() => (
          <div className="stack">
            <NumberField
              name="clearanceDays"
              label="Días para remate"
              description="Un lote pasa a remate cuando supera estos días desde su ingreso."
              min={1}
              max={365}
              required
            />
            <NumberField
              name="clearanceWarningDays"
              label="Días de aviso previo"
              description="Cuántos días antes del remate se marca el lote como próximo a vencer."
              min={0}
              max={90}
              required
            />
            <Button type="submit" loading={saving}>
              Guardar
            </Button>
          </div>
        )}
      </Form>
    </Drawer>
  );
}

export function InventoryPage() {
  const { data, isLoading, isError, error, refetch } = useInventory();
  const voidStock = useVoidStock();
  const push = useNotificationStore((s) => s.push);

  const [receiveOpen, setReceiveOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [statusFilter, setStatusFilter] = useState('all');
  const [showClearanceOnly, setShowClearanceOnly] = useState(false);
  const [adjustTarget, setAdjustTarget] = useState<InventoryItem | null>(null);
  const [voidTarget, setVoidTarget] = useState<InventoryItem | null>(null);

  const items = data ?? [];
  const live = items.filter((i) => i.status !== 'voided');
  const totalUnits = live.reduce((s, i) => s + i.quantity, 0);
  const inventoryValue = live.reduce((s, i) => s + i.quantity * i.unitCost, 0);
  const clearance = items.filter((i) => i.isClearance).length;
  const expiringSoon = live.filter((i) => i.daysRemaining >= 0 && i.daysRemaining < 5).length;

  const filteredItems = useMemo(() => {
    let result = items;
    if (showClearanceOnly) {
      result = result.filter((i) => i.isClearance && i.status !== 'voided');
    } else {
      if (statusFilter === 'clearance') result = result.filter((i) => i.isClearance && i.status !== 'voided');
      else if (statusFilter === 'expiring') result = live.filter((i) => i.daysRemaining >= 0 && i.daysRemaining < 5);
      else if (statusFilter === 'voided') result = result.filter((i) => i.status === 'voided');
    }
    return result;
  }, [items, live, statusFilter, showClearanceOnly]);

  const openCreate = () => setReceiveOpen(true);

  const tableColumns = useMemo<Column<InventoryItem>[]>(() => [
    ...columns,
    {
      id: 'actions',
      header: '',
      width: 72,
      cell: (row) =>
        row.status !== 'voided' ? (
          <RowActions
            actions={[
              {
                label: 'Ajustar stock',
                icon: Pencil,
                onSelect: () => setAdjustTarget(row),
              },
              {
                label: 'Anular lote',
                icon: Ban,
                danger: true,
                onSelect: () => setVoidTarget(row),
              },
            ]}
            label={`Acciones de ${row.productDescription}`}
          />
        ) : null,
    },
  ], []);

  return (
    <PageContainer>
      <PageHeader
        title="Inventario"
        subtitle="Lotes, existencias y control de remate"
        actions={
          <>
            <Button variant="outline" onClick={() => setSettingsOpen(true)}>
              <Settings2 /> Reglas
            </Button>
            <Button onClick={openCreate}>
              <Plus /> Nuevo ingreso
            </Button>
          </>
        }
      />

      <Grid cols={5}>
        <StatCard label="Lotes en almacén" value={String(live.length)} icon={Boxes} />
        <StatCard label="Unidades en stock" value={formatNumber(totalUnits)} />
        <StatCard label="Valor de inventario" value={formatCurrency(inventoryValue)} />
        <StatCard label="En remate" value={String(clearance)} icon={AlertTriangle} />
        <StatCard label="Por vencer (5 días)" value={String(expiringSoon)} />
      </Grid>

      {clearance > 0 && (
        <div className="hstack" style={{ gap: '0.75rem', marginBottom: '1rem' }}>
          <Button
            variant={showClearanceOnly ? 'primary' : 'outline'}
            size="sm"
            onClick={() => setShowClearanceOnly(!showClearanceOnly)}
          >
            <AlertTriangle />
            {showClearanceOnly ? 'Mostrando productos en remate' : `Ver productos en remate (${clearance})`}
          </Button>
        </div>
      )}

      <DataTable
        columns={tableColumns}
        data={filteredItems}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        preferencesKey="inventory"
        toolbarLeft={
          <Select
            items={[
              { value: 'all', label: 'Todos los lotes' },
              { value: 'clearance', label: 'En remate' },
              { value: 'expiring', label: 'Por vencer (5 días)' },
              { value: 'voided', label: 'Anulados' },
            ]}
            value={statusFilter}
            onValueChange={(v) => setStatusFilter(v ?? 'all')}
          >
            <SelectTrigger style={{ width: '13rem' }} aria-label="Filtrar por estado">
              <SelectValue placeholder="Todos los lotes" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Todos los lotes</SelectItem>
              <SelectItem value="clearance">En remate</SelectItem>
              <SelectItem value="expiring">Por vencer (5 días)</SelectItem>
              <SelectItem value="voided">Anulados</SelectItem>
            </SelectContent>
          </Select>
        }
        empty={
          <EmptyState
            title="Sin existencias en inventario"
            description="Registra tu primer ingreso de stock; los lotes llegarán desde compras."
            action={{ label: 'Nuevo ingreso', onClick: openCreate }}
          />
        }
      />

      <InventoryReceiveDialog open={receiveOpen} onOpenChange={setReceiveOpen} />
      <InventoryAdjustDialog open={!!adjustTarget} onOpenChange={(o) => { if (!o) setAdjustTarget(null); }} batch={adjustTarget} />
      <InventorySettingsDrawer open={settingsOpen} onOpenChange={setSettingsOpen} />

      <ConfirmDialog
        open={!!voidTarget}
        onOpenChange={(open) => {
          if (!open) setVoidTarget(null);
        }}
        title="Anular lote"
        description={
          voidTarget
            ? `Se anulará el ingreso de ${formatNumber(voidTarget.quantity)} unidades de ${voidTarget.productDescription}. El lote quedará marcado como anulado y no se podrá editar ni vender. Esta acción no se puede deshacer.`
            : undefined
        }
        confirmLabel="Anular"
        loading={voidStock.isPending}
        onConfirm={() => {
          if (!voidTarget) return;
          voidStock.mutate(
            { batchId: voidTarget.id, reason: 'Anulado por error en el ingreso' },
            {
              onSuccess: () => {
                push({ title: 'Lote anulado', variant: 'success' });
                setVoidTarget(null);
              },
              onError: (err: unknown) => {
                push({ title: 'No se pudo anular el lote', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
                setVoidTarget(null);
              },
            },
          );
        }}
      />
    </PageContainer>
  );
}
