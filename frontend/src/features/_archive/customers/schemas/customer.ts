import { z } from 'zod';
import { DocumentTypes } from '@/constants/countries';

const CustomerStatusSchema = z.enum(['active', 'inactive', 'blocked']);

const CustomerDocumentTypeSchema = z.enum(
  Object.keys(DocumentTypes) as [keyof typeof DocumentTypes, ...(keyof typeof DocumentTypes)[]],
);

const documentNumberValidators: Partial<Record<keyof typeof DocumentTypes, (v: string) => boolean | string>> = {
  DNI: (v) => /^\d{8}$/.test(v) || 'El DNI debe tener 8 dígitos.',
  RUC: (v) => /^(10|20)\d{9}$/.test(v) || 'El RUC debe tener 11 dígitos y comenzar con 10 o 20.',
  CE: (v) => /^\d{12}$/.test(v) || 'El Carnet de Extranjería debe tener 12 dígitos.',
  PASSPORT: () => true,
};

const CustomerSchema = z
  .object({
    id: z.string(),
    documentType: CustomerDocumentTypeSchema,
    documentNumber: z.string().min(1, 'Requerido'),
    businessName: z.string().min(1, 'Requerido').max(200),
    contactName: z.string().optional(),
    phone: z.string().regex(/^9\d{8}$/, 'Teléfono inválido (9 dígitos)').optional().or(z.literal('')),
    email: z.email('Correo inválido').optional().or(z.literal('')),
    address: z.string().optional(),
    creditLimit: z.number().min(0, 'Debe ser >= 0'),
    status: CustomerStatusSchema,
  })
  .superRefine((data, ctx) => {
    const validate = documentNumberValidators[data.documentType];
    if (!validate) return;
    const result = validate(data.documentNumber);
    if (result !== true) ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['documentNumber'], message: result as string });
  });

export const CustomerCreateSchema = CustomerSchema.omit({ id: true });
