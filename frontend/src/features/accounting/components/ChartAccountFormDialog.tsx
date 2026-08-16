import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField, SelectField, TextareaField, CheckboxField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { useChartOfAccounts, useCreateChartOfAccount, useUpdateChartOfAccount } from '@/features/accounting/hooks/useAccounting';
import type { Account, AccountType } from '@/services/accounting';
import { useNotificationStore } from '@/stores/notification';

const CodeSchema = z
  .string()
  .min(1, 'Requerido')
  .regex(/^\d{1,2}(\.\d{1,2}){0,4}$/, 'Código inválido: 1–2 dígitos por nivel, máx. 5 niveles');

const ChartAccountSchema = z.object({
  code: CodeSchema,
  name: z.string().min(1, 'Requerido').max(120),
  type: z.enum(['asset', 'liability', 'equity', 'income', 'expense']),
  parentId: z.string().optional(),
  allowsMovement: z.boolean(),
  description: z.string().max(255).optional(),
});

type ChartAccountFormValues = z.infer<typeof ChartAccountSchema>;

interface ChartAccountFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  account?: Account | null;
}

const accountTypeOptions: { value: AccountType; label: string }[] = [
  { value: 'asset', label: 'Activo' },
  { value: 'liability', label: 'Pasivo' },
  { value: 'equity', label: 'Patrimonio' },
  { value: 'income', label: 'Ingreso' },
  { value: 'expense', label: 'Gasto' },
];

export function ChartAccountFormDialog({ open, onOpenChange, account }: ChartAccountFormDialogProps) {
  const create = useCreateChartOfAccount();
  const update = useUpdateChartOfAccount();
  const { data: accounts = [] } = useChartOfAccounts();
  const push = useNotificationStore((s) => s.push);

  const parentOptions = useMemo(() => {
    const excluded = new Set<string>();
    if (account) {
      excluded.add(account.id);
      for (const a of accounts) {
        if (account.path && a.path.startsWith(`${account.path}.`)) excluded.add(a.id);
      }
    }
    return accounts
      .filter((a) => a.isActive && !excluded.has(a.id))
      .map((a) => ({ value: a.id, label: `${a.code} — ${a.name}` }));
  }, [accounts, account]);

  const defaults = useMemo<ChartAccountFormValues>(
    () => ({
      code: account?.code ?? '',
      name: account?.name ?? '',
      type: account?.type ?? 'asset',
      parentId: account?.parentId || undefined,
      allowsMovement: account?.allowsMovement ?? false,
      description: account?.description || undefined,
    }),
    [account],
  );

  const handleSubmit = (values: ChartAccountFormValues) => {
    const onSuccess = () => {
      push({ title: account ? 'Cuenta actualizada' : 'Cuenta creada', variant: 'success' });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({ title: 'No se pudo guardar la cuenta', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    };

    const input = {
      code: values.code,
      name: values.name,
      type: values.type,
      parentId: values.parentId,
      allowsMovement: values.allowsMovement,
      description: values.description,
    };

    if (account) {
      update.mutate({ id: account.id, input: { ...input, isActive: account.isActive } }, { onSuccess, onError });
      return;
    }

    create.mutate(input, { onSuccess, onError });
  };

  const loading = create.isPending || update.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{account ? 'Editar cuenta contable' : 'Nueva cuenta contable'}</DialogTitle>
          <DialogDescription>
            {account ? 'Actualiza los datos de la cuenta del plan de cuentas.' : 'Registra una cuenta en el plan de cuentas.'}
          </DialogDescription>
        </DialogHeader>

        <Form<ChartAccountFormValues> schema={ChartAccountSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <Grid cols={2}>
                  <TextField name="code" label="Código" required placeholder="Ej. 104.01" />
                  <TextField name="name" label="Nombre" required placeholder="Ej. Bancos — cuenta corriente" />
                </Grid>
                <Grid cols={2}>
                  <SelectField name="type" label="Tipo" required options={accountTypeOptions} />
                  <SelectField name="parentId" label="Cuenta padre" placeholder="Sin cuenta padre" options={parentOptions} />
                </Grid>
                <CheckboxField
                  name="allowsMovement"
                  label="Permite movimiento"
                  description="Marque solo las cuentas de nivel hoja donde se registrarán asientos."
                />
                <TextareaField name="description" label="Descripción" placeholder="Detalle opcional de la cuenta." rows={2} />
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
