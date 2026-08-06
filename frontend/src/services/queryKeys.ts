export const queryKeys = {
  customers: {
    all: ['customers'] as const,
    list: (q: unknown) => ['customers', 'list', q] as const,
    detail: (id: string) => ['customers', 'detail', id] as const,
    options: ['customers', 'options'] as const,
  },
  suppliers: {
    all: ['suppliers'] as const,
    list: (q: unknown) => ['suppliers', 'list', q] as const,
    options: ['suppliers', 'options'] as const,
  },
  products: {
    all: ['products'] as const,
    list: (q: unknown) => ['products', 'list', q] as const,
    options: ['products', 'options'] as const,
  },
  sales: { all: ['sales'] as const, list: ['sales', 'list'] as const },
  inventory: {
    all: ['inventory'] as const,
    list: ['inventory', 'list'] as const,
    clearance: ['inventory', 'clearance'] as const,
    lowStock: ['inventory', 'lowStock'] as const,
  },
  treasury: {
    all: ['treasury'] as const,
    accounts: ['treasury', 'accounts'] as const,
    transactions: (id?: string) => ['treasury', 'transactions', id] as const,
    exchangeRate: (from: string, to: string) => ['treasury', 'exchange', from, to] as const,
  },
  accounting: {
    all: ['accounting'] as const,
    chart: ['accounting', 'chart'] as const,
    entries: ['accounting', 'entries'] as const,
  },
  reports: { all: ['reports'] as const, list: ['reports', 'list'] as const },
  administration: { all: ['administration'] as const, users: ['administration', 'users'] as const, audit: ['administration', 'audit'] as const },
  settings: {
    business: ['settings', 'business'] as const,
    preferences: ['settings', 'preferences'] as const,
    currencies: ['settings', 'currencies'] as const,
    taxes: ['settings', 'taxes'] as const,
    all: ['settings', 'all'] as const,
  },
  profile: (token: string) => ['profile', token] as const,
  auditLog: (page: number, eventType: string) => ['auditLog', page, eventType] as const,
} as const;
