import { useMemo, useState } from 'react';
import { Package, TrendingUp, Eye, Download, AlertTriangle, Trash2, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { CancelDialog } from '@/components/dialog';
import { useDebounce } from '@/utils/debounce';
import {
  useCustomerOrders,
  useCancelCustomerOrder,
  useMarkCustomerOrderFaulty,
  useMarkPurchaseReceived,
  useRegisterCustomerOrderPayment,
} from '@/features/purchasing/hooks/usePurchases';
import { useCreditCards } from '@/features/treasury/hooks/useTreasury';
import type { SelectOption } from '@/components/form';
import { CreateCustomerOrderDialog } from '@/features/purchasing/components/CreateCustomerOrderDialog';
import { CustomerOrderDetailDialog } from '@/features/purchasing/components/CustomerOrderDetailDialog';
import { ArrivalAndPaymentDialog, type ArrivalAndPaymentInput } from '@/features/purchasing/components/ArrivalAndPaymentDialog';
import { MarkFaultyDialog, type MarkFaultyInput } from '@/features/purchasing/components/MarkFaultyDialog';
import type { CustomerOrder } from '@/types/domain';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const statusMap: Record<string, { variant: 'success' | 'warning' | 'info' | 'destructive' | 'muted'; label: string }> = {
  pending: { variant: 'warning', label: 'Pendiente' },
  received: { variant: 'info', label: 'Recibida' },
  paid: { variant: 'success', label: 'Pagada' },
  reconciled: { variant: 'success', label: 'Conciliada' },
  cancelled: { variant: 'destructive', label: 'Anulada' },
};

const columns: Column<CustomerOrder>[] = [
  {
    id: 'number',
    header: 'N° Pedido',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="font-medium tabular-nums">{row.number}</span>,
  },
  { id: 'customerName', header: 'Cliente', sortable: true, cell: (row) => row.customerName || '—' },
  {
    id: 'supplierOrderNumber',
    header: 'Orden Proveedor',
    cell: (row) => <span className="text-muted-foreground">{row.supplierOrderNumber || '—'}</span>,
  },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="text-muted-foreground">{formatDate(row.date)}</span>,
  },
  {
    id: 'salePricePEN',
    header: 'Venta',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.salePricePEN)}</span>,
  },
  {
    id: 'realCostPEN',
    header: 'Costo real (PEN)',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.realCostPEN)}</span>,
  },
  {
    id: 'anticipo',
    header: 'Anticipo',
    align: 'right',
    cell: (row) => <span className="tabular-nums">{formatCurrency(row.anticipo)}</span>,
  },
  {
    id: 'porCobrar',
    header: 'Por cobrar',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular-nums font-medium">{formatCurrency(row.porCobrar)}</span>,
  },
  {
    id: 'projectedProfitPEN',
    header: 'Utilidad proy.',
    align: 'right',
    cell: (row) => (
      <span className={`tabular-nums ${row.projectedProfitPEN < 0 ? 'text-destructive' : 'text-success'}`}>
        {formatCurrency(row.projectedProfitPEN)}
      </span>
    ),
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => {
      const cfg = statusMap[row.status] ?? { variant: 'muted' as const, label: row.status };
      return (
        <div className="flex flex-wrap items-center gap-1">
          <Badge variant={cfg.variant}>{cfg.label}</Badge>
          {row.faulty && <Badge variant="destructive">Defectuoso</Badge>}
        </div>
      );
    },
  },
];

type DebtFilter = 'all' | 'pending' | 'paid';

