export const Countries = {
  PE: { code: 'PE', name: 'Perú', locale: 'es-PE', currency: 'PEN', taxIdLabel: 'RUC', personIdLabel: 'DNI' },
  MX: { code: 'MX', name: 'México', locale: 'es-MX', currency: 'MXN', taxIdLabel: 'RFC', personIdLabel: 'CURP' },
  CO: { code: 'CO', name: 'Colombia', locale: 'es-CO', currency: 'COP', taxIdLabel: 'NIT', personIdLabel: 'CC' },
  CL: { code: 'CL', name: 'Chile', locale: 'es-CL', currency: 'CLP', taxIdLabel: 'RUT', personIdLabel: 'RUN' },
  AR: { code: 'AR', name: 'Argentina', locale: 'es-AR', currency: 'ARS', taxIdLabel: 'CUIT', personIdLabel: 'DNI' },
  EC: { code: 'EC', name: 'Ecuador', locale: 'es-EC', currency: 'USD', taxIdLabel: 'RUC', personIdLabel: 'Cédula' },
  BO: { code: 'BO', name: 'Bolivia', locale: 'es-BO', currency: 'BOB', taxIdLabel: 'NIT', personIdLabel: 'CI' },
  US: { code: 'US', name: 'Estados Unidos', locale: 'en-US', currency: 'USD', taxIdLabel: 'EIN', personIdLabel: 'SSN' },
  ES: { code: 'ES', name: 'España', locale: 'es-ES', currency: 'EUR', taxIdLabel: 'NIF', personIdLabel: 'DNI' },
} as const;

export type CountryCode = keyof typeof Countries;
export const DefaultCountry: CountryCode = 'PE';

export const DocumentTypes = {
  DNI: { code: 'DNI', name: 'DNI', country: 'PE', length: 8 },
  RUC: { code: 'RUC', name: 'RUC', country: 'PE', length: 11 },
  CE: { code: 'CE', name: 'Carnet de Extranjería', country: 'PE', length: 12 },
  PASSPORT: { code: 'PASSPORT', name: 'Pasaporte', country: '*', length: 12 },
} as const;

export type DocumentTypeCode = keyof typeof DocumentTypes;
