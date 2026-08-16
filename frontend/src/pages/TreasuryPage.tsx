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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import {
  useBankAccounts,
  useBankTransactions,
  useDeleteBankAccount,
  useReconcileBankTransaction,
} from '@/features/treasury/hooks/useTreasury';
import { BankAccountFormDialog } from '@/features/treasury/components/BankAccountFormDialog';
import { BankTransactionFormDialog } from '@/features/treasury/components/BankTransactionFormDialog';
import type { BankAccount, BankTransaction, BankTxType } from '@/services/treasury';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

const TxTypeBadge = ({ type }: { type: BankTxType }) => {
  switch (type) {
    case 'deposit':
      return <Badge variant="success">Depósito</Badge>;
    case 'withdrawal':
      return <Badge variant="destructive">Retiro</Badge>;
    case 'fee':
      return <Badge variant="warning">Comisión</Badge>;
    case 'interest':
      return <Badge variant="info">Interés</Badge>;
    case 'transfer':
      return <Badge variant="info">Transferencia</Badge>;
    default:
      return <Badge variant="muted">Otro</Badge>;
  }
};

function signedAmount(tx: BankTransaction): number {
  if (tx.type === 'withdrawal' || tx.type === 'fee' || tx.type === 'transfer') {
    return -tx.amount;
  }
  return tx.amount;
}

