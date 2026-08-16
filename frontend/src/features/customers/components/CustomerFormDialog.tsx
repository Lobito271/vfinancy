import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField, SelectField, MoneyField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { useCreateCustomer, useUpdateCustomer } from '@/features/customers';
import { CustomerCreateSchema } from '@/features/customers/schemas/customer';
import type { Customer } from '@/types/domain';
import { DocumentTypes } from '@/constants/countries';
import { useNotificationStore } from '@/stores/notification';

const CustomerFormSchema = CustomerCreateSchema.extend({
  status: z.enum(['active', 'inactive']),
});

type CustomerFormValues = z.infer<typeof CustomerFormSchema>;

const documentOptions = Object.values(DocumentTypes).map((d) => ({ value: d.code, label: d.name }));

const statusOptions = [
  { value: 'active', label: 'Activo' },
  { value: 'inactive', label: 'Inactivo' },
];

interface CustomerFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  customer?: Customer | null;
}

export function CustomerFormDialog({ open, onOpenChange, customer }: CustomerFormDialogProps) {
  const create = useCreateCustomer();
  const update = useUpdateCustomer();
  const push = useNotificationStore((s) => s.push);

  const defaults = useMemo<CustomerFormValues>(
    () => ({
      documentType: customer?.documentType ?? 'DNI',
      documentNumber: customer?.documentNumber ?? '',
      businessName: customer?.businessName ?? '',
      contactName: customer?.contactName ?? '',
      phone: customer?.phone ?? '',
      email: customer?.email ?? '',
      address: customer?.address ?? '',
      creditLimit: customer?.creditLimit ?? 0,
      status: customer ? (customer.status === 'active' ? 'active' : 'inactive') : 'active',
    }),
    [customer],
  );

  const handleSubmit = (values: CustomerFormValues) => {
    const payload = {
      documentType: values.documentType,
      documentNumber: values.documentNumber,
      businessName: values.businessName,
      contactName: values.contactName || undefined,
      phone: values.phone || undefined,
      email: values.email || undefined,
      address: values.address || undefined,
      creditLimit: values.creditLimit,
    };
    const onSuccess = () => {
      push({ title: customer ? 'Cliente actualizado' : 'Cliente creado', variant: 'success' });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({ title: 'No se pudo guardar el cliente', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    };
    if (customer) {
      update.mutate({ id: customer.id, ...payload, status: values.status }, { onSuccess, onError });
    } else {
      create.mutate(payload, { onSuccess, onError });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{customer ? 'Editar cliente' : 'Nuevo cliente'}</DialogTitle>
          <DialogDescription>
            {customer ? 'Actualiza los datos del cliente.' : 'Registra un nuevo cliente en el sistema.'}
          </DialogDescription>
        </DialogHeader>

        <Form schema={CustomerFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <Grid cols={2}>
                  <SelectField name="documentType" label="Tipo de documento" required options={documentOptions} clearable={false} />
                  <TextField name="documentNumber" label="Número de documento" required />
                </Grid>
                <TextField name="businessName" label="Razón social / Nombre" required />
                <Grid cols={2}>
                  <TextField name="contactName" label="Persona de contacto" />
                  <MoneyField name="creditLimit" label="Límite de crédito" />
                </Grid>
                <Grid cols={2}>
                  <TextField name="phone" label="Teléfono" type="tel" />
                  <TextField name="email" label="Correo electrónico" type="email" />
                </Grid>
                <TextField name="address" label="Dirección" />
                {customer && (
                  <Grid cols={2}>
                    <SelectField name="status" label="Estado" options={statusOptions} clearable={false} />
                  </Grid>
                )}
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending || update.isPending}>
                  Cancelar
                </Button>
                <Button type="submit" loading={create.isPending || update.isPending} disabled={!formState.isValid}>
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
