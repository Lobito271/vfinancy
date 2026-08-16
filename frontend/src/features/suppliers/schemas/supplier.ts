import { z } from 'zod';

export const SupplierDocumentTypeSchema = z.enum(['RUC', 'DNI', 'CE', 'PASSPORT']);

export const SupplierSchema = z.object({
  id: z.string(),
  documentType: SupplierDocumentTypeSchema,
  documentNumber: z.string().min(8, 'Mínimo 8 caracteres'),
  businessName: z.string().min(1, 'Requerido').max(200),
  contactName: z.string().optional(),
  phone: z.string().optional(),
  email: z.string().email('Correo inválido').optional().or(z.literal('')),
  address: z.string().optional(),
  paymentTermDays: z.number().min(0, 'Debe ser >= 0').max(365, 'Máximo 365 días').optional(),
});

export type SupplierFormValues = z.infer<typeof SupplierSchema>;

export const SupplierCreateSchema = SupplierSchema.omit({ id: true });

export type SupplierCreateFormValues = z.infer<typeof SupplierCreateSchema>;
