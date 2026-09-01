import { z } from 'zod';

const SupplierDocumentTypeSchema = z.enum(['RUC', 'DNI', 'CE', 'PASSPORT']);

const SupplierSchema = z.object({
  id: z.string(),
  documentType: SupplierDocumentTypeSchema,
  documentNumber: z.string().min(8, 'Mínimo 8 caracteres'),
  businessName: z.string().min(1, 'Requerido').max(200),
  contactName: z.string().optional(),
  phone: z.string().optional(),
  email: z.email('Correo inválido').optional().or(z.literal('')),
  address: z.string().optional(),
  paymentTermDays: z.number().min(0, 'Debe ser >= 0').max(365, 'Máximo 365 días').optional(),
});

export const SupplierCreateSchema = SupplierSchema.omit({ id: true });
