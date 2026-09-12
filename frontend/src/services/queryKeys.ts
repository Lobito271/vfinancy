export const queryKeys = {
  setup: ['setup'] as const,
  auth: {
    state: ['auth', 'state'] as const,
  },
  customers: {
    all: ['customers'] as const,
    list: (q: unknown) => ['customers', 'list', q] as const,
    options: ['customers', 'options'] as const,
    document: (docType: string, docNumber: string) => ['customers', 'document', docType, docNumber] as const,
  },
  products: {
    all: ['products'] as const,
    list: (q: unknown) => ['products', 'list', q] as const,
    options: ['products', 'options'] as const,
    stock: (id: string) => ['products', 'stock', id] as const,
  },
  sales: {
    all: ['sales'] as const,
    list: (q: unknown) => ['sales', 'list', q] as const,
    detail: (id: string) => ['sales', 'detail', id] as const,
    payments: (q: unknown) => ['sales', 'payments', q] as const,
  },
  purchasing: {
    all: ['purchasing'] as const,
    list: (q: unknown) => ['purchasing', 'list', q] as const,
    detail: (id: string) => ['purchasing', 'detail', id] as const,
  },
  importLots: {
    all: ['importLots'] as const,
    list: (q: unknown) => ['importLots', 'list', q] as const,
    detail: (id: string) => ['importLots', 'detail', id] as const,
    members: (id: string) => ['importLots', 'members', id] as const,
  },
  inventory: {
    all: ['inventory'] as const,
    list: (q: unknown) => ['inventory', 'list', q] as const,
    movements: (productId?: string) => ['inventory', 'movements', productId] as const,
    clearance: ['inventory', 'clearance'] as const,
  },
  treasury: {
    all: ['treasury'] as const,
    creditCards: ['treasury', 'creditCards'] as const,
    cardProjections: ['treasury', 'cardProjections'] as const,
    exchangeRate: (from: string, to: string) => ['treasury', 'exchange', from, to] as const,
  },
  settings: {
    preferences: ['settings', 'preferences'] as const,
    sync: ['settings', 'sync'] as const,
    backup: ['settings', 'backup'] as const,
  },
  dashboard: {
    overview: ['dashboard', 'overview'] as const,
  },
} as const;
