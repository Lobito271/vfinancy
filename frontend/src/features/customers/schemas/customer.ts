import { z } from 'zod';
import { DocumentTypes } from '@/constants/countries';

export const CustomerStatusSchema = z.enum(['active', 'inactive', 'blocked']);

export const CustomerDocumentTypeSchema = z.enum(
  Object.keys(DocumentTypes) as [keyof typeof DocumentTypes, ...(keyof typeof DocumentTypes)[]],
);

export const CustomerSchema = z.object({
  id: z.string(),
  documentType: CustomerDocumentTypeSchema,
  documentNumber: z.string().min(8, 'Mínimo 8 caracteres'),
  businessName: z.string().min(1, 'Requerido').max(200),
  contactName: z.string().optional(),
  phone: z.string().regex(/^9\d{8}$/, 'Teléfono inválido (9 dígitos)').optional().or(z.literal('')),
  email: z.string().email('Correo inválido').optional().or(z.literal('')),
  address: z.string().optional(),
  creditLimit: z.number().min(0, 'Debe ser >= 0'),
  status: CustomerStatusSchema,
});

export type CustomerFormValues = z.infer<typeof CustomerSchema>;

export const CustomerCreateSchema = CustomerSchema.omit({ id: true });

export type CustomerCreateFormValues = z.infer<typeof CustomerCreateSchema>;

export const customerFilterSchema = z.object({
  search: z.string().optional(),
  status: CustomerStatusSchema.optional(),
});

export type CustomerFilterValues = z.infer<typeof customerFilterSchema>;
