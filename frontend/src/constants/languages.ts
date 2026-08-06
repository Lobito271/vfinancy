export const Languages = {
  'es-PE': { code: 'es-PE', name: 'Español (Perú)', nativeName: 'Español', flag: 'PE', rtl: false },
  'es-MX': { code: 'es-MX', name: 'Español (México)', nativeName: 'Español', flag: 'MX', rtl: false },
  'es-CO': { code: 'es-CO', name: 'Español (Colombia)', nativeName: 'Español', flag: 'CO', rtl: false },
  'es-AR': { code: 'es-AR', name: 'Español (Argentina)', nativeName: 'Español', flag: 'AR', rtl: false },
  'en-US': { code: 'en-US', name: 'English (United States)', nativeName: 'English', flag: 'US', rtl: false },
  'pt-BR': { code: 'pt-BR', name: 'Português (Brasil)', nativeName: 'Português', flag: 'BR', rtl: false },
} as const;

export type LanguageCode = keyof typeof Languages;
export const DefaultLanguage: LanguageCode = 'es-PE';
