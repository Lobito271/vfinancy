import { useMemo, useState } from 'react';
import { Package, CreditCard, AlertTriangle, Ban, Plus, Download } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { SearchInput } from '@/components/input';
import { CancelDialog, RegisterPaymentDialog, type RegisterPaymentInput } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import { useDebounce } from '@/utils/debounce';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import {
  usePurchases,
  useCancelPurchase,
  useMarkPurchasePaid,
  useMarkPurchaseReceived,
  useMarkPurchaseFaulty,
} from '@/features/purchasing/hooks/usePurchases';
import { useCreditCards } from '@/features/treasury/hooks/useTreasury';
import type { SelectOption } from '@/components/form';
import { PurchaseFormDialog } from '@/features/purchasing/components/PurchaseFormDialog';
import { MarkReceivedDialog } from '@/features/purchasing/components/MarkReceivedDialog';
import { MarkFaultyDialog, type MarkFaultyInput } from '@/features/purchasing/components/MarkFaultyDialog';
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
    cell: (row) => <span className="fw-medium tabular">{row.number}</span>,
  },
  { id: 'supplierName', header: 'Proveedor', sortable: true, cell: (row) => row.supplierName || '—' },
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
    id: 'realCostPEN',
    header: 'Costo real (PEN)',
    sortable: true,
    align: 'right',
    cell: (row) => <span className="tabular">{formatCurrency(row.realCostPEN)}</span>,
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
  {
    id: 'total',
    header: 'Total',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="fw-medium tabular">{formatCurrency(row.total)}</span>,
  },
];

export function PurchasesPage() {
  const [searchInput, setSearchInput] = useState('');
  const search = useDebounce(searchInput);
  const [statusFilter, setStatusFilter] = useState('all');
  const { data, isLoading, isError, error, refetch } = usePurchases(search);
  const cancel = useCancelPurchase();
  const markPaid = useMarkPurchasePaid();
  const markReceived = useMarkPurchaseReceived();
  const markFaulty = useMarkPurchaseFaulty();
  const cardsQuery = useCreditCards();
  const push = useNotificationStore((s) => s.push);

  const cardOptions = useMemo<SelectOption[]>(() => {
    const opts: SelectOption[] = [];
    for (const c of cardsQuery.data ?? []) {
      if (c.isActive && c.currencyCode === 'USD') opts.push({ value: c.id, label: `${c.issuer} •••• ${c.lastFour} (${c.currencyCode})` });
    }
    return opts;
  }, [cardsQuery.data]);

  const [formOpen, setFormOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Purchase | null>(null);
  const [payTarget, setPayTarget] = useState<Purchase | null>(null);
  const [receivedTarget, setReceivedTarget] = useState<Purchase | null>(null);
  const [faultyTarget, setFaultyTarget] = useState<Purchase | null>(null);

  const purchases = data ?? [];
  const filtered = useMemo(() => {
    if (statusFilter === 'all') return purchases;
    return purchases.filter((p) => p.status === statusFilter);
  }, [purchases, statusFilter]);

  const totalAmount = purchases.reduce((s, p) => s + p.total, 0);
  const pending = purchases.filter((p) => p.status === 'pending' || p.status === 'received').length;
  const cancelled = purchases.filter((p) => p.status === 'cancelled').length;

  const openCreate = () => setFormOpen(true);

  const tableColumns = useMemo<Column<Purchase>[]>(() => {
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => {
          const open = row.status !== 'cancelled';
          const receivable = !row.arrivalDate && !row.faulty && row.status !== 'cancelled';
          const payable = row.status === 'pending' || row.status === 'received';
          const actions = [];
          if (receivable) {
            actions.push({
              label: 'Marcar como recibido',
              icon: Download,
              onSelect: () => setReceivedTarget(row),
            });
          }
          if (open && !row.faulty) {
            actions.push({
              label: 'Mal estado',
              icon: AlertTriangle,
              onSelect: () => setFaultyTarget(row),
            });
          }
          if (payable) {
            actions.push({
              label: 'Registrar pago',
              icon: CreditCard,
              onSelect: () => setPayTarget(row),
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
          if (actions.length === 0) return null;
          return <RowActions actions={actions} label={`Acciones de ${row.number}`} />;
        },
      },
    ];
  }, []);

  return (
    <PageContainer>
      <PageHeader
        title="Compras"
        subtitle="Órdenes de compra a proveedores"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nueva compra
          </Button>
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
        data={filtered}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        globalSearch={false}
        preferencesKey="purchases"
        toolbarLeft={
          <>
            <SearchInput
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onClear={() => setSearchInput('')}
              placeholder="Número u orden del proveedor…"
              className="datatable-search"
              aria-label="Buscar compra"
            />
            <Select
              items={[
                { value: 'all', label: 'Estado: todos' },
                { value: 'pending', label: 'Pendientes' },
                { value: 'received', label: 'Recibidas' },
                { value: 'paid', label: 'Pagadas' },
                { value: 'reconciled', label: 'Conciliadas' },
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
                <SelectItem value="received">Recibidas</SelectItem>
                <SelectItem value="paid">Pagadas</SelectItem>
                <SelectItem value="reconciled">Conciliadas</SelectItem>
                <SelectItem value="cancelled">Anuladas</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        empty={
          <EmptyState
            title="No hay órdenes de compra"
            description="Crea tu primera orden de compra para abastecer el inventario."
            action={{ label: 'Nueva compra', onClick: openCreate }}
          />
        }
      />

      <PurchaseFormDialog open={formOpen} onOpenChange={setFormOpen} />

      <MarkReceivedDialog
        open={!!receivedTarget}
        onOpenChange={(open) => {
          if (!open) setReceivedTarget(null);
        }}
        documentNumber={receivedTarget?.number ?? ''}
        loading={markReceived.isPending}
        onConfirm={({ arrivalDate }) => {
          if (!receivedTarget) return;
          markReceived.mutate(
            { id: receivedTarget.id, arrivalDate },
            {
              onSuccess: () => {
                push({ title: 'Pedido marcado como recibido', variant: 'success' });
                setReceivedTarget(null);
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
        }}
      />

      <MarkFaultyDialog
        open={!!faultyTarget}
        onOpenChange={(open) => {
          if (!open) setFaultyTarget(null);
        }}
        documentNumber={faultyTarget?.number ?? ''}
        loading={markFaulty.isPending}
        onConfirm={(input: MarkFaultyInput) => {
          if (!faultyTarget) return;
          markFaulty.mutate(
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
        }}
      />

      <RegisterPaymentDialog
        open={!!payTarget}
        onOpenChange={(open) => {
          if (!open) setPayTarget(null);
        }}
        title="Registrar pago"
        description="Registra el pago de la orden de compra al proveedor."
        documentNumber={payTarget?.number ?? ''}
        amount={payTarget?.total ?? 0}
        amountLabel="Total de la orden"
        confirmLabel="Registrar pago"
        loading={markPaid.isPending}
        currencyCode="USD"
        creditCardOptions={cardOptions}
        creditCardLoading={cardsQuery.isLoading}
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
        key={cancelTarget?.id ?? 'cancel'}
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
