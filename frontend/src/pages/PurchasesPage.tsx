import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Package, Ban, Plus, Download, Users, Filter, Eye, Pencil } from 'lucide-react';
import { z } from 'zod';
import { PageContainer, PageHeader, Section, StatBand } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { EmptyState, Spinner } from '@/components/feedback';
import { Button } from '@/components/button';
import { DateInput, Label, SearchInput } from '@/components/input';
import { CancelDialog, Dialog, DialogBody, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Drawer, ListRow, RowActions, type RowAction } from '@/components/misc';
import { useDebounce } from '@/hooks/useDebounce';
import { Form, TextField } from '@/components/form';
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
  useUpdatePurchaseNumber,
} from '@/features/purchasing/hooks/usePurchases';
import { PurchaseFormDialog } from '@/features/purchasing/components/PurchaseFormDialog';
import { ExtraCostsEditor } from '@/features/purchasing/components/ExtraCostsEditor';
import { SuppliersDrawer } from '@/features/suppliers/components/SuppliersDrawer';
import { MarkReceivedDialog } from '@/features/purchasing/components/MarkReceivedDialog';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import type { CreditCardDTO } from '@/services/wails-types';
import type { Purchase } from '@/services/purchasing';
import { formatCurrency, formatDate, formatNumber, truncate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const statusMap: Record<string, { variant: 'success' | 'warning' | 'info' | 'destructive' | 'muted'; label: string }> = {
  pending: { variant: 'warning', label: 'Pendiente' },
  received: { variant: 'info', label: 'Recibida' },
  cancelled: { variant: 'destructive', label: 'Anulada' },
};

const paymentMethodLabels: Record<string, string> = {
  card: 'Tarjeta de crédito',
  cash: 'Efectivo',
  digital_wallet: 'Billetera digital',
};

const CANCEL_REASONS = ['Mal estado', 'Error en el ingreso', 'Compra duplicada', 'Cancelado por el proveedor'];

// A capped cell keeps the full value reachable on hover.
function capped(text: string, max: number) {
  if (!text) return '—';
  return <span title={text}>{truncate(text, max)}</span>;
}

const columns: Column<Purchase>[] = [
  {
    id: 'date',
    header: 'Fecha de compra',
    sortable: true,
    cell: (row) => <span className="tabular muted">{formatDate(row.date)}</span>,
  },
  {
    id: 'number',
    header: 'Número',
    sortable: true,
    cell: (row) => <span className="fw-medium tabular">{row.number}</span>,
  },
  {
    id: 'supplierName',
    header: 'Proveedor',
    sortable: true,
    maxChars: 30,
    cell: (row) => capped(row.supplierName, 30),
  },
  {
    id: 'productsText',
    header: 'Productos',
    maxChars: 40,
    cell: (row) => capped(row.productsText, 40),
  },
  {
    id: 'cost',
    header: 'Costo (PEN | USD)',
    sortable: true,
    accessor: (row) => row.realCostPen,
    align: 'numeric',
    cell: (row) => (
      <span className="tabular">
        Costo: {formatCurrency(row.realCostPen)} | {formatCurrency(row.totalCostUsd, 'USD')}
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
          {row.customerId && <Badge variant="info">Cliente</Badge>}
          {row.faulty && <Badge variant="destructive">Defectuoso</Badge>}
          {row.soldOut && <Badge variant="success">Vendido</Badge>}
        </div>
      );
    },
  },
];

const numberSchema = z.object({
  number: z.string().trim().min(1, 'Ingrese el número'),
});