export function TreasuryPage() {
  const [accountFormOpen, setAccountFormOpen] = useState(false);
  const [editing, setEditing] = useState<BankAccount | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<BankAccount | null>(null);
  const [txFormOpen, setTxFormOpen] = useState(false);
  const [txAccountId, setTxAccountId] = useState('');
  const [txFilter, setTxFilter] = useState('');

  const { data: accounts = [], isLoading: accountsLoading, isError: accountsError, error: accountsErrorObj, refetch: refetchAccounts } = useBankAccounts();
  const transactionsQuery = useBankTransactions(txFilter || undefined);
  const deleteMutation = useDeleteBankAccount();
  const reconcileMutation = useReconcileBankTransaction();
  const push = useNotificationStore((s) => s.push);

  const canEdit = usePermission(Permissions.Treasury.Edit);
  const canDelete = usePermission(Permissions.Treasury.Delete);
  const canConciliate = usePermission(Permissions.Treasury.Conciliate);

  const totalBalancePen = accounts.filter((a) => a.currency === 'PEN').reduce((s, a) => s + a.balance, 0);
  const totalBalanceUsd = accounts.filter((a) => a.currency === 'USD').reduce((s, a) => s + a.balance, 0);
  const activeCount = accounts.filter((a) => a.isActive).length;

  const currencyOf = useMemo(() => {
    const map = new Map<string, string>();
    for (const a of accounts) map.set(a.id, a.currency);
    return map;
  }, [accounts]);

  const accountColumns = useMemo<Column<BankAccount>[]>(() => {
    const base: Column<BankAccount>[] = [
      {
        id: 'bank',
        header: 'Banco',
        sortable: true,
        sticky: true,
        cell: (row) => <span className="font-medium">{row.bank}</span>,
      },
      {
        id: 'accountNumber',
        header: 'N.º de cuenta',
        cell: (row) => <span className="tabular-nums">{row.accountNumber}</span>,
      },
      {
        id: 'accountType',
        header: 'Tipo',
        cell: (row) => (row.accountType === 'checking' ? 'Cuenta corriente' : 'Cuenta de ahorros'),
      },
      { id: 'currency', header: 'Moneda', cell: (row) => <span className="font-medium">{row.currency}</span> },
      {
        id: 'balance',
        header: 'Saldo',
        align: 'right',
        sortable: true,
        cell: (row) => <span className="tabular-nums">{formatCurrency(row.balance, row.currency)}</span>,
      },
      {
        id: 'isDefault',
        header: 'Principal',
        cell: (row) => (row.isDefault ? <Badge variant="info">Principal</Badge> : '—'),
      },
      {
        id: 'isActive',
        header: 'Estado',
        cell: (row) =>
          row.isActive ? <Badge variant="success">Activa</Badge> : <Badge variant="muted">Inactiva</Badge>,
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
                aria-label={`Editar ${row.bank}`}
                onClick={() => {
                  setEditing(row);
                  setAccountFormOpen(true);
                }}
              >
                <Icons.Action.Edit />
              </Button>
            )}
            {canDelete && (
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Eliminar ${row.bank}`}
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

  const txColumns = useMemo<Column<BankTransaction>[]>(() => {
    const base: Column<BankTransaction>[] = [
      {
        id: 'date',
        header: 'Fecha',
        sortable: true,
        sticky: true,
        cell: (row) => <span className="tabular-nums">{formatDate(row.date)}</span>,
      },
      {
        id: 'description',
        header: 'Descripción',
        cell: (row) => (
          <div className="flex flex-col">
            <span>{row.description}</span>
            {row.reference && <span className="text-xs text-muted-foreground">Ref. {row.reference}</span>}
          </div>
        ),
      },
      { id: 'type', header: 'Tipo', cell: (row) => <TxTypeBadge type={row.type} /> },
      {
        id: 'amount',
        header: 'Monto',
        align: 'right',
        sortable: true,
        cell: (row) => {
          const currency = currencyOf.get(row.accountId) ?? 'PEN';
          const value = signedAmount(row);
          return (
            <span className={`tabular-nums ${value >= 0 ? 'text-success' : 'text-destructive'}`}>
              {formatCurrency(value, currency)}
            </span>
          );
        },
      },
      {
        id: 'balanceAfter',
        header: 'Saldo después',
        align: 'right',
        cell: (row) => {
          const currency = currencyOf.get(row.accountId) ?? 'PEN';
          return <span className="tabular-nums">{formatCurrency(row.balanceAfter, currency)}</span>;
        },
      },
      {
        id: 'isReconciled',
        header: 'Estado',
        cell: (row) => (row.isReconciled ? <Badge variant="success">Conciliado</Badge> : <Badge variant="muted">Pendiente</Badge>),
      },
    ];
    if (!canConciliate) return base;
    return [
      ...base,
      {
        id: 'actions',
        header: '',
        width: 88,
        exportable: false,
        cell: (row) =>
          !row.isReconciled ? (
            <div className="flex justify-end">
              <Button
                variant="ghost"
                size="sm"
                loading={reconcileMutation.isPending}
                onClick={() => {
                  reconcileMutation.mutate(row.id, {
                    onSuccess: () => push({ title: 'Movimiento conciliado', variant: 'success' }),
                    onError: (err: unknown) =>
                      push({ title: 'No se pudo conciliar', description: err instanceof Error ? err.message : undefined, variant: 'destructive' }),
                  });
                }}
              >
                Conciliar
              </Button>
            </div>
          ) : null,
      },
    ];
  }, [canConciliate, currencyOf, reconcileMutation, push]);

  const transactions = transactionsQuery.data ?? [];
  const txAccountOptions = [
    { value: '', label: 'Todas las cuentas' },
    ...accounts.map((a) => ({ value: a.id, label: `${a.bank} — ${a.accountNumber}` })),
  ];

  return (
    <PageContainer>
      <PageHeader
        title="Tesorería"
        subtitle="Cuentas bancarias, movimientos y conciliaciones"
        actions={
          <>
            <Can permission={Permissions.Treasury.Create}>
              <Button
                variant="outline"
                onClick={() => {
                  setTxAccountId(txFilter);
                  setTxFormOpen(true);
                }}
              >
                <Icons.Action.Create /> Registrar movimiento
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
        <StatCard label="Total de cuentas" value={String(accounts.length)} icon={Icons.Navigation.Treasury} />
        <StatCard label="Cuentas activas" value={String(activeCount)} />
        <StatCard label="Saldo PEN" value={formatCurrency(totalBalancePen, 'PEN')} />
        <StatCard label="Saldo USD" value={formatCurrency(totalBalanceUsd, 'USD')} />
      </Grid>

      <Section title="Cuentas bancarias" description="Cuentas bancarias de la empresa y sus saldos actuales.">
        <DataTable
          columns={accountColumns}
          data={accounts}
          keyField="id"
          loading={accountsLoading}
          error={accountsError ? (accountsErrorObj as Error) : null}
          onRetry={() => refetchAccounts()}
          globalSearch
          exportFilename="cuentas-bancarias.csv"
          empty={
            <EmptyState
              title="No hay cuentas bancarias"
              description="Crea una cuenta para empezar a registrar movimientos de tesorería."
            />
          }
        />
      </Section>

      <Section
        title="Movimientos"
        description="Depósitos, retiros, comisiones e intereses registrados contra las cuentas."
        actions={
          <Can permission={Permissions.Treasury.Create}>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setTxAccountId(txFilter);
                setTxFormOpen(true);
              }}
            >
              <Icons.Action.Create /> Registrar movimiento
            </Button>
          </Can>
        }
      >
        <DataTable
          columns={txColumns}
          data={transactions}
          keyField="id"
          loading={transactionsQuery.isLoading}
          error={transactionsQuery.isError ? (transactionsQuery.error as Error) : null}
          onRetry={() => transactionsQuery.refetch()}
          globalSearch={false}
          exportFilename="movimientos-bancarios.csv"
          toolbarLeft={
            <Select value={txFilter} onValueChange={(v) => setTxFilter(v)}>
              <SelectTrigger className="w-56" aria-label="Filtrar por cuenta">
                <SelectValue placeholder="Todas las cuentas" />
              </SelectTrigger>
              <SelectContent>
                {txAccountOptions.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          }
          empty={
            <EmptyState
              title="No hay movimientos"
              description="Registra un depósito o retiro para verlo aquí."
            />
          }
        />
      </Section>

      <BankAccountFormDialog open={accountFormOpen} onOpenChange={setAccountFormOpen} account={editing} />

      <BankTransactionFormDialog open={txFormOpen} onOpenChange={setTxFormOpen} accountId={txAccountId} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar cuenta bancaria"
        description={`¿Eliminar la cuenta ${deleteTarget?.bank ?? 'esta cuenta'} — ${deleteTarget?.accountNumber ?? ''}? La cuenta quedará desactivada. Esta acción no se puede deshacer.`}
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
