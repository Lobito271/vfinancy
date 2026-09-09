import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Package, AlertTriangle, Ban, Plus, Download, Boxes, Filter } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { Button } from '@/components/button';
import { Input, Label, SearchInput } from '@/components/input';
import { CancelDialog } from '@/components/dialog';
import { Drawer, RowActions } from '@/components/misc';
import { useDebounce } from '@/hooks/useDebounce';
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
  useMarkPurchaseReceived,
  useMarkPurchaseFaulty,
} from '@/features/purchasing/hooks/usePurchases';
import { PurchaseFormDialog } from '@/features/purchasing/components/PurchaseFormDialog';
import { ImportLotsDialog } from '@/features/purchasing/components/ImportLotsDialog';
import { MarkReceivedDialog } from '@/features/purchasing/components/MarkReceivedDialog';
import { MarkFaultyDialog, type MarkFaultyInput } from '@/features/purchasing/components/MarkFaultyDialog';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import type { CreditCardDTO, ImportLotDTO } from '@/services/wails-types';
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
  { id: 'date', header: 'Fecha', cell: (row) => <span className="muted">{formatDate(row.date)}</span> },
  {
    id: 'realCostPen',
    header: 'Costo real (PEN)',
    sortable: true,
    align: 'numeric',
    cell: (row) => <span className="tabular">{formatCurrency(row.realCostPen)}</span>,
  },
  {
    id: 'projectedProfitPen',
    header: 'Utilidad proy.',
    align: 'numeric',
    cell: (row) => (
      <span className={`tabular ${row.projectedProfitPen < 0 ? 'text-destructive' : 'text-success'}`}>
        {formatCurrency(row.projectedProfitPen)}
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
    id: 'costUsd',
    header: 'Costo (USD)',
    align: 'numeric',
    sortable: true,
    cell: (row) => <span className="fw-medium tabular">{formatCurrency(row.costUsd, 'USD')}</span>,
  },
];

export function PurchasesPage() {
  const [searchInput, setSearchInput] = useState('');
  const search = useDebounce(searchInput);
  const [statusFilter, setStatusFilter] = useState('all');
  const [importLotId, setImportLotId] = useState('');
  const [creditCardId, setCreditCardId] = useState('');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const { data, isLoading, isError, error, refetch } = usePurchases({ search, importLotId, creditCardId, from, to });
  const cancel = useCancelPurchase();
  const markReceived = useMarkPurchaseReceived();
  const markFaulty = useMarkPurchaseFaulty();
  const push = useNotificationStore((s) => s.push);

  const [formOpen, setFormOpen] = useState(false);
  const [lotsOpen, setLotsOpen] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Purchase | null>(null);
  const [receivedTarget, setReceivedTarget] = useState<Purchase | null>(null);
  const [faultyTarget, setFaultyTarget] = useState<Purchase | null>(null);

  const lotsQuery = useQuery({
    queryKey: queryKeys.importLots.list({ page: 1, pageSize: 100, search: '' }),
    queryFn: () => wailsClient.listImportLots({ page: 1, pageSize: 100 }, ''),
    enabled: filtersOpen,
  });
  const cardsQuery = useQuery({
    queryKey: queryKeys.treasury.creditCards,
    queryFn: () => wailsClient.listCreditCards(),
    enabled: filtersOpen,
  });

  const activeFilterCount = [importLotId, creditCardId, from, to].filter(Boolean).length;
  const clearFilters = () => {
    setImportLotId('');
    setCreditCardId('');
    setFrom('');
    setTo('');
  };

  const purchases = data ?? [];
  const filtered = useMemo(() => {
    if (statusFilter === 'all') return purchases;
    return purchases.filter((p) => p.status === statusFilter);
  }, [purchases, statusFilter]);

  const totalAmount = purchases.reduce((s, p) => s + p.costUsd, 0);
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
          <div className="hstack hstack--sm">
            <Button variant="outline" onClick={() => setFiltersOpen(true)}>
              <Filter /> Filtros{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
            </Button>
            <Button variant="outline" onClick={() => setLotsOpen(true)}>
              <Boxes /> Lotes
            </Button>
            <Button onClick={openCreate}>
              <Plus /> Nueva compra
            </Button>
          </div>
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

      <Drawer
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        title="Filtros avanzados"
        description="Combina lote, rango de fechas y tarjeta de crédito."
        footer={
          <div className="hstack hstack--sm">
            <Button variant="outline" onClick={clearFilters} disabled={activeFilterCount === 0}>
              Limpiar
            </Button>
            <Button onClick={() => setFiltersOpen(false)}>Aplicar</Button>
          </div>
        }
      >
        <div className="stack">
          <div className="field">
            <Label htmlFor="purchase-filter-lot">Lote de importación</Label>
            <Select
              items={[{ value: '', label: 'Todos los lotes' }, ...(lotsQuery.data?.items ?? []).map((l: ImportLotDTO) => ({ value: l.id, label: `${l.code} · ${l.description || 'sin descripción'}` }))]}
              value={importLotId}
              onValueChange={(v) => setImportLotId(v ?? '')}
            >
              <SelectTrigger id="purchase-filter-lot" aria-label="Filtrar por lote">
                <SelectValue placeholder="Todos los lotes" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">Todos los lotes</SelectItem>
                {(lotsQuery.data?.items ?? []).map((l: ImportLotDTO) => (
                  <SelectItem key={l.id} value={l.id}>
                    {l.code} · {l.description || 'sin descripción'}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="field">
            <Label>Rango de fechas</Label>
            <div className="hstack hstack--sm">
              <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} aria-label="Desde" />
              <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} aria-label="Hasta" />
            </div>
          </div>
          <div className="field">
            <Label htmlFor="purchase-filter-card">Tarjeta de crédito</Label>
            <Select
              items={[{ value: '', label: 'Todas las tarjetas' }, ...(cardsQuery.data ?? []).map((c: CreditCardDTO) => ({ value: c.id, label: `${c.issuer} •••• ${c.lastFour}` }))]}
              value={creditCardId}
              onValueChange={(v) => setCreditCardId(v ?? '')}
            >
              <SelectTrigger id="purchase-filter-card" aria-label="Filtrar por tarjeta">
                <SelectValue placeholder="Todas las tarjetas" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">Todas las tarjetas</SelectItem>
                {(cardsQuery.data ?? []).map((c: CreditCardDTO) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.issuer} •••• {c.lastFour}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </Drawer>

      <PurchaseFormDialog open={formOpen} onOpenChange={setFormOpen} />
      <ImportLotsDialog open={lotsOpen} onOpenChange={setLotsOpen} />

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
            { id: receivedTarget.id, receivedDate: arrivalDate },
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
            { id: faultyTarget.id, reason: input.reason },
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
