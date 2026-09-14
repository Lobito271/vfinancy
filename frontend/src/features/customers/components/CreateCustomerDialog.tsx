import { useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { Dialog, DialogBody, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Button } from '@/components/button';
import { Form, TextField } from '@/components/form';
import { customersService } from '@/services/customers';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';

const schema = z.object({
  businessName: z.string().min(1, 'Ingrese el nombre o razón social'),
  phone: z.string().optional(),
  email: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

interface CreateCustomerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (customerId: string) => void;
}

export function CreateCustomerDialog({ open, onOpenChange, onCreated }: CreateCustomerDialogProps) {
  const push = useNotificationStore((s) => s.push);
  const qc = useQueryClient();
  const create = useMutation({
    mutationFn: (values: FormValues) =>
      customersService.create({
        businessName: values.businessName,
        documentType: '',
        documentNumber: '',
        email: values.email ?? '',
        phone: values.phone ?? '',
        address: '',
      }),
    onSuccess: (created) => {
      void qc.invalidateQueries({ queryKey: queryKeys.customers.all });
      onCreated(created.id);
      push({ title: 'Cliente creado', variant: 'success' });
      onOpenChange(false);
    },
    onError: (err: unknown) => {
      push({
        title: 'No se pudo crear el cliente',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    },
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Nuevo cliente</DialogTitle>
          <DialogDescription>
            Se creará un cliente sin documento fiscal; podrás completarlo después.
          </DialogDescription>
        </DialogHeader>
        <Form<FormValues> schema={schema} defaultValues={{ businessName: '', phone: '', email: '' }} onSubmit={(values) => create.mutate(values)}>
          <DialogBody>
            <TextField name="businessName" label="Nombre / Razón social" required />
            <TextField name="phone" label="Teléfono" />
            <TextField name="email" label="Correo" type="email" />
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" type="button" onClick={() => onOpenChange(false)} disabled={create.isPending}>
              Cancelar
            </Button>
            <Button type="submit" loading={create.isPending}>
              Guardar
            </Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}