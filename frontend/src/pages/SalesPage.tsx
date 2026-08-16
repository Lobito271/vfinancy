import { useMemo, useState } from 'react';
import { ShoppingCart } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { SaleStatusBadge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { CancelDialog, RegisterPaymentDialog, type RegisterPaymentInput } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
import { useSales, useCancelSale, useCollectSalePayment } from '@/features/sales/hooks/useSales';
import { SaleFormDialog } from '@/features/sales/components/SaleFormDialog';
import type { Sale } from '@/types/domain';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<Sale>[] = [
  {
    id: 'number',
    header: 'Número',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="font-medium tabular-nums">{row.number}</span>,
  },
  { id: 'customerName', header: 'Cliente', sortable: true, cell: (row) => row.customerName || '—' },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="text-muted-foreground">{formatDate(row.date)}</span>,
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => <SaleStatusBadge status={row.status} />,
  },
  {
    id: 'total',
    header: 'Total',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="font-medium tabular-nums">{formatCurrency(row.total)}</span>,
  },
  {
    id: 'profit',
    header: 'Utilidad',
    align: 'right',
    cell: (row) => (
      <span className={row.profit >= 0 ? 'tabular-nums' : 'tabular-nums text-destructive'}>
        {formatCurrency(row.profit)}
      </span>
    ),
  },
];

export function SalesPage() {
  const { data, isLoading, isError, error, refetch } = useSales();
  const cancel = useCancelSale();
  const collect = useCollectSalePayment();
  const push = useNotificationStore((s) => s.push);

  const [formOpen, setFormOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Sale | null>(null);
  const [collectTarget, setCollectTarget] = useState<Sale | null>(null);

  const canDelete = usePermission(Permissions.Sales.Delete);
  const canCollect = usePermission(Permissions.Sales.Edit);

  const sales = data ?? [];
  const totalAmount = sales.reduce((s, x) => s + x.total, 0);
  const totalProfit = sales.reduce((s, x) => s + x.profit, 0);
  const pending = sales.filter((x) => x.status === 'pending' || x.status === 'partial').length;

  const tableColumns = useMemo<Column<Sale>[]>(() => {
    if (!canDelete && !canCollect) return columns;
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 120,
        exportable: false,
        cell: (row) => {
          const collectable = row.status === 'pending' || row.status === 'partial';
          const cancellable = row.status !== 'cancelled';
          return (
            <div className="flex items-center justify-end gap-1">
              {canCollect && collectable && (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Cobrar ${row.number}`}
                  onClick={() => setCollectTarget(row)}
                >
                  <Icons.Action.Payment />
                </Button>
              )}
              {canDelete && cancellable && (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Anular ${row.number}`}
                  onClick={() => setCancelTarget(row)}
                >
                  <Icons.Action.Delete />
                </Button>
              )}
            </div>
          );
        },
      },
    ];
  }, [canDelete, canCollect]);

  return (
    <PageContainer>
      <PageHeader
        title="Ventas"
        subtitle="Documentos de venta y estado de cobranza"
        actions={
          <Can permission={Permissions.Sales.Create}>
            <Button onClick={() => setFormOpen(true)}>
              <Icons.Action.Create /> Nueva venta
            </Button>
          </Can>
        }
      />

      <Grid cols={4}>
        <StatCard label="Ventas registradas" value={String(sales.length)} icon={ShoppingCart} />
        <StatCard label="Monto total" value={formatCurrency(totalAmount)} />
        <StatCard label="Utilidad" value={formatCurrency(totalProfit)} />
        <StatCard label="Por Cobrar" value={String(pending)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={sales}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        exportFilename="ventas.csv"
        empty={
          <EmptyState
            title="No hay ventas registradas"
            description="Las ventas generadas desde el sistema aparecerán aquí con su estado de cobro."
          />
        }
      />

      <SaleFormDialog open={formOpen} onOpenChange={setFormOpen} />

      <RegisterPaymentDialog
        open={!!collectTarget}
        onOpenChange={(open) => {
          if (!open) setCollectTarget(null);
        }}
        title="Cobrar venta"
        description="Registra el cobro total de la venta."
        documentNumber={collectTarget?.number ?? ''}
        amount={collectTarget?.total ?? 0}
        amountLabel="Total de la venta"
        confirmLabel="Cobrar"
        loading={collect.isPending}
        onConfirm={(input: RegisterPaymentInput) => {
          if (!collectTarget) return;
          collect.mutate(
            { id: collectTarget.id, input },
            {
              onSuccess: () => {
                push({ title: 'Cobro registrado', variant: 'success' });
                setCollectTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo registrar el cobro',
                  description: err instanceof Error ? err.message : undefined,
                  variant: 'destructive',
                });
              },
            },
          );
        }}
      />

      <CancelDialog
        open={!!cancelTarget}
        onOpenChange={(open) => {
          if (!open) setCancelTarget(null);
        }}
        title="Anular venta"
        description={`Se anulará la venta ${cancelTarget?.number ?? ''} y se revertirá el stock y la deuda asociada.`}
        loading={cancel.isPending}
        onConfirm={(reason) => {
          if (!cancelTarget) return;
          cancel.mutate(
            { id: cancelTarget.id, reason },
            {
              onSuccess: () => {
                push({ title: 'Venta anulada', variant: 'success' });
                setCancelTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo anular la venta',
                  description: err instanceof Error ? err.message : undefined,
                  variant: 'destructive',
                });
                setCancelTarget(null);
              },
            },
          );
        }}
      />
    </PageContainer>
  );
}
