import { useMemo, useState } from 'react';
import { PageContainer, PageHeader, Grid } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { CustomerStatusBadge } from '@/components/badge';
import { Input } from '@/components/input';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { useCustomers, useDeleteCustomer } from '@/features/customers';
import { CustomerFormDialog } from '@/features/customers/components/CustomerFormDialog';
import type { Customer, CustomerStatus } from '@/types/domain';
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
  const [status, setStatus] = useState<CustomerStatus | ''>('');
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Customer | null>(null);

  const { data, isLoading, isError, error, refetch } = useCustomers({ search, status: status || undefined });
  const deleteMutation = useDeleteCustomer();
  const push = useNotificationStore((s) => s.push);
  const canEdit = usePermission(Permissions.Customers.Edit);
  const canDelete = usePermission(Permissions.Customers.Delete);

  const customers = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalDebt = customers.reduce((s, c) => s + c.currentDebt, 0);
  const active = customers.filter((c) => c.status === 'active').length;
  const withDebt = customers.filter((c) => c.currentDebt > 0).length;

  const tableColumns = useMemo<Column<Customer>[]>(() => {
    if (!canEdit && !canDelete) return columns;
    return [
      ...columns,
      {
        id: 'actions',
        header: '',
        width: 88,
        exportable: false,
        cell: (row) => (
          <div className="row-actions">
            {canEdit && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Editar ${row.businessName}`}
                onClick={() => {
                  setEditing(row);
                  setFormOpen(true);
                }}
              >
                <Icons.Action.Edit />
              </Button>
            )}
            {canDelete && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Eliminar ${row.businessName}`}
                onClick={() => setDeleteTarget(row)}
              >
                <Icons.Action.Delete />
              </Button>
            )}
          </div>
        ),
      },
    ];
  }, [canEdit, canDelete]);

  return (
    <PageContainer>
      <PageHeader
        title="Clientes"
        subtitle="Gestión de clientes y cuentas por cobrar"
        actions={
          <div className="hstack hstack--sm">
            <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Buscar cliente…" style={{ width: "16rem" }} aria-label="Buscar cliente" />
            <Can permission={Permissions.Customers.Create}>
              <Button
                onClick={() => {
                  setEditing(null);
                  setFormOpen(true);
                }}
              >
                <Icons.Action.Create /> Nuevo cliente
              </Button>
            </Can>
          </div>
        }
      />

      <Grid cols={4}>
        <StatCard label="Total clientes" value={String(total)} icon={Icons.Navigation.Customers} />
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
        globalSearch={false}
        exportFilename="clientes.csv"
        toolbarLeft={
          <Select value={status} onValueChange={(v) => setStatus((v === 'all' ? '' : v) as CustomerStatus | '')}>
            <SelectTrigger style={{ width: "11rem" }} aria-label="Filtrar por estado">
              <SelectValue placeholder="Estado: todos" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Estado: todos</SelectItem>
              <SelectItem value="active">Activos</SelectItem>
              <SelectItem value="inactive">Inactivos</SelectItem>
              <SelectItem value="blocked">Bloqueados</SelectItem>
            </SelectContent>
          </Select>
        }
        empty={
          <EmptyState
            title="No hay clientes registrados"
            description="Cuando se registren clientes, aparecerán aquí con su estado de cuenta."
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
