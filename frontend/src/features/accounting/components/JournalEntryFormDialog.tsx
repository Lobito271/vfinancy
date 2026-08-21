import { useMemo } from 'react';
import { useFieldArray, useFormContext, type Path } from 'react-hook-form';
import { z } from 'zod';
import { Form, TextField, SelectField, MoneyField, DateField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { Icons } from '@/design-system/icons';
import { useChartOfAccounts, useCreateJournalEntry } from '@/features/accounting/hooks/useAccounting';
import { useNotificationStore } from '@/stores/notification';
import { formatCurrency } from '@/utils/format';

const LineSchema = z.object({
  accountId: z.string().min(1, 'Seleccione una cuenta'),
  debit: z.number().min(0),
  credit: z.number().min(0),
});

const JournalEntrySchema = z
  .object({
    entryDate: z.string().min(1, 'Requerido'),
    description: z.string().min(1, 'Requerido').max(200),
    lines: z.array(LineSchema).min(2, 'Registre al menos dos líneas'),
  })
  .superRefine((val, ctx) => {
    for (const [i, l] of val.lines.entries()) {
      if ((l.debit ?? 0) === 0 && (l.credit ?? 0) === 0) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Indique débito o crédito', path: ['lines', i, 'debit'] });
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Indique débito o crédito', path: ['lines', i, 'credit'] });
      }
    }
    const totalDebit = val.lines.reduce((s, l) => s + (l.debit ?? 0), 0);
    const totalCredit = val.lines.reduce((s, l) => s + (l.credit ?? 0), 0);
    if (Math.abs(totalDebit - totalCredit) > 0.005) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'El total de débitos debe ser igual al de créditos', path: ['lines'] });
    }
  });

type JournalEntryFormValues = z.infer<typeof JournalEntrySchema>;

interface JournalEntryFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface AccountSelectOption {
  value: string;
  label: string;
}

function today(): string {
  const d = new Date();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${m}-${day}`;
}

function newLine() {
  return { accountId: '', debit: 0, credit: 0 };
}

function JournalLineRow({
  index,
  accountOptions,
  onRemove,
}: {
  index: number;
  accountOptions: AccountSelectOption[];
  onRemove: () => void;
}) {
  return (
    <div className="journal-line">
      <div className="hstack hstack--start hstack--sm hstack--between">
        <div className="grow">
          <SelectField
            name={`lines.${index}.accountId` as Path<JournalEntryFormValues>}
            label={`Cuenta ${index + 1}`}
            required
            options={accountOptions}
            clearable={false}
          />
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="journal-line__remove"
          aria-label={`Quitar línea ${index + 1}`}
          onClick={onRemove}
        >
          <Icons.Action.Delete />
        </Button>
      </div>
      <Grid cols={2}>
        <MoneyField name={`lines.${index}.debit` as Path<JournalEntryFormValues>} label="Débito" currency="PEN" />
        <MoneyField name={`lines.${index}.credit` as Path<JournalEntryFormValues>} label="Crédito" currency="PEN" />
      </Grid>
    </div>
  );
}

function BalanceFooter() {
  const { watch } = useFormContext<JournalEntryFormValues>();
  const lines = watch('lines');
  const totalDebit = lines?.reduce((s, l) => s + (l?.debit ?? 0), 0) ?? 0;
  const totalCredit = lines?.reduce((s, l) => s + (l?.credit ?? 0), 0) ?? 0;
  const unbalanced = Math.abs(totalDebit - totalCredit) > 0.005;

  return (
    <div className="balance-box">
      <div className="split-row">
        <span className="muted">Total débitos</span>
        <span className="tabular">{formatCurrency(totalDebit, 'PEN')}</span>
      </div>
      <div className="split-row">
        <span className="muted">Total créditos</span>
        <span className="tabular">{formatCurrency(totalCredit, 'PEN')}</span>
      </div>
      <div className={`split-row fw-medium ${unbalanced ? 'color-destructive' : 'color-success'}`}>
        <span>{unbalanced ? 'Asiento descuadrado' : 'Asiento cuadrado'}</span>
        <span className="tabular">{formatCurrency(totalDebit - totalCredit, 'PEN')}</span>
      </div>
      {unbalanced && <p className="error-note">El total de débitos debe ser igual al de créditos.</p>}
    </div>
  );
}

function JournalLinesEditor({ accountOptions }: { accountOptions: AccountSelectOption[] }) {
  const { control } = useFormContext<JournalEntryFormValues>();
  const { fields, append, remove } = useFieldArray<JournalEntryFormValues, 'lines', 'id'>({ control, name: 'lines' });

  return (
    <div className="stack stack--md">
      <div className="stack stack--md">
        {fields.map((field, index) => (
          <JournalLineRow key={field.id} index={index} accountOptions={accountOptions} onRemove={() => remove(index)} />
        ))}
      </div>
      <Button type="button" variant="outline" size="sm" onClick={() => append(newLine())}>
        <Icons.Action.Create /> Agregar línea
      </Button>
      <BalanceFooter />
    </div>
  );
}

export function JournalEntryFormDialog({ open, onOpenChange }: JournalEntryFormDialogProps) {
  const create = useCreateJournalEntry();
  const push = useNotificationStore((s) => s.push);
  const { data: accounts = [] } = useChartOfAccounts();

  const accountOptions = useMemo(
    () =>
      accounts
        .filter((a) => a.isActive && a.allowsMovement)
        .map((a) => ({ value: a.id, label: `${a.code} — ${a.name}` })),
    [accounts],
  );

  const defaults = useMemo<JournalEntryFormValues>(
    () => ({ entryDate: today(), description: '', lines: [newLine(), newLine()] }),
    [],
  );

  const handleSubmit = (values: JournalEntryFormValues) => {
    create.mutate(
      {
        entryDate: values.entryDate,
        description: values.description,
        lines: values.lines.map((l) => ({
          accountId: l.accountId,
          debit: l.debit,
          credit: l.credit,
        })),
      },
      {
        onSuccess: () => {
          push({ title: 'Asiento creado', variant: 'success' });
          onOpenChange(false);
        },
        onError: (err: unknown) => {
          push({ title: 'No se pudo crear el asiento', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
        },
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>Nuevo asiento</DialogTitle>
          <DialogDescription>Registra un asiento de diario. Los débitos deben ser iguales a los créditos.</DialogDescription>
        </DialogHeader>

        <Form<JournalEntryFormValues> schema={JournalEntrySchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="dialog-body-scroll">
                <Grid cols={2}>
                  <DateField name="entryDate" label="Fecha" required />
                  <TextField name="description" label="Descripción" required placeholder="Ej. Pago a proveedor, apertura…" />
                </Grid>
                <JournalLinesEditor accountOptions={accountOptions} />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={create.isPending} disabled={!formState.isValid}>
                  Guardar asiento
                </Button>
              </DialogFooter>
            </>
          )}
        </Form>
      </DialogContent>
    </Dialog>
  );
}
