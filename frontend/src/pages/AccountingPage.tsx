import { useMemo, useState } from 'react';
import { PageContainer, PageHeader, Grid, Section } from '@/components/layout';
import { StatCard } from '@/components/card';
import { DataTable, type Column } from '@/components/table';
import { Badge } from '@/components/badge';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';
import {
  useChartOfAccounts,
  useDeleteChartOfAccount,
  useJournalEntries,
  usePostJournalEntry,
} from '@/features/accounting/hooks/useAccounting';
import { ChartAccountFormDialog } from '@/features/accounting/components/ChartAccountFormDialog';
import { JournalEntryFormDialog } from '@/features/accounting/components/JournalEntryFormDialog';
import type { Account, AccountType, JournalEntry, JournalEntryStatus } from '@/services/accounting';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const AccountTypeBadge = ({ type }: { type: AccountType }) => {
  switch (type) {
    case 'asset':
      return <Badge variant="info">Activo</Badge>;
    case 'liability':
      return <Badge variant="warning">Pasivo</Badge>;
    case 'equity':
      return <Badge variant="muted">Patrimonio</Badge>;
    case 'income':
      return <Badge variant="success">Ingreso</Badge>;
    default:
      return <Badge variant="destructive">Gasto</Badge>;
  }
};

const EntryStatusBadge = ({ status }: { status: JournalEntryStatus }) => {
  switch (status) {
    case 'posted':
      return <Badge variant="success">Publicado</Badge>;
    case 'reversed':
      return <Badge variant="muted">Revertido</Badge>;
    default:
      return <Badge variant="warning">Borrador</Badge>;
  }
};

function entryTotals(entry: JournalEntry) {
  return entry.lines.reduce(
    (acc, l) => ({ debit: acc.debit + l.debit, credit: acc.credit + l.credit }),
    { debit: 0, credit: 0 },
  );
}