function EditNumberDialog({ open, onOpenChange, purchase }: { open: boolean; onOpenChange: (open: boolean) => void; purchase: Purchase | null }) {
  const push = useNotificationStore((s) => s.push);
  const updateNumber = useUpdatePurchaseNumber();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>Editar número de compra</DialogTitle>
        </DialogHeader>
        <Form<{ number: string }>
          key={purchase?.id ?? 'number'}
          schema={numberSchema}
          defaultValues={{ number: purchase?.number ?? '' }}
          onSubmit={(values) => {
            if (!purchase) return;
            updateNumber.mutate(
              { id: purchase.id, number: values.number },
              {
                onSuccess: (updated) => {
                  push({ title: 'Número actualizado', description: updated.number, variant: 'success' });
                  onOpenChange(false);
                },
                onError: (err: unknown) => {
                  push({ title: 'No se pudo actualizar el número', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
                },
              },
            );
          }}
        >
          <DialogBody>
            <TextField name="number" label="Número de compra" required autoFocus />
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={updateNumber.isPending}>
              Cancelar
            </Button>
            <Button type="submit" loading={updateNumber.isPending}>
              Guardar
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

export function PurchasesPage() {
  const [searchInput, setSearchInput] = useState('');
  const search = useDebounce(searchInput);
  const [statusFilter, setStatusFilter] = useState('all');
  const [creditCardId, setCreditCardId] = useState('');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const { data, isLoading, isError, error, refetch } = usePurchases({ search, creditCardId, from, to });
  const cancel = useCancelPurchase();
  const markReceived = useMarkPurchaseReceived();
  const markFaulty = useMarkPurchaseFaulty();
  const push = useNotificationStore((s) => s.push);

  const [formOpen, setFormOpen] = useState(false);
  const [suppliersOpen, setSuppliersOpen] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Purchase | null>(null);
  const [receivedTarget, setReceivedTarget] = useState<Purchase | null>(null);
  const [detailTarget, setDetailTarget] = useState<{ id: string; number?: string } | null>(null);
  const [numberEditTarget, setNumberEditTarget] = useState<Purchase | null>(null);

  // Sales link here with ?purchase=<id>; clear it when the drawer closes so
  // the back button does not re-open the drawer.
  const [searchParams, setSearchParams] = useSearchParams();
  const requestedPurchaseId = searchParams.get('purchase') ?? '';
  useEffect(() => {
    if (requestedPurchaseId) setDetailTarget({ id: requestedPurchaseId });
  }, [requestedPurchaseId]);

  const closeDetail = () => {
    setDetailTarget(null);
    setSearchParams({}, { replace: true });
  };
  const openDetail = (row: Purchase) => setDetailTarget({ id: row.id, number: row.number });

  const detailQuery = useQuery({
    queryKey: queryKeys.purchasing.detail(detailTarget?.id ?? ''),
    queryFn: () => wailsClient.getPurchase(detailTarget!.id),
    enabled: Boolean(detailTarget),
  });

  const cardsQuery = useQuery({
    queryKey: queryKeys.treasury.creditCards,
    queryFn: () => wailsClient.listCreditCards(),
    enabled: filtersOpen,
  });

  const activeFilterCount = [creditCardId, from, to].filter(Boolean).length;
  const clearFilters = () => {
    setCreditCardId('');
    setFrom('');
    setTo('');
  };

  const purchases = data ?? [];
  const filtered = useMemo(() => {
    if (statusFilter === 'all') return purchases;
    return purchases.filter((p) => p.status === statusFilter);
  }, [purchases, statusFilter]);

  const baseCost = purchases.reduce((s, p) => s + p.costUsd, 0);
  const extraCost = purchases.reduce((s, p) => s + p.extraCostUsd, 0);
  const totalCost = purchases.reduce((s, p) => s + p.totalCostUsd, 0);
  const landedCost = purchases.reduce((s, p) => s + p.realCostPen, 0);
  const pending = purchases.filter((p) => p.status === 'pending' || p.status === 'received').length;
  const cancelled = purchases.filter((p) => p.status === 'cancelled').length;

  const openCreate = () => setFormOpen(true);

  const buildActions = (row: Purchase): RowAction[] => {
    const open = row.status !== 'cancelled';
    const receivable = !row.arrivalDate && !row.faulty && row.status !== 'cancelled';
    const actions: RowAction[] = [
      { label: 'Ver detalle', icon: Eye, onSelect: () => openDetail(row) },
      { label: 'Editar número', icon: Pencil, onSelect: () => setNumberEditTarget(row) },
    ];
    if (receivable) {
      actions.push({
        label: 'Marcar como recibido',
        icon: Download,
        onSelect: () => setReceivedTarget(row),
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
    return actions;
  };

  const tableColumns = useMemo<Column<Purchase>[]>(() => {
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => (
          <div onClick={(e) => e.stopPropagation()}>
            <RowActions actions={buildActions(row)} label={`Acciones de ${row.number}`} />
          </div>
        ),
      },
    ];
  }, []);

  return (
    <PageContainer>
      <PageHeader
        title="Compras"
        actions={
          <div className="hstack hstack--sm">
            <Button variant="outline" onClick={() => setSuppliersOpen(true)}>
              <Users /> Proveedores
            </Button>
            <Button onClick={openCreate}>
              <Plus /> Nueva compra
            </Button>
          </div>
        }
      />

      <StatBand>
        <StatCard label="Compras" value={String(purchases.length)} icon={Package} />
        <StatCard label="Costo base (USD)" value={formatCurrency(baseCost, 'USD')} />
        <StatCard label="Costos extras (USD)" value={formatCurrency(extraCost, 'USD')} />
        <StatCard label="Costo total (USD)" value={formatCurrency(totalCost, 'USD')} />
        <StatCard label="Costo real (PEN)" value={formatCurrency(landedCost)} />
        <StatCard label="Por Pagar" value={String(pending)} />
        <StatCard label="Anuladas" value={String(cancelled)} />
      </StatBand>

      <DataTable
        columns={tableColumns}
        data={filtered}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        onRowClick={(row) => openDetail(row)}
        rowClassName={(row) => (row.soldOut ? 'row-sold-out' : undefined)}
        rowActions={buildActions}
        preferencesKey="purchases-journal"
        defaultPreferences={{ sort: { id: 'date', direction: 'desc' } }}
        toolbarLeft={
          <>
            <SearchInput
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onClear={() => setSearchInput('')}
              placeholder="Número o proveedor…"
              className="datatable-search"
              aria-label="Buscar compra"
            />
            <Select
              items={[
                { value: 'all', label: 'Todos' },
                { value: 'pending', label: 'Pendientes' },
                { value: 'received', label: 'Recibidas' },
                { value: 'cancelled', label: 'Anuladas' },
              ]}
              value={statusFilter}
              onValueChange={(v) => setStatusFilter(v ?? 'all')}
            >
              <SelectTrigger style={{ width: '11rem' }} aria-label="Filtrar por estado">
                <SelectValue placeholder="Todos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Todos</SelectItem>
                <SelectItem value="pending">Pendientes</SelectItem>
                <SelectItem value="received">Recibidas</SelectItem>
                <SelectItem value="cancelled">Anuladas</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        toolbarRight={
          <Button variant="outline" onClick={() => setFiltersOpen(true)}>
            <Filter /> Filtros{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
          </Button>
        }
        empty={
          <EmptyState
            title="No hay compras"
            description="Crea tu primera compra para abastecer el inventario."
            action={{ label: 'Nueva compra', onClick: openCreate }}
          />
        }
      />

      <Drawer
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        title="Filtros avanzados"
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
            <Label>Rango de fechas</Label>
            <div className="hstack hstack--sm">
              <DateInput value={from} onChange={(e) => setFrom(e.target.value)} aria-label="Desde" placeholder="DD/MM/AAAA" />
              <DateInput value={to} onChange={(e) => setTo(e.target.value)} aria-label="Hasta" placeholder="DD/MM/AAAA" />
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
      <SuppliersDrawer open={suppliersOpen} onOpenChange={setSuppliersOpen} />
      <EditNumberDialog
        open={!!numberEditTarget}
        onOpenChange={(open) => {
          if (!open) setNumberEditTarget(null);
        }}
        purchase={numberEditTarget}
      />

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
                push({ title: 'Compra marcada como recibida', variant: 'success' });
                setReceivedTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo marcar la compra',
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
        title="Anular compra"
        confirmLabel="Anular"
        reasons={CANCEL_REASONS}
        loading={cancel.isPending || markFaulty.isPending}
        onConfirm={(reason, preset) => {
          if (!cancelTarget) return;
          const onSuccess = () => {
            push({ title: 'Compra anulada', variant: 'success' });
            setCancelTarget(null);
          };
          const onError = (err: unknown) => {
            push({
              title: 'No se pudo anular la compra',
              description: err instanceof Error ? err.message : undefined,
              variant: 'destructive',
            });
            setCancelTarget(null);
          };
          if (preset === 'Mal estado') {
            markFaulty.mutate({ id: cancelTarget.id, reason }, { onSuccess, onError });
          } else {
            cancel.mutate({ id: cancelTarget.id, reason }, { onSuccess, onError });
          }
        }}
      />

      <Drawer
        open={!!detailTarget}
        onOpenChange={(open) => {
          if (!open) closeDetail();
        }}
        title="Detalle de compra"
        description={detailTarget?.number || detailQuery.data?.number}
      >
        {detailTarget &&
          (detailQuery.isLoading ? (
            <Spinner />
          ) : detailQuery.isError ? (
            <EmptyState title="No se pudo cargar" description="No se pudo cargar el detalle de la compra." />
          ) : detailQuery.data ? (
            <div className="stack">
              <div className="doc-summary">
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Tipo</div>
                  <div className="doc-summary__amount">
                    {detailQuery.data.customerId ? 'Cliente a pedido' : 'General (stock)'}
                  </div>
                </div>
                {detailQuery.data.customerId && (
                  <div className="doc-summary__row">
                    <div className="doc-summary__meta">Cliente</div>
                    <div className="doc-summary__amount">
                      {detailQuery.data.customerName || '—'}
                    </div>
                  </div>
                )}
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Proveedor</div>
                  <div className="doc-summary__amount">{detailQuery.data.supplierName || '—'}</div>
                </div>
                {detailQuery.data.saleNumber && (
                  <div className="doc-summary__row">
                    <div className="doc-summary__meta">Venta de origen</div>
                    <div className="doc-summary__amount tabular">
                      {detailQuery.data.saleNumber}
                    </div>
                  </div>
                )}
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Forma de pago</div>
                  <div className="doc-summary__amount">
                    {paymentMethodLabels[detailQuery.data.paymentMethod] ?? detailQuery.data.paymentMethod}
                  </div>
                </div>
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Fecha de compra</div>
                  <div className="doc-summary__amount">{formatDate(detailQuery.data.date)}</div>
                </div>
                {detailQuery.data.expectedDate && (
                  <div className="doc-summary__row">
                    <div className="doc-summary__meta">Fecha estimada</div>
                    <div className="doc-summary__amount">{formatDate(detailQuery.data.expectedDate)}</div>
                  </div>
                )}
                {detailQuery.data.receivedDate && (
                  <div className="doc-summary__row">
                    <div className="doc-summary__meta">Recepción</div>
                    <div className="doc-summary__amount">{formatDate(detailQuery.data.receivedDate)}</div>
                  </div>
                )}
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Tipo de cambio</div>
                  <div className="doc-summary__amount tabular">{formatNumber(detailQuery.data.exchangeRate, 4)}</div>
                </div>
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Costo base (USD)</div>
                  <div className="doc-summary__amount">
                    {formatCurrency(detailQuery.data.costUsd, 'USD')}
                  </div>
                </div>
                <div className="doc-summary__row">
                  <div className="doc-summary__meta">Costos extras (USD)</div>
                  <div className="doc-summary__amount">
                    {formatCurrency(detailQuery.data.extraCostUsd, 'USD')}
                  </div>
                </div>
                <div className="doc-summary__row doc-summary__row--total">
                  <div className="doc-summary__meta">Costo total (USD)</div>
                  <div className="doc-summary__amount">
                    {formatCurrency(detailQuery.data.totalCostUsd, 'USD')}
                  </div>
                </div>
                <div className="doc-summary__row doc-summary__row--total">
                  <div className="doc-summary__meta">Costo real (PEN)</div>
                  <div className="doc-summary__amount">{formatCurrency(detailQuery.data.realCostPen)}</div>
                </div>
                {detailQuery.data.refundAmount > 0 && (
                  <div className="doc-summary__row">
                    <div className="doc-summary__meta">Reintegro (USD)</div>
                    <div className="doc-summary__amount">{formatCurrency(detailQuery.data.refundAmount, 'USD')}</div>
                  </div>
                )}
              </div>
              {detailQuery.data.notes && <p className="muted">{detailQuery.data.notes}</p>}
              {detailQuery.data.faultyReason && <p className="field-hint">Defectuoso: {detailQuery.data.faultyReason}</p>}
              {detailQuery.data.cancelledReason && (
                <p className="field-hint">Anulada: {detailQuery.data.cancelledReason}</p>
              )}
              <div className="stack">
                {detailQuery.data.items.map((it) => {
                  const soldOut = it.quantitySold >= it.quantity;
                  return (
                    <ListRow
                      key={it.id}
                      className={soldOut ? 'list-row--sold' : undefined}
                      title={it.description}
                      meta={
                        <>
                          {formatNumber(it.quantity)} × {formatCurrency(it.unitCostUsd, 'USD')}
                          {it.salePricePen > 0 ? ` · Venta ${formatCurrency(it.salePricePen)}` : ''}
                          {` · Vendidas ${formatNumber(it.quantitySold)}/${formatNumber(it.quantity)}`}
                        </>
                      }
                      trailing={
                        <>
                          {soldOut && <Badge variant="muted">Agotado</Badge>}
                          <span className="tabular">{formatCurrency(it.lineTotalUsd, 'USD')}</span>
                        </>
                      }
                    />
                  );
                })}
              </div>
              <Section title="Costos extras" description="Gastos adicionales de la compra: fletes, aranceles, manejo.">
                <ExtraCostsEditor
                  purchaseId={detailQuery.data.id}
                  disabled={detailQuery.data.status === 'cancelled'}
                  defaultRate={detailQuery.data.exchangeRate}
                />
              </Section>
            </div>
          ) : null)}
      </Drawer>
    </PageContainer>
  );
}
