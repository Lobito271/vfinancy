import { useMemo, useState } from 'react';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Form, NumberField, SelectField, TextField } from '@/components/form';
import { DialogBody, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, AlertDialog } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';
import type { CompanyDTO } from '@/services/wails-types';

const companySchema = z.object({
  code: z.string().trim().min(2, 'Ingresa un código de empresa.'),
  legalName: z.string().trim().min(2, 'Ingresa la razón social.'),
  tradeName: z.string().trim().min(2, 'Ingresa el nombre comercial.'),
  taxId: z.string().trim().min(8, 'Ingresa el número de identificación fiscal.'),
  address: z.string().trim(),
  phone: z.string().trim(),
  email: z.string().trim().email('Ingresa un correo válido.'),
  countryCode: z.string().min(2),
  functionalCurrency: z.string().min(3),
  timezone: z.string().min(1),
  fiscalYearStartMonth: z.number().int().min(1).max(12),
});

type CompanyValues = z.infer<typeof companySchema>;

interface CompanyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company?: CompanyDTO | null;
  isActive?: boolean;
}

function toDefaults(company?: CompanyDTO | null): CompanyValues {
  return {
    code: company?.code ?? '',
    legalName: company?.legalName ?? '',
    tradeName: company?.tradeName ?? '',
    taxId: company?.taxId ?? '',
    address: company?.address ?? '',
    phone: company?.phone ?? '',
    email: company?.email ?? '',
    countryCode: company?.countryCode ?? 'PE',
    functionalCurrency: company?.functionalCurrency ?? 'PEN',
    timezone: company?.timezone ?? 'America/Lima',
    fiscalYearStartMonth: company?.fiscalYearStartMonth ?? 1,
  };
}

export function CompanyDialog({ open, onOpenChange, company, isActive }: CompanyDialogProps) {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const defaults = useMemo(() => toDefaults(company), [company]);

  const refetchAll = async () => {
    await queryClient.clear();
    await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
  };

  const save = useMutation({
    mutationFn: async (values: CompanyValues) => {
      const payload = { ...values, id: company?.id ?? '', isActive: true };
      if (company) {
        await wailsClient.updateCompany(payload);
      } else {
        const created = await wailsClient.createCompany(payload);
        await wailsClient.setActiveCompany(created.id);
      }
    },
    onSuccess: async () => {
      push({ title: company ? 'Empresa actualizada' : 'Empresa creada', variant: 'success' });
      onOpenChange(false);
      await refetchAll();
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar la empresa', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const deactivate = useMutation({
    mutationFn: async () => {
      if (company) await wailsClient.deactivateCompany(company.id);
    },
    onSuccess: async () => {
      push({ title: 'Empresa desactivada', variant: 'success' });
      onOpenChange(false);
      await refetchAll();
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo desactivar la empresa', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent size="lg">
          <DialogHeader>
            <DialogTitle>{company ? 'Editar empresa' : 'Nueva empresa'}</DialogTitle>
            <DialogDescription>
              {company ? 'Actualiza los datos del negocio.' : 'Registra un negocio adicional y actívalo.'}
            </DialogDescription>
          </DialogHeader>

          <Form<CompanyValues> schema={companySchema} defaultValues={defaults} onSubmit={(values) => save.mutate(values)}>
            {({ formState }) => (
              <>
                <DialogBody>
                  <Grid cols={2}>
                    <TextField name="code" label="Código interno" required description="Ejemplo: ACME" />
                    <TextField name="taxId" label="RUC o identificación fiscal" required />
                  </Grid>
                  <Grid cols={2}>
                    <TextField name="legalName" label="Razón social" required />
                    <TextField name="tradeName" label="Nombre comercial" required />
                  </Grid>
                  <Grid cols={2}>
                    <TextField name="phone" label="Teléfono" type="tel" />
                    <TextField name="email" label="Correo" type="email" />
                  </Grid>
                  <TextField name="address" label="Dirección" />
                  <Grid cols={2}>
                    <SelectField name="countryCode" label="País" required options={[{ value: 'PE', label: 'Perú' }]} clearable={false} />
                    <SelectField name="functionalCurrency" label="Moneda funcional" required options={[{ value: 'PEN', label: 'PEN · Sol peruano' }, { value: 'USD', label: 'USD · Dólar estadounidense' }]} clearable={false} />
                  </Grid>
                  <Grid cols={2}>
                    <SelectField name="timezone" label="Zona horaria" required options={[{ value: 'America/Lima', label: 'America/Lima' }]} clearable={false} />
                    <NumberField name="fiscalYearStartMonth" label="Mes de inicio fiscal" required min={1} max={12} />
                  </Grid>
                </DialogBody>
                <DialogFooter>
                  {company && !isActive && (
                    <Button variant="destructive" type="button" onClick={() => setConfirmOpen(true)} disabled={save.isPending}>
                      Desactivar
                    </Button>
                  )}
                  <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={save.isPending}>
                    Cancelar
                  </Button>
                  <Button type="submit" loading={save.isPending} disabled={!formState.isValid}>
                    Guardar
                  </Button>
                </DialogFooter>
              </>
            )}
          </Form>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        variant="destructive"
        title="Desactivar empresa"
        description={`Desactivarás ${company?.legalName ?? 'esta empresa'} y dejará de aparecer en el selector.`}
        confirmLabel="Desactivar"
        onConfirm={() => deactivate.mutate()}
        loading={deactivate.isPending}
      />
    </>
  );
}