import { useMemo, useState } from 'react';
import { Package, TrendingUp, Eye, Download, AlertTriangle, Ban, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { SearchInput } from '@/components/input';
import { CancelDialog } from '@/components/dialog';
import { RowActions, type RowAction } from '@/components/misc';
import { useDebounce } from '@/utils/debounce';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
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
    cell: (row) => <span className="fw-medium tabular">{row.number}</span>,
  },
  { id: 'customerName', header: 'Cliente', sortable: true, cell: (row) => row.customerName || '—' },
  {
    id: 'supplierOrderNumber',
    header: 'Orden Proveedor',
    cell: (row) => <span className="muted">{row.supplierOrderNumber || '—'}</span>,
  },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="muted">{formatDate(row.date)}</span>,
  },
  {
    id: 'salePricePEN',
    header: 'Venta',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular">{formatCurrency(row.salePricePEN)}</span>,
  },
  {
    id: 'realCostPEN',
    header: 'Costo real (PEN)',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular">{formatCurrency(row.realCostPEN)}</span>,
  },
  {
    id: 'anticipo',
    header: 'Anticipo',
    align: 'right',
    cell: (row) => <span className="tabular">{formatCurrency(row.anticipo)}</span>,
  },
  {
    id: 'porCobrar',
    header: 'Por cobrar',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="fw-medium tabular">{formatCurrency(row.porCobrar)}</span>,
  },
  {
    id: 'projectedProfitPEN',
    header: 'Utilidad proy.',
    align: 'right',
    cell: (row) => (
      <span className={`tabular ${row.projectedProfitPEN < 0 ? 'text-destructive' : 'text-success'}`}>
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
        <div className="hstack hstack--sm">
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

  const cardOptions = useMemo<SelectOption[]>(() => {
    const opts: SelectOption[] = [];
    for (const c of cardsQuery.data ?? []) {
      if (c.isActive) opts.push({ value: c.id, label: `${c.issuer} •••• ${c.lastFour} (${c.currencyCode})` });
    }
    return opts;
  }, [cardsQuery.data]);

  const [formOpen, setFormOpen] = useState(false);
  const [detailOrder, setDetailOrder] = useState<CustomerOrder | null>(null);
  const [cancelTarget, setCancelTarget] = useState<CustomerOrder | null>(null);
  const [faultyTarget, setFaultyTarget] = useState<CustomerOrder | null>(null);
  const [arrivalPayTarget, setArrivalPayTarget] = useState<CustomerOrder | null>(null);

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

  const openCreate = () => setFormOpen(true);

  const tableColumns = useMemo<Column<CustomerOrder>[]>(() => {
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => {
          const open = row.status !== 'cancelled';
          const receivable = !row.arrivalDate && !row.faulty && row.status !== 'cancelled';
          const canMarkFaulty = open && !row.faulty;
          const actions: RowAction[] = [
            {
              label: 'Ver detalle',
              icon: Eye,
              onSelect: () => setDetailOrder(row),
            },
          ];
          if (receivable) {
            actions.push({
              label: 'Llegada y cobro',
              icon: Download,
              onSelect: () => setArrivalPayTarget(row),
            });
          }
          if (canMarkFaulty) {
            actions.push({
              label: 'Mal estado',
              icon: AlertTriangle,
              onSelect: () => setFaultyTarget(row),
            });
          }
          if (open) {
            actions.push({
              label: 'Anular',
              icon: Ban,
              danger: true,
              onSelect: () => setCancelTarget(row),
            });
          }
          return <RowActions actions={actions} label={`Acciones de ${row.number}`} />;
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
          <Button onClick={openCreate}>
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

      <DataTable
        columns={tableColumns}
        data={orders}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        preferencesKey="customer-orders"
        toolbarLeft={
          <>
            <SearchInput
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onClear={() => setSearchInput('')}
              placeholder="Número u orden del proveedor…"
              className="datatable-search"
              aria-label="Buscar pedido"
            />
            <Select
              items={[
                { value: 'all', label: 'Todos' },
                { value: 'pending', label: 'Con deuda' },
                { value: 'paid', label: 'Cancelados' },
              ]}
              value={debtFilter}
              onValueChange={(v) => setDebtFilter((v ?? 'all') as DebtFilter)}
            >
              <SelectTrigger style={{ width: '10rem' }} aria-label="Filtrar por cobranza">
                <SelectValue placeholder="Todos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Todos</SelectItem>
                <SelectItem value="pending">Con deuda</SelectItem>
                <SelectItem value="paid">Cancelados</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        empty={
          <EmptyState
            title="No hay pedidos de cliente"
            description="Crea el primer pedido de importación con anticipo para tus clientes."
            action={{ label: 'Nuevo pedido', onClick: openCreate }}
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
        key={cancelTarget?.id ?? 'cancel'}
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
