import { useMemo, useState } from 'react';
import { ShoppingCart, CreditCard, Ban, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { SaleStatusBadge } from '@/components/badge';
import { SearchInput } from '@/components/input';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { CancelDialog, RegisterPaymentDialog, type RegisterPaymentInput } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
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
    cell: (row) => <span className="fw-medium tabular">{row.number}</span>,
  },
  { id: 'customerName', header: 'Cliente', sortable: true, cell: (row) => row.customerName || '—' },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="muted">{formatDate(row.date)}</span>,
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => <SaleStatusBadge status={row.status} />,
  },
  {
    id: 'total',
    header: 'Total',
    align: 'numeric',
    sortable: true,
    cell: (row) => <span className="fw-medium tabular">{formatCurrency(row.total)}</span>,
  },
  {
    id: 'profit',
    header: 'Utilidad',
    align: 'numeric',
    cell: (row) => (
      <span className={row.profit >= 0 ? 'tabular' : 'tabular text-destructive'}>
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

  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [formOpen, setFormOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Sale | null>(null);
  const [collectTarget, setCollectTarget] = useState<Sale | null>(null);

  const sales = data ?? [];
  const filtered = useMemo(() => {
    let rows = sales;
    if (statusFilter !== 'all') rows = rows.filter((x) => x.status === statusFilter);
    if (search) {
      const q = search.toLowerCase();
      rows = rows.filter(
        (x) =>
          x.number.toLowerCase().includes(q) ||
          (x.customerName ?? '').toLowerCase().includes(q),
      );
    }
    return rows;
  }, [sales, statusFilter, search]);

  const totalAmount = sales.reduce((s, x) => s + x.total, 0);
  const totalProfit = sales.reduce((s, x) => s + x.profit, 0);
  const pending = sales.filter((x) => x.status === 'pending' || x.status === 'partial').length;

  const openCreate = () => setFormOpen(true);

  const tableColumns = useMemo<Column<Sale>[]>(() => {
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => {
          const collectable = row.status === 'pending' || row.status === 'partial';
          const cancellable = row.status !== 'cancelled';
          const actions = [];
          if (collectable) {
            actions.push({
              label: 'Cobrar',
              icon: CreditCard,
              onSelect: () => setCollectTarget(row),
            });
          }
          if (cancellable) {
            actions.push({
              label: 'Anular',
              icon: Ban,
              danger: true,
              onSelect: () => setCancelTarget(row),
            });
          }
          if (actions.length === 0) return null;
          return <RowActions actions={actions} label={`Acciones de ${row.number}`} />;
        },
      },
    ];
  }, []);

  return (
    <PageContainer>
      <PageHeader
        title="Ventas"
        subtitle="Documentos de venta y estado de cobranza"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nueva venta
          </Button>
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
        data={filtered}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        preferencesKey="sales"
        toolbarLeft={
          <>
            <SearchInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onClear={() => setSearch('')}
              placeholder="Buscar venta…"
              className="datatable-search"
              aria-label="Buscar venta"
            />
            <Select
              items={[
                { value: 'all', label: 'Estado: todos' },
                { value: 'pending', label: 'Pendientes' },
                { value: 'partial', label: 'Parciales' },
                { value: 'paid', label: 'Pagadas' },
                { value: 'cancelled', label: 'Anuladas' },
              ]}
              value={statusFilter}
              onValueChange={(v) => setStatusFilter(v ?? 'all')}
            >
              <SelectTrigger style={{ width: '11rem' }} aria-label="Filtrar por estado">
                <SelectValue placeholder="Estado: todos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Estado: todos</SelectItem>
                <SelectItem value="pending">Pendientes</SelectItem>
                <SelectItem value="partial">Parciales</SelectItem>
                <SelectItem value="paid">Pagadas</SelectItem>
                <SelectItem value="cancelled">Anuladas</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        empty={
          <EmptyState
            title="No hay ventas registradas"
            description="Registra tu primera venta para llevar el control de cobranza y utilidad."
            action={{ label: 'Nueva venta', onClick: openCreate }}
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
        key={cancelTarget?.id ?? 'cancel'}
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
