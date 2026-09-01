import { z } from 'zod';
import { DocumentTypes } from '@/constants/countries';

const CustomerStatusSchema = z.enum(['active', 'inactive', 'blocked']);

const CustomerDocumentTypeSchema = z.enum(
  Object.keys(DocumentTypes) as [keyof typeof DocumentTypes, ...(keyof typeof DocumentTypes)[]],
);

const CustomerSchema = z.object({
  id: z.string(),
  documentType: CustomerDocumentTypeSchema,
  documentNumber: z.string().min(8, 'Mínimo 8 caracteres'),
  businessName: z.string().min(1, 'Requerido').max(200),
  contactName: z.string().optional(),
  phone: z.string().regex(/^9\d{8}$/, 'Teléfono inválido (9 dígitos)').optional().or(z.literal('')),
  email: z.email('Correo inválido').optional().or(z.literal('')),
  address: z.string().optional(),
  creditLimit: z.number().min(0, 'Debe ser >= 0'),
  status: CustomerStatusSchema,
});

export const CustomerCreateSchema = CustomerSchema.omit({ id: true });
