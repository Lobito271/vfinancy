import { useMemo, useState } from 'react';
import { Plus, Pencil, Trash2, Copy } from 'lucide-react';
import { PageContainer, PageHeader, Section } from '@/components/layout';
import { Button } from '@/components/button';
import { EmptyState, Spinner } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { Badge } from '@/components/badge';
import { DataTable, type Column } from '@/components/table';
import { RowActions, type RowAction } from '@/components/misc';
import { SearchInput } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import type { ShipmentDTO } from '@/services/wails-types';
import { useShipments, useDeleteShipment } from '@/features/shipments/hooks/useShipments';
import { ShipmentFormDialog } from '@/features/shipments/components/ShipmentFormDialog';
import { SHIPMENT_STATUSES } from '@/services/shipments';
import { useCustomers } from '@/features/customers/hooks/useCustomers';
import { formatDate, truncate } from '@/utils/format';
import { copyText } from '@/utils/clipboard';
import { useNotificationStore } from '@/stores/notification';

const STATUS_VARIANT: Record<ShipmentDTO['status'], 'secondary' | 'info' | 'success'> = {
  pending: 'secondary',
  shipped: 'info',
  delivered: 'success',
};

export function ShipmentsPage() {
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | ShipmentDTO['status']>('all');
  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ShipmentDTO | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ShipmentDTO | null>(null);

  const { data: shipments = [], isLoading } = useShipments(search ? { search } : {});
  const { data: customersPage } = useCustomers();
  const customers = customersPage?.items ?? [];
  const deleteMutation = useDeleteShipment();
  const push = useNotificationStore((s) => s.push);

  const visibleShipments = useMemo(
    () => (statusFilter === 'all' ? shipments : shipments.filter((s) => s.status === statusFilter)),
    [shipments, statusFilter],
  );

  const nameById = useMemo(() => new Map(customers.map((c) => [c.id, c.businessName])), [customers]);

  const buildActions = (row: ShipmentDTO): RowAction[] => [
    { label: 'Editar', icon: Pencil, onSelect: () => { setEditTarget(row); setFormOpen(true); } },
    { label: 'Eliminar', icon: Trash2, danger: true, onSelect: () => setDeleteTarget(row) },
  ];

  const columns = useMemo<Column<ShipmentDTO>[]>(
    () => [
      {
        id: 'createdAt',
        header: 'Fecha',
        sortable: true,
        cell: (row) => <span className="tabular muted">{formatDate(row.createdAt)}</span>,
      },
      {
        id: 'code',
        header: 'Código',
        sortable: true,
        cell: (row) => (
          <span className="hstack" style={{ gap: '0.5rem' }}>
            <span className="fw-medium tabular mono">{row.code}</span>
            <Button variant="ghost" size="sm" aria-label={`Copiar código ${row.code}`} onClick={() => void copyText(row.code).then(() => {
              push({ title: `Código ${row.code} copiado`, variant: 'success' });
            })}>
              <Copy size={14} />
            </Button>
          </span>
        ),
      },
      {
        id: 'securityCode',
        header: 'Clave',
        cell: (row) => (
          <span className="hstack" style={{ gap: '0.5rem' }}>
            <span className="fw-medium tabular mono">{row.securityCode}</span>
            <Button variant="ghost" size="sm" aria-label={`Copiar clave ${row.securityCode}`} onClick={() => void copyText(row.securityCode).then(() => {
              push({ title: `Clave ${row.securityCode} copiada`, variant: 'success' });
            })}>
              <Copy size={14} />
            </Button>
          </span>
        ),
      },
      {
        id: 'sale',
        header: 'Venta',
        cell: (row) => (row.saleNumber ? <span className="mono">{row.saleNumber}</span> : '—'),
      },
      {
        id: 'customer',
        header: 'Cliente',
        sortable: true,
        maxChars: 30,
        cell: (row) => {
          const name = nameById.get(row.customerId);
          return name ? <span title={name}>{truncate(name, 30)}</span> : '—';
        },
      },
      {
        id: 'location',
        header: 'Ubicación',
        maxChars: 24,
        cell: (row) =>
          row.location ? (
            <span className="muted" title={row.location}>
              {truncate(row.location, 24)}
            </span>
          ) : (
            '—'
          ),
      },
      {
        id: 'shipmentDate',
        header: 'Envío',
        cell: (row) => <span className="tabular">{row.shipmentDate ? formatDate(row.shipmentDate) : '—'}</span>,
      },
      {
        id: 'deliveredAt',
        header: 'Entrega',
        cell: (row) => <span className="tabular">{row.deliveredAt ? formatDate(row.deliveredAt) : '—'}</span>,
      },
      {
        id: 'description',
        header: 'Descripción',
        maxChars: 40,
        cell: (row) =>
          row.description ? (
            <span className="muted" title={row.description}>
              {truncate(row.description, 40)}
            </span>
          ) : (
            '—'
          ),
      },
      {
        id: 'status',
        header: 'Estado',
        cell: (row) => (
          <Badge variant={STATUS_VARIANT[row.status]}>{SHIPMENT_STATUSES.find((s) => s.value === row.status)?.label}</Badge>
        ),
      },
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => (
          <div onClick={(e) => e.stopPropagation()}>
            <RowActions label={`Acciones de envío #${row.code}`} actions={buildActions(row)} />
          </div>
        ),
      },
    ],
    [nameById, push],
  );

  function openCreate() {
    setEditTarget(null);
    setFormOpen(true);
  }

  return (
    <PageContainer>
      <PageHeader
        title="Envíos"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nuevo envío
          </Button>
        }
      />

      <Section title="Envíos" flat>
        {isLoading ? (
          <div className="page-loader">
            <Spinner />
          </div>
        ) : shipments.length === 0 ? (
          <EmptyState
            title="No hay envíos"
            description="Registra un envío para generar su código de seguimiento."
            action={{ label: 'Nuevo envío', onClick: openCreate }}
          />
        ) : (
          <DataTable
            columns={columns}
            data={visibleShipments}
            keyField="id"
            onRowClick={(row) => {
              setEditTarget(row);
              setFormOpen(true);
            }}
            rowActions={buildActions}
            defaultPreferences={{ sort: { id: 'createdAt', direction: 'desc' } }}
            toolbarLeft={
              <SearchInput
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                onClear={() => setSearch('')}
                placeholder="Buscar por código, clave o descripción…"
                className="datatable-search"
                aria-label="Buscar envío"
              />
            }
            toolbarRight={
              <Select
                items={[{ value: 'all', label: 'Todos' }, ...SHIPMENT_STATUSES.map((s) => ({ value: s.value, label: s.label }))]}
                value={statusFilter}
                onValueChange={(v) => setStatusFilter((v ?? 'all') as 'all' | ShipmentDTO['status'])}
              >
                <SelectTrigger style={{ width: '11rem' }} aria-label="Filtrar por estado">
                  <SelectValue placeholder="Todos" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Todos</SelectItem>
                  {SHIPMENT_STATUSES.map((s) => (
                    <SelectItem key={s.value} value={s.value}>
                      {s.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            }
          />
        )}
      </Section>

      <ShipmentFormDialog open={formOpen} onOpenChange={setFormOpen} edit={editTarget} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar envío"
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Envío eliminado', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({
                title: 'No se pudo eliminar',
                description: err instanceof Error ? err.message : undefined,
                variant: 'destructive',
              });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </PageContainer>
  );
}