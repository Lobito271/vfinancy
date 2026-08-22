import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField, SelectField, CheckboxField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { Currencies, type CurrencyCode } from '@/constants/currencies';
import { useCreateBankAccount, useUpdateBankAccount } from '@/features/treasury/hooks/useTreasury';
import type { BankAccount, BankAccountType } from '@/services/treasury';
import { useNotificationStore } from '@/stores/notification';

const BankAccountSchema = z.object({
  bank: z.string().min(1, 'Requerido').max(120),
  accountNumber: z.string().min(1, 'Requerido').max(40),
  accountType: z.enum(['checking', 'savings']),
  currency: z.enum(['PEN', 'USD', 'EUR', 'MXN', 'COP', 'CLP', 'BRL']),
  isDefault: z.boolean(),
});

type BankAccountFormValues = z.infer<typeof BankAccountSchema>;

interface BankAccountFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  account?: BankAccount | null;
}

const currencyOptions = Object.entries(Currencies).map(([code, c]) => ({
  value: code,
  label: `${c.code} — ${c.name}`,
}));

const accountTypeOptions: { value: BankAccountType; label: string }[] = [
  { value: 'checking', label: 'Cuenta corriente' },
  { value: 'savings', label: 'Cuenta de ahorros' },
];

export function BankAccountFormDialog({ open, onOpenChange, account }: BankAccountFormDialogProps) {
  const create = useCreateBankAccount();
  const update = useUpdateBankAccount();
  const push = useNotificationStore((s) => s.push);

  const defaults = useMemo<BankAccountFormValues>(
    () => ({
      bank: account?.bank ?? '',
      accountNumber: account?.accountNumber ?? '',
      accountType: account?.accountType ?? 'checking',
      currency: (account?.currency as CurrencyCode) ?? 'PEN',
      isDefault: account?.isDefault ?? false,
    }),
    [account],
  );

  const handleSubmit = (values: BankAccountFormValues) => {
    const onSuccess = () => {
      push({ title: account ? 'Cuenta actualizada' : 'Cuenta creada', variant: 'success' });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({ title: 'No se pudo guardar la cuenta', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    };

    const input = {
      bank: values.bank,
      accountNumber: values.accountNumber,
      accountType: values.accountType,
      currency: values.currency,
      isDefault: values.isDefault,
    };

    if (account) {
      update.mutate(
        { id: account.id, input: { ...input, isActive: account.isActive } },
        { onSuccess, onError },
      );
      return;
    }

    create.mutate(input, { onSuccess, onError });
  };

  const loading = create.isPending || update.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{account ? 'Editar cuenta bancaria' : 'Nueva cuenta bancaria'}</DialogTitle>
          <DialogDescription>
            {account ? 'Actualiza los datos de la cuenta bancaria.' : 'Registra una cuenta bancaria de la empresa.'}
          </DialogDescription>
        </DialogHeader>

        <Form<BankAccountFormValues> schema={BankAccountSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="dialog-body-scroll">
                <Grid cols={2}>
                  <TextField name="bank" label="Banco" required placeholder="Ej. Banco de Crédito" />
                  <TextField name="accountNumber" label="N.º de cuenta" required placeholder="Ej. 191-1234567-0-12" />
                </Grid>
                <Grid cols={2}>
                  <SelectField name="accountType" label="Tipo de cuenta" required options={accountTypeOptions} />
                  <SelectField name="currency" label="Moneda" required options={currencyOptions} />
                </Grid>
                <CheckboxField name="isDefault" label="Cuenta principal por defecto" description="Se usará por omisión para los movimientos de tesorería." />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={loading}>
                  Cancelar
                </Button>
                <Button type="submit" loading={loading} disabled={!formState.isValid}>
                  Guardar
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