export function CustomerOrdersPage() {
  const [searchInput, setSearchInput] = useState('');
  const search = useDebounce(searchInput);
  const [debtFilter, setDebtFilter] = useState<DebtFilter>('all');
  const { data, isLoading, isError, error, refetch } = useCustomerOrders(search);
  const cancel = useCancelCustomerOrder();
  const faulty = useMarkCustomerOrderFaulty();
  const pay = useRegisterCustomerOrderPayment();
  const markReceived = useMarkPurchaseReceived();
  const cardsQuery = useCreditCards();
  const push = useNotificationStore((s) => s.push);

  const cardOptions = useMemo<SelectOption[]>(
    () =>
      (cardsQuery.data ?? [])
        .filter((c) => c.isActive)
        .map((c) => ({
          value: c.id,
          label: `${c.issuer} •••• ${c.lastFour} (${c.currencyCode})`,
        })),
    [cardsQuery.data],
  );

  const [formOpen, setFormOpen] = useState(false);
  const [detailOrder, setDetailOrder] = useState<CustomerOrder | null>(null);
  const [cancelTarget, setCancelTarget] = useState<CustomerOrder | null>(null);
  const [faultyTarget, setFaultyTarget] = useState<CustomerOrder | null>(null);
  const [arrivalPayTarget, setArrivalPayTarget] = useState<CustomerOrder | null>(null);

  const canEdit = true;
  const canDelete = true;

  const orders = useMemo(() => {
    const all = data ?? [];
    if (debtFilter === 'pending') return all.filter((o) => o.porCobrar > 0 && o.status !== 'cancelled');
    if (debtFilter === 'paid') return all.filter((o) => o.porCobrar === 0 || o.status === 'cancelled');
    return all;
  }, [data, debtFilter]);
  const totalSale = orders.reduce((s, o) => s + o.salePricePEN, 0);
  const totalAnticipo = orders.reduce((s, o) => s + o.anticipo, 0);
  const totalPorCobrar = orders.reduce((s, o) => s + o.porCobrar, 0);
  const totalProfit = orders.reduce((s, o) => s + o.projectedProfitPEN, 0);

  const tableColumns = useMemo<Column<CustomerOrder>[]>(() => {
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 280,
        exportable: false,
        cell: (row) => {
          const open = row.status !== 'cancelled';
          const receivable = !row.arrivalDate && !row.faulty && row.status !== 'cancelled';
          const canMarkFaulty = open && !row.faulty;
          return (
            <div className="flex items-center justify-end gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Ver ${row.number}`}
                onClick={() => setDetailOrder(row)}
              >
                <Eye />
              </Button>
              {canEdit && receivable && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setArrivalPayTarget(row)}
                  title="Registra la llegada y cobra el saldo pendiente en un solo paso"
                >
                  <Download /> Llegada y Cobro
                </Button>
              )}
              {canEdit && canMarkFaulty && (
                <Button
                  variant="outline"
                  size="sm"
                  aria-label={`Llegó en mal estado ${row.number}`}
                  onClick={() => setFaultyTarget(row)}
                  title="Anula el pedido y reembolsa el anticipo"
                >
                  <AlertTriangle /> Mal estado
                </Button>
              )}
              {canDelete && open && (
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Anular ${row.number}`}
                  onClick={() => setCancelTarget(row)}
                >
                  <Trash2 />
                </Button>
              )}
            </div>
          );
        },
      },
    ];
  }, []);

  const handleArrivalAndPayment = (input: ArrivalAndPaymentInput) => {
    if (!arrivalPayTarget) return;
    markReceived.mutate(
      { id: arrivalPayTarget.id, arrivalDate: input.arrivalDate },
      {
        onSuccess: () => {
          pay.mutate(
            {
              id: arrivalPayTarget.id,
              input: {
                paymentDate: input.arrivalDate,
                amount: input.amount,
                method: 'card',
                reference: input.reference,
                notes: input.notes,
              },
            },
            {
              onSuccess: () => {
                push({ title: 'Llegada registrada y cobro exitoso', variant: 'success' });
                setArrivalPayTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'Llegada registrada, pero falló el cobro',
                  description: err instanceof Error ? err.message : undefined,
                  variant: 'destructive',
                });
                setArrivalPayTarget(null);
              },
            },
          );
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo registrar la llegada',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          });
        },
      },
    );
  };

  const handleFaulty = (input: MarkFaultyInput) => {
    if (!faultyTarget) return;
    faulty.mutate(
      { id: faultyTarget.id, input },
      {
        onSuccess: () => {
          push({ title: 'Pedido marcado como defectuoso', variant: 'success' });
          setFaultyTarget(null);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo marcar el pedido',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          });
        },
      },
    );
  };

  const handleCancel = (reason: string) => {
    if (!cancelTarget) return;
    cancel.mutate(
      { id: cancelTarget.id, reason },
      {
        onSuccess: () => {
          push({ title: 'Pedido anulado', variant: 'success' });
          setCancelTarget(null);
        },
        onError: (err: unknown) => {
          push({
            title: 'No se pudo anular el pedido',
            description: err instanceof Error ? err.message : undefined,
            variant: 'destructive',
          });
          setCancelTarget(null);
        },
      },
    );
  };

  return (
    <PageContainer>
      <PageHeader
        title="Pedidos de cliente"
        subtitle="Órdenes de importación para clientes con anticipos y utilidad proyectada"
        actions={
          <Button onClick={() => setFormOpen(true)}>
            <Plus /> Nuevo pedido
          </Button>
        }
      />

      <Grid cols={5}>
        <StatCard label="Pedidos" value={String(orders.length)} icon={Package} />
        <StatCard label="Ventas proyectadas" value={formatCurrency(totalSale)} />
        <StatCard label="Anticipos cobrados" value={formatCurrency(totalAnticipo)} />
        <StatCard label="Por cobrar" value={formatCurrency(totalPorCobrar)} />
        <StatCard label="Utilidad proyectada" value={formatCurrency(totalProfit)} icon={TrendingUp} />
      </Grid>

      <div className="flex items-center gap-3">
        <input
          type="text"
          placeholder="Buscar por número o orden del proveedor…"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          className="h-9 w-full max-w-sm rounded-md border bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
        />
        <select
          value={debtFilter}
          onChange={(e) => setDebtFilter(e.target.value as DebtFilter)}
          className="h-9 rounded-md border bg-background px-2 text-sm outline-none focus:ring-2 focus:ring-ring"
        >
          <option value="all">Todos</option>
          <option value="pending">Con Deuda</option>
          <option value="paid">Cancelados</option>
        </select>
      </div>

      <DataTable
        columns={tableColumns}
        data={orders}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        exportFilename="pedidos-cliente.csv"
        empty={
          <EmptyState
            title="No hay pedidos de cliente"
            description="Los pedidos de importación para clientes aparecerán aquí con su anticipo y utilidad proyectada."
          />
        }
      />

      <CreateCustomerOrderDialog open={formOpen} onOpenChange={setFormOpen} />

      <CustomerOrderDetailDialog
        order={detailOrder}
        onOpenChange={(open) => !open && setDetailOrder(null)}
        onMarkReceived={(o) => setArrivalPayTarget(o)}
        onMarkFaulty={(o) => setFaultyTarget(o)}
      />

      <ArrivalAndPaymentDialog
        open={!!arrivalPayTarget}
        onOpenChange={(open) => {
          if (!open) setArrivalPayTarget(null);
        }}
        documentNumber={arrivalPayTarget?.number ?? ''}
        outstandingAmount={arrivalPayTarget?.porCobrar ?? 0}
        loading={markReceived.isPending || pay.isPending}
        creditCardOptions={cardOptions}
        creditCardLoading={cardsQuery.isLoading}
        onConfirm={handleArrivalAndPayment}
      />

      <MarkFaultyDialog
        open={!!faultyTarget}
        onOpenChange={(open) => {
          if (!open) setFaultyTarget(null);
        }}
        documentNumber={faultyTarget?.number ?? ''}
        loading={faulty.isPending}
        onConfirm={handleFaulty}
      />

      <CancelDialog
        open={!!cancelTarget}
        onOpenChange={(open) => {
          if (!open) setCancelTarget(null);
        }}
        title="Anular pedido de cliente"
        description={`Se anulará el pedido ${cancelTarget?.number ?? ''} y se reembolsarán los anticipos registrados. Esta acción no se puede deshacer.`}
        loading={cancel.isPending}
        onConfirm={handleCancel}
      />
    </PageContainer>
  );
}