export function AccountingPage() {
  const [accountFormOpen, setAccountFormOpen] = useState(false);
  const [editing, setEditing] = useState<Account | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Account | null>(null);
  const [entryFormOpen, setEntryFormOpen] = useState(false);

  const {
    data: accounts = [],
    isLoading: accountsLoading,
    isError: accountsError,
    error: accountsErrorObj,
    refetch: refetchAccounts,
  } = useChartOfAccounts();
  const entriesQuery = useJournalEntries();
  const deleteMutation = useDeleteChartOfAccount();
  const postMutation = usePostJournalEntry();
  const push = useNotificationStore((s) => s.push);

  const canEdit = usePermission(Permissions.Accounting.Edit);
  const canDelete = usePermission(Permissions.Accounting.Delete);

  const entries = entriesQuery.data ?? [];
  const movementAccounts = accounts.filter((a) => a.allowsMovement).length;
  const postedCount = entries.filter((e) => e.status === 'posted').length;

  const accountColumns = useMemo<Column<Account>[]>(() => {
    const base: Column<Account>[] = [
      {
        id: 'code',
        header: 'Código',
        sortable: true,
        sticky: true,
        cell: (row) => <span className="tabular-nums font-medium">{row.code}</span>,
      },
      {
        id: 'name',
        header: 'Nombre',
        cell: (row) => (
          <div className="flex flex-col">
            <span>{row.name}</span>
            {row.description && <span className="text-xs text-muted-foreground">{row.description}</span>}
          </div>
        ),
      },
      { id: 'type', header: 'Tipo', cell: (row) => <AccountTypeBadge type={row.type} /> },
      {
        id: 'parent',
        header: 'Cuenta padre',
        cell: (row) => {
          const parent = row.parentId ? accounts.find((a) => a.id === row.parentId) : undefined;
          return parent ? `${parent.code} — ${parent.name}` : '—';
        },
      },
      {
        id: 'allowsMovement',
        header: 'Movimiento',
        cell: (row) => (row.allowsMovement ? <Badge variant="success">Sí</Badge> : <Badge variant="muted">No</Badge>),
      },
      {
        id: 'isActive',
        header: 'Estado',
        cell: (row) => (row.isActive ? <Badge variant="success">Activa</Badge> : <Badge variant="muted">Inactiva</Badge>),
      },
    ];
    if (!canEdit && !canDelete) return base;
    return [
      ...base,
      {
        id: 'actions',
        header: '',
        width: 88,
        exportable: false,
        cell: (row) => (
          <div className="flex items-center justify-end gap-1">
            {canEdit && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Editar ${row.name}`}
                onClick={() => {
                  setEditing(row);
                  setAccountFormOpen(true);
                }}
              >
                <Icons.Action.Edit />
              </Button>
            )}
            {canDelete && !accounts.some((a) => row.path && a.path.startsWith(`${row.path}.`)) && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Eliminar ${row.name}`}
                onClick={() => setDeleteTarget(row)}
              >
                <Icons.Action.Delete />
              </Button>
            )}
          </div>
        ),
      },
    ];
  }, [accounts, canEdit, canDelete]);

  const entryColumns = useMemo<Column<JournalEntry>[]>(() => {
    const base: Column<JournalEntry>[] = [
      {
        id: 'number',
        header: 'N.º',
        sortable: true,
        sticky: true,
        cell: (row) => <span className="tabular-nums font-medium">{row.number}</span>,
      },
      {
        id: 'entryDate',
        header: 'Fecha',
        sortable: true,
        cell: (row) => <span className="tabular-nums">{formatDate(row.entryDate)}</span>,
      },
      { id: 'description', header: 'Descripción', cell: (row) => row.description },
      { id: 'status', header: 'Estado', cell: (row) => <EntryStatusBadge status={row.status} /> },
      {
        id: 'debit',
        header: 'Débito',
        align: 'right',
        cell: (row) => <span className="tabular-nums">{formatCurrency(entryTotals(row).debit, 'PEN')}</span>,
      },
      {
        id: 'credit',
        header: 'Crédito',
        align: 'right',
        cell: (row) => <span className="tabular-nums">{formatCurrency(entryTotals(row).credit, 'PEN')}</span>,
      },
      {
        id: 'lines',
        header: 'Líneas',
        cell: (row) => <span className="tabular-nums">{row.lines.length}</span>,
      },
    ];
    if (!canEdit) return base;
    return [
      ...base,
      {
        id: 'actions',
        header: '',
        width: 110,
        exportable: false,
        cell: (row) =>
          row.status === 'draft' ? (
            <div className="flex justify-end">
              <Button
                variant="ghost"
                size="sm"
                loading={postMutation.isPending}
                onClick={() => {
                  postMutation.mutate(row.id, {
                    onSuccess: () => push({ title: 'Asiento publicado', variant: 'success' }),
                    onError: (err: unknown) =>
                      push({ title: 'No se pudo publicar', description: err instanceof Error ? err.message : undefined, variant: 'destructive' }),
                  });
                }}
              >
                Publicar
              </Button>
            </div>
          ) : null,
      },
    ];
  }, [canEdit, postMutation, push]);

  return (
    <PageContainer>
      <PageHeader
        title="Contabilidad"
        subtitle="Plan de cuentas, libro diario y asientos contables"
        actions={
          <>
            <Can permission={Permissions.Accounting.Create}>
              <Button variant="outline" onClick={() => setEntryFormOpen(true)}>
                <Icons.Action.Create /> Nuevo asiento
              </Button>
              <Button
                onClick={() => {
                  setEditing(null);
                  setAccountFormOpen(true);
                }}
              >
                <Icons.Action.Create /> Nueva cuenta
              </Button>
            </Can>
          </>
        }
      />

      <Grid cols={4}>
        <StatCard label="Total de cuentas" value={String(accounts.length)} icon={Icons.Navigation.Accounting} />
        <StatCard label="Cuentas con movimiento" value={String(movementAccounts)} />
        <StatCard label="Asientos registrados" value={String(entries.length)} />
        <StatCard label="Asientos publicados" value={String(postedCount)} />
      </Grid>

      <Section title="Plan de cuentas" description="Estructura jerárquica de cuentas contables de la empresa.">
        <DataTable
          columns={accountColumns}
          data={accounts}
          keyField="id"
          loading={accountsLoading}
          error={accountsError ? (accountsErrorObj as Error) : null}
          onRetry={() => refetchAccounts()}
          globalSearch
          exportFilename="plan-de-cuentas.csv"
          empty={
            <EmptyState
              title="No hay cuentas contables"
              description="Crea la primera cuenta del plan contable para empezar a registrar asientos."
            />
          }
        />
      </Section>

      <Section title="Libro diario" description="Asientos contables registrados en orden cronológico.">
        <DataTable
          columns={entryColumns}
          data={entries}
          keyField="id"
          loading={entriesQuery.isLoading}
          error={entriesQuery.isError ? (entriesQuery.error as Error) : null}
          onRetry={() => entriesQuery.refetch()}
          globalSearch
          exportFilename="libro-diario.csv"
          empty={
            <EmptyState
              title="No hay asientos"
              description="Registra un asiento manualmente o genera uno desde ventas, compras y tesorería."
            />
          }
        />
      </Section>

      <ChartAccountFormDialog open={accountFormOpen} onOpenChange={setAccountFormOpen} account={editing} />

      <JournalEntryFormDialog open={entryFormOpen} onOpenChange={setEntryFormOpen} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar cuenta contable"
        description={`¿Eliminar la cuenta ${deleteTarget?.code ?? ''} — ${deleteTarget?.name ?? ''}? La cuenta quedará desactivada. Esta acción no se puede deshacer.`}
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Cuenta eliminada', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({ title: 'No se pudo eliminar la cuenta', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </PageContainer>
  );
}
