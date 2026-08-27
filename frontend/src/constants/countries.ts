export const DocumentTypes = {
  DNI: { code: 'DNI', name: 'DNI', country: 'PE', length: 8 },
  RUC: { code: 'RUC', name: 'RUC', country: 'PE', length: 11 },
  CE: { code: 'CE', name: 'Carnet de Extranjería', country: 'PE', length: 12 },
  PASSPORT: { code: 'PASSPORT', name: 'Pasaporte', country: '*', length: 12 },
} as const;
