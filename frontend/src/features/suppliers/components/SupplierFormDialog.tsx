import { useMemo } from 'react';
import { z } from 'zod';
import { Form, TextField, SelectField, NumberField } from '@/components/form';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { useCreateSupplier, useUpdateSupplier } from '@/features/suppliers/hooks/useSuppliers';
import { SupplierCreateSchema } from '@/features/suppliers/schemas/supplier';
import type { Supplier } from '@/types/domain';
import { DocumentTypes } from '@/constants/countries';
import { useNotificationStore } from '@/stores/notification';

const SupplierFormSchema = SupplierCreateSchema.extend({
  status: z.enum(['active', 'inactive']),
});

type SupplierFormValues = z.infer<typeof SupplierFormSchema>;

const documentOptions = Object.values(DocumentTypes).map((d) => ({ value: d.code, label: d.name }));

const statusOptions = [
  { value: 'active', label: 'Activo' },
  { value: 'inactive', label: 'Inactivo' },
];

interface SupplierFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  supplier?: Supplier | null;
}

export function SupplierFormDialog({ open, onOpenChange, supplier }: SupplierFormDialogProps) {
  const create = useCreateSupplier();
  const update = useUpdateSupplier();
  const push = useNotificationStore((s) => s.push);

  const defaults = useMemo<SupplierFormValues>(
    () => ({
      documentType: supplier?.documentType ?? 'RUC',
      documentNumber: supplier?.documentNumber ?? '',
      businessName: supplier?.businessName ?? '',
      contactName: supplier?.contactName ?? '',
      phone: supplier?.phone ?? '',
      email: supplier?.email ?? '',
      address: supplier?.address ?? '',
      paymentTermDays: supplier?.paymentTermDays ?? 0,
      status: supplier ? (supplier.status === 'active' ? 'active' : 'inactive') : 'active',
    }),
    [supplier],
  );

  const handleSubmit = (values: SupplierFormValues) => {
    const payload = {
      documentType: values.documentType,
      documentNumber: values.documentNumber,
      businessName: values.businessName,
      contactName: values.contactName || undefined,
      phone: values.phone || undefined,
      email: values.email || undefined,
      address: values.address || undefined,
      paymentTermDays: values.paymentTermDays ?? 0,
    };
    const onSuccess = () => {
      push({ title: supplier ? 'Proveedor actualizado' : 'Proveedor creado', variant: 'success' });
      onOpenChange(false);
    };
    const onError = (err: unknown) => {
      push({ title: 'No se pudo guardar el proveedor', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    };
    if (supplier) {
      update.mutate({ id: supplier.id, ...payload, status: values.status }, { onSuccess, onError });
    } else {
      create.mutate(payload, { onSuccess, onError });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{supplier ? 'Editar proveedor' : 'Nuevo proveedor'}</DialogTitle>
          <DialogDescription>
            {supplier ? 'Actualiza los datos del proveedor.' : 'Registra un nuevo proveedor en el sistema.'}
          </DialogDescription>
        </DialogHeader>

        <Form<SupplierFormValues> schema={SupplierFormSchema} defaultValues={defaults} onSubmit={handleSubmit}>
          {({ formState }) => (
            <>
              <div className="max-h-[70vh] space-y-4 overflow-y-auto pr-2">
                <Grid cols={2}>
                  <SelectField name="documentType" label="Tipo de documento" required options={documentOptions} clearable={false} />
                  <TextField name="documentNumber" label="Número de documento" required />
                </Grid>
                <TextField name="businessName" label="Razón social" required />
                <Grid cols={2}>
                  <TextField name="contactName" label="Persona de contacto" />
                  <NumberField name="paymentTermDays" label="Días de plazo de pago" min={0} max={365} />
                </Grid>
                <Grid cols={2}>
                  <TextField name="phone" label="Teléfono" type="tel" />
                  <TextField name="email" label="Correo electrónico" type="email" />
                </Grid>
                <TextField name="address" label="Dirección" />
                {supplier && (
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
