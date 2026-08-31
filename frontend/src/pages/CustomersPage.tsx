import { useMemo, useState } from 'react';
import { Users, Pencil, Trash2, Plus } from 'lucide-react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { CustomerStatusBadge } from '@/components/badge';
import { SearchInput } from '@/components/input';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useCustomers, useDeleteCustomer } from '@/features/customers';
import { CustomerFormDialog } from '@/features/customers/components/CustomerFormDialog';
import type { Customer } from '@/types/domain';
import { formatCurrency } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const columns: Column<Customer>[] = [
  {
    id: 'businessName',
    header: 'Cliente',
    sortable: true,
    sticky: true,
    cell: (row) => (
      <div className="vstack">
        <span className="fw-medium">{row.businessName}</span>
        {row.contactName && <span className="subtle-text">{row.contactName}</span>}
      </div>
    ),
  },
  {
    id: 'documentNumber',
    header: 'Documento',
    cell: (row) => (
      <span className="muted">
        {row.documentType} · {row.documentNumber}
      </span>
    ),
  },
  { id: 'phone', header: 'Teléfono', cell: (row) => row.phone || '—' },
  { id: 'email', header: 'Correo', cell: (row) => row.email || '—' },
  {
    id: 'creditLimit',
    header: 'Límite de crédito',
    align: 'right',
    sortable: true,
    cell: (row) => <span className="tabular">{formatCurrency(row.creditLimit)}</span>,
  },
  {
    id: 'currentDebt',
    header: 'Deuda',
    align: 'right',
    sortable: true,
    cell: (row) => (
      <span className={row.currentDebt > 0 ? 'fw-medium tabular' : 'tabular muted'}>
        {formatCurrency(row.currentDebt)}
      </span>
    ),
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (row) => <CustomerStatusBadge status={row.status} />,
  },
];

export function CustomersPage() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Customer | null>(null);

  const { data, isLoading, isError, error, refetch } = useCustomers({ search, status: (status || undefined) as Customer['status'] | undefined });
  const deleteMutation = useDeleteCustomer();
  const push = useNotificationStore((s) => s.push);

  const customers = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalDebt = customers.reduce((s, c) => s + c.currentDebt, 0);
  const active = customers.filter((c) => c.status === 'active').length;
  const withDebt = customers.filter((c) => c.currentDebt > 0).length;

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const tableColumns = useMemo<Column<Customer>[]>(() => [
    ...columns,
    {
      id: 'actions',
      header: '',
      width: 72,
      cell: (row) => (
        <RowActions
          actions={[
            {
              label: 'Editar',
              icon: Pencil,
              onSelect: () => {
                setEditing(row);
                setFormOpen(true);
              },
            },
            {
              label: 'Eliminar',
              icon: Trash2,
              danger: true,
              onSelect: () => setDeleteTarget(row),
            },
          ]}
          label={`Acciones de ${row.businessName}`}
        />
      ),
    },
  ], []);

  return (
    <PageContainer>
      <PageHeader
        title="Clientes"
        subtitle="Gestión de clientes y cuentas por cobrar"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nuevo cliente
          </Button>
        }
      />

      <Grid cols={4}>
        <StatCard label="Total clientes" value={String(total)} icon={Users} />
        <StatCard label="Clientes activos" value={String(active)} />
        <StatCard label="Con deuda" value={String(withDebt)} />
        <StatCard label="Deuda total" value={formatCurrency(totalDebt)} />
      </Grid>

      <DataTable
        columns={tableColumns}
        data={customers}
        keyField="id"
        loading={isLoading}
        error={isError ? (error as Error) : null}
        onRetry={() => refetch()}
        preferencesKey="customers"
        toolbarLeft={
          <>
            <SearchInput
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onClear={() => setSearch('')}
              placeholder="Buscar cliente…"
              className="datatable-search"
              aria-label="Buscar cliente"
            />
            <Select
              items={[
                { value: 'all', label: 'Estado: todos' },
                { value: 'active', label: 'Activos' },
                { value: 'inactive', label: 'Inactivos' },
                { value: 'blocked', label: 'Bloqueados' },
              ]}
              value={status || 'all'}
              onValueChange={(v) => setStatus(v === 'all' ? '' : (v ?? ''))}
            >
              <SelectTrigger style={{ width: '11rem' }} aria-label="Filtrar por estado">
                <SelectValue placeholder="Estado: todos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Estado: todos</SelectItem>
                <SelectItem value="active">Activos</SelectItem>
                <SelectItem value="inactive">Inactivos</SelectItem>
                <SelectItem value="blocked">Bloqueados</SelectItem>
              </SelectContent>
            </Select>
          </>
        }
        empty={
          <EmptyState
            title="No hay clientes registrados"
            description="Registra tu primer cliente para comenzar a gestionar ventas y cuentas por cobrar."
            action={{ label: 'Nuevo cliente', onClick: openCreate }}
          />
        }
      />

      <CustomerFormDialog open={formOpen} onOpenChange={setFormOpen} customer={editing} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar cliente"
        description={`¿Eliminar a ${deleteTarget?.businessName ?? 'este cliente'}? Esta acción no se puede deshacer.`}
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Cliente eliminado', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({ title: 'No se pudo eliminar el cliente', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </PageContainer>
  );
}
