import { useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import { z } from 'zod';
import { Form, TextField, SelectField, MoneyField, DateField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { Currencies, type CurrencyCode } from '@/constants/currencies';
import { useBankAccounts, useCreateBankTransaction } from '@/features/treasury/hooks/useTreasury';
import { useNotificationStore } from '@/stores/notification';

const BankTransactionSchema = z.object({
  accountId: z.string().min(1, 'Seleccione una cuenta'),
  type: z.enum(['deposit', 'withdrawal', 'fee', 'interest', 'transfer', 'other']),
  date: z.string().min(1, 'Requerido'),
  amount: z.number().min(0.01, 'Debe ser mayor a 0'),
  description: z.string().min(1, 'Requerido').max(200),
  reference: z.string().max(60).optional(),
});

type BankTransactionFormValues = z.infer<typeof BankTransactionSchema>;

interface BankTransactionFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  accountId?: string;
}

const transactionTypeOptions = [
  { value: 'deposit', label: 'Depósito' },
  { value: 'withdrawal', label: 'Retiro' },
  { value: 'fee', label: 'Comisión' },
  { value: 'interest', label: 'Interés' },
  { value: 'transfer', label: 'Transferencia' },
  { value: 'other', label: 'Otro' },
];

function today(): string {
  const d = new Date();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

function AmountField() {
  const { watch } = useFormContext<BankTransactionFormValues>();
  const accountId = watch('accountId');
  const { data: accounts } = useBankAccounts();
  const account = accounts?.find((a) => a.id === accountId);
  const currency = (account?.currency ?? 'PEN') as CurrencyCode;
  return <MoneyField name="amount" label="Monto" required currency={currency} />;
}

export function BankTransactionFormDialog({ open, onOpenChange, accountId }: BankTransactionFormDialogProps) {
  const create = useCreateBankTransaction();
  const push = useNotificationStore((s) => s.push);
  const { data: accounts } = useBankAccounts();

  const accountOptions = useMemo(
    () =>
      (accounts ?? [])
        .filter((a) => a.isActive)
        .map((a) => ({
          value: a.id,
          label: `${a.bank} — ${a.accountNumber} (${Currencies[a.currency as CurrencyCode]?.code ?? a.currency})`,
        })),
    [accounts],
  );

  const defaults = useMemo<BankTransactionFormValues>(
    () => ({
      accountId: accountId ?? '',
      type: 'deposit',
      date: today(),
      amount: 0,
      description: '',
      reference: '',
    }),
    [accountId],
  );

  const handleSubmit = (values: BankTransactionFormValues) => {
    create.mutate(
      {
        accountId: values.accountId,
        type: values.type,
        date: values.date,
        amount: values.amount,
        description: values.description,
        reference: values.reference || undefined,
      },
      {
        onSuccess: () => {
          push({ title: 'Movimiento registrado', variant: 'success' });
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({ title: 'No se pudo registrar el movimiento', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
        },
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>Registrar movimiento</DialogTitle>
          <DialogDescription>Registra un depósito, retiro, comisión u otro movimiento bancario.</DialogDescription>
        </DialogHeader>

        <Form<BankTransactionFormValues> schema={BankTransactionSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <Grid cols={2}>
                  <SelectField name="accountId" label="Cuenta bancaria" required placeholder="Seleccione la cuenta…" options={accountOptions} />
                  <SelectField name="type" label="Tipo de movimiento" required options={transactionTypeOptions} />
                </Grid>
                <Grid cols={2}>
                  <DateField name="date" label="Fecha" required />
                  <AmountField />
                </Grid>
                <TextField name="description" label="Descripción" required placeholder="Ej. Deposito de venta del día" />
                <TextField name="reference" label="Referencia" placeholder="Ej. N.º de operación" />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={create.isPending} disabled={!formState.isValid}>
                  Registrar
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
