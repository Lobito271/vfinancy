import { useMemo, useState } from 'react';
import { Package } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { CancelDialog, RegisterPaymentDialog, type RegisterPaymentInput } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
import { usePurchases, useCancelPurchase, useMarkPurchasePaid } from '@/features/purchasing/hooks/usePurchases';
import { PurchaseFormDialog } from '@/features/purchasing/components/PurchaseFormDialog';
import type { Purchase } from '@/types/domain';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const statusMap: Record<string, { variant: 'success' | 'warning' | 'info' | 'destructive' | 'muted'; label: string }> = {
  pending: { variant: 'warning', label: 'Pendiente' },
  received: { variant: 'info', label: 'Recibida' },
  paid: { variant: 'success', label: 'Pagada' },
  reconciled: { variant: 'success', label: 'Conciliada' },
  cancelled: { variant: 'destructive', label: 'Anulada' },
};

const columns: Column<Purchase>[] = [
  {
    id: 'number',
    header: 'Número',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="font-medium tabular-nums">{row.number}</span>,
  },
  { id: 'supplierName', header: 'Proveedor', sortable: true, cell: (row) => row.supplierName || '—' },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="text-muted-foreground">{formatDate(row.date)}</span>,
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => {
      const cfg = statusMap[row.status] ?? { variant: 'muted' as const, label: row.status };
      return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
    },
  },
  {
    id: 'total',
    header: 'Total',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="font-medium tabular-nums">{formatCurrency(row.total)}</span>,
  },
];

export function PurchasesPage() {
  const { data, isLoading, isError, error, refetch } = usePurchases();
  const cancel = useCancelPurchase();
  const markPaid = useMarkPurchasePaid();
  const push = useNotificationStore((s) => s.push);

  const [formOpen, setFormOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Purchase | null>(null);
  const [payTarget, setPayTarget] = useState<Purchase | null>(null);

  const canDelete = usePermission(Permissions.Purchases.Delete);
  const canPay = usePermission(Permissions.Purchases.Edit);

  const purchases = data ?? [];
  const totalAmount = purchases.reduce((s, p) => s + p.total, 0);
  const pending = purchases.filter((p) => p.status === 'pending' || p.status === 'received').length;
  const cancelled = purchases.filter((p) => p.status === 'cancelled').length;

  const tableColumns = useMemo<Column<Purchase>[]>(() => {
    if (!canDelete && !canPay) return columns;
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 120,
        exportable: false,
        cell: (row) => {
          const payable = row.status === 'pending' || row.status === 'received';
          const cancellable = row.status !== 'cancelled';
          return (
            <div className="flex items-center justify-end gap-1">
              {canPay && payable && (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Registrar pago de ${row.number}`}
                  onClick={() => setPayTarget(row)}
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
  }, [canDelete, canPay]);

  return (
    <PageContainer>
      <PageHeader
        title="Compras"
        subtitle="Órdenes de compra a proveedores"
        actions={
          <Can permission={Permissions.Purchases.Create}>
            <Button onClick={() => setFormOpen(true)}>
              <Icons.Action.Create /> Nueva compra
            </Button>
          </Can>
        }
      />

      <Grid cols={4}>
        <StatCard label="Órdenes de compra" value={String(purchases.length)} icon={Package} />
        <StatCard label="Monto total" value={formatCurrency(totalAmount)} />
        <StatCard label="Por Pagar" value={String(pending)} />
        <StatCard label="Anuladas" value={String(cancelled)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={purchases}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        exportFilename="compras.csv"
        empty={
          <EmptyState
            title="No hay órdenes de compra"
            description="Las órdenes creadas para los proveedores aparecerán aquí con su estado."
          />
        }
      />

      <PurchaseFormDialog open={formOpen} onOpenChange={setFormOpen} />

      <RegisterPaymentDialog
        open={!!payTarget}
        onOpenChange={(open) => {
          if (!open) setPayTarget(null);
        }}
        title="Registrar pago"
        description="Registra el pago total de la orden de compra."
        documentNumber={payTarget?.number ?? ''}
        amount={payTarget?.total ?? 0}
        amountLabel="Total de la orden"
        confirmLabel="Registrar pago"
        loading={markPaid.isPending}
        onConfirm={(input: RegisterPaymentInput) => {
          if (!payTarget) return;
          markPaid.mutate(
            { id: payTarget.id, input },
            {
              onSuccess: () => {
                push({ title: 'Pago registrado', variant: 'success' });
                setPayTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo registrar el pago',
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
        title="Anular orden de compra"
        description={`Se anulará la orden ${cancelTarget?.number ?? ''}. Esta acción no se puede deshacer.`}
        loading={cancel.isPending}
        onConfirm={(reason) => {
          if (!cancelTarget) return;
          cancel.mutate(
            { id: cancelTarget.id, reason },
            {
              onSuccess: () => {
                push({ title: 'Orden de compra anulada', variant: 'success' });
                setCancelTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo anular la orden',
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
