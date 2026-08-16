import { useMemo, useState } from 'react';
import { Boxes } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { ConfirmDialog } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
import { useInventory, useVoidStock } from '@/features/inventory/hooks/useInventory';
import { InventoryReceiveDialog } from '@/features/inventory/components/InventoryReceiveDialog';
import { InventoryAdjustDialog } from '@/features/inventory/components/InventoryAdjustDialog';
import type { InventoryItem } from '@/types/domain';
import { formatCurrency, formatDate, formatNumber } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<InventoryItem>[] = [
  {
    id: 'productSku',
    header: 'SKU',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="font-medium tabular-nums">{row.productSku}</span>,
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
    align: 'right',
    sortable: true,
    cell: (row) => <span className="tabular-nums">{formatNumber(row.quantity)}</span>,
  },
  {
    id: 'unitCost',
    header: 'Costo unitario',
    align: 'right',
    sortable: true,
    exportable: true,
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.unitCost, row.currencyCode)}</span>,
  },
  {
    id: 'totalCost',
    header: 'Costo total',
    align: 'right',
    sortable: true,
    exportable: true,
    accessor: (row) => row.quantity * row.unitCost,
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.quantity * row.unitCost, row.currencyCode)}</span>,
  },
  {
    id: 'arrivalDate',
    header: 'Ingreso',
    cell: (row) => <span className="text-muted-foreground">{row.arrivalDate ? formatDate(row.arrivalDate) : '—'}</span>,
  },
  {
    id: 'maxSaleDate',
    header: 'Venta máxima',
    cell: (row) => <span className="text-muted-foreground">{row.maxSaleDate ? formatDate(row.maxSaleDate) : '—'}</span>,
  },
  {
    id: 'daysRemaining',
    header: 'Días restantes',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="tabular-nums">{row.daysRemaining}</span>,
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => {
      if (row.status === 'voided') return <Badge variant="destructive">Anulado</Badge>;
      if (row.status === 'written_off') return <Badge variant="muted">Baja</Badge>;
      if (row.status === 'depleted') return <Badge variant="muted">Agotado</Badge>;
      if (row.isClearance) return <Badge variant="destructive">Remate</Badge>;
      return <Badge variant="success">Normal</Badge>;
    },
  },
];

export function InventoryPage() {
  const { data, isLoading, isError, error, refetch } = useInventory();
  const voidStock = useVoidStock();
  const push = useNotificationStore((s) => s.push);

  const [receiveOpen, setReceiveOpen] = useState(false);
  const [adjustTarget, setAdjustTarget] = useState<InventoryItem | null>(null);
  const [voidTarget, setVoidTarget] = useState<InventoryItem | null>(null);

  const canEdit = usePermission(Permissions.Inventory.Edit);
  const canDelete = usePermission(Permissions.Inventory.Delete);

  const items = data ?? [];
  const live = items.filter((i) => i.status !== 'voided');
  const totalUnits = live.reduce((s, i) => s + i.quantity, 0);
  const inventoryValue = live.reduce((s, i) => s + i.quantity * i.unitCost, 0);
  const clearance = items.filter((i) => i.isClearance).length;
  const expiringSoon = live.filter((i) => i.daysRemaining >= 0 && i.daysRemaining < 5).length;

  const tableColumns = useMemo<Column<InventoryItem>[]>(() => {
    if (!canEdit && !canDelete) return columns;
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 88,
        exportable: false,
        cell: (row) => (
          <div className="flex items-center justify-end gap-1">
            {canEdit && row.status !== 'voided' && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Ajustar ${row.productDescription}`}
                onClick={() => setAdjustTarget(row)}
              >
                <Icons.Action.Edit />
              </Button>
            )}
            {canDelete && row.status !== 'voided' && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Anular ${row.productDescription}`}
                onClick={() => setVoidTarget(row)}
              >
                <Icons.Action.Delete />
              </Button>
            )}
          </div>
        ),
      },
    ];
  }, [canEdit, canDelete]);

  return (
    <PageContainer>
      <PageHeader
        title="Inventario"
        subtitle="Lotes, existencias y control de remate"
        actions={
          <Can permission={Permissions.Inventory.Create}>
            <Button onClick={() => setReceiveOpen(true)}>
              <Icons.Action.Create /> Nuevo ingreso
            </Button>
          </Can>
        }
      />

      <Grid cols={5}>
        <StatCard label="Lotes en almacén" value={String(live.length)} icon={Boxes} />
        <StatCard label="Unidades en stock" value={formatNumber(totalUnits)} />
        <StatCard label="Valor de inventario" value={formatCurrency(inventoryValue)} />
        <StatCard label="En remate" value={String(clearance)} />
        <StatCard label="Por vencer (5 días)" value={String(expiringSoon)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={items}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        rowClassName={(row) => (row.status === 'voided' ? 'opacity-50' : undefined)}
        globalSearch={false}
        exportFilename="inventario.csv"
        empty={
          <EmptyState
            title="Sin existencias en inventario"
            description="Los lotes recibidos por compras aparecerán aquí con sus fechas de remate."
          />
        }
      />

      <InventoryReceiveDialog open={receiveOpen} onOpenChange={setReceiveOpen} />
      <InventoryAdjustDialog open={!!adjustTarget} onOpenChange={(o) => { if (!o) setAdjustTarget(null); }} batch={adjustTarget} />

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
