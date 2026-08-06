import type {
  AppBindings,
  AuditLogResult,
  BusinessInfoDTO,
  ConnectionConfigDTO,
  ModuleDTO,
  PreferencesDTO,
  ProfileDTO,
} from './wails-types';

function delay(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

export const mockModules: ModuleDTO[] = [
  { id: 'dashboard', name: 'Inicio', description: 'Panel principal con indicadores', enabled: true },
  { id: 'customers', name: 'Clientes', description: 'Gestión de clientes y cartera', enabled: true },
  { id: 'suppliers', name: 'Proveedores', description: 'Gestión de proveedores', enabled: true },
  { id: 'products', name: 'Productos', description: 'Catálogo de productos', enabled: true },
  { id: 'inventory', name: 'Inventario', description: 'Stock, lotes y movimientos', enabled: true },
  { id: 'purchasing', name: 'Compras', description: 'Órdenes de compra y pagos', enabled: true },
  { id: 'sales', name: 'Ventas', description: 'Ventas, pagos y adelantos', enabled: true },
  { id: 'treasury', name: 'Tesorería', description: 'Cuentas bancarias y transacciones', enabled: true },
  { id: 'accounting', name: 'Contabilidad', description: 'Plan contable y asientos', enabled: true },
  { id: 'reports', name: 'Reportes', description: 'Reportes gerenciales', enabled: true },
  { id: 'settings', name: 'Configuración', description: 'Preferencias y conexión', enabled: true },
  { id: 'administration', name: 'Administración', description: 'Usuarios y auditoría', enabled: true },
];

export const mockBindings: AppBindings = {
  async Login(req) {
    await delay(500);
    if (!req.username || !req.password) throw new Error('Credenciales invalidas');
    if (req.username === 'admin' && req.password === 'admin123') {
      return {
        token: 'mock-session-token-' + Date.now(),
        expiresAt: new Date(Date.now() + 3600000).toISOString(),
        userId: '00000000-0000-0000-0000-0000000000aa',
        fullName: 'Administrador del Sistema',
        email: 'admin@vfinancy.local',
        username: 'admin',
        roles: ['admin'],
        companyId: '00000000-0000-0000-0000-000000000001',
        mustChangePassword: false,
      };
    }
    throw new Error('Usuario o contrasena incorrectos');
  },
  async Logout() {
    await delay(200);
  },
  async ChangePassword() {
    await delay(300);
  },
  async ValidateSession() {
    await delay(100);
    return true;
  },
  async GetBusinessInfo(): Promise<BusinessInfoDTO> {
    await delay(150);
    return { name: 'vfinancy S.A.C.', tradeName: 'vfinancy', taxId: '20600000001', address: 'Lima, Peru', phone: '', email: 'admin@vfinancy.local', logo: '' };
  },
  async UpdateBusinessInfo() {
    await delay(200);
  },
  async GetPreferences(): Promise<PreferencesDTO> {
    await delay(150);
    return {
      defaultCurrency: 'PEN', defaultTaxCode: 'IGV', expiryAlertDays: 25, defaultCountry: 'PE',
      dateFormat: 'DD/MM/YYYY', numberFormat: 'es-PE', decimalPlaces: 2, language: 'es-PE',
      theme: 'light', timezone: 'America/Lima', fiscalYearStart: 1, backupFolder: '', exportFolder: '', backupFrequency: 'daily',
    };
  },
  async UpdatePreference() {
    await delay(200);
  },
  async GetCurrencies() {
    await delay(100);
    return [
      { code: 'PEN', symbol: 'S/', name: 'Sol peruano', decimalPlaces: 2, type: 'fiat', isActive: true },
      { code: 'USD', symbol: '$', name: 'Dolar estadounidense', decimalPlaces: 2, type: 'fiat', isActive: true },
    ];
  },
  async GetTaxes() {
    await delay(100);
    return [
      { id: '1', code: 'IGV', name: 'Impuesto General a las Ventas', shortName: 'IGV', countryCode: 'PE', defaultRate: 0.18, isInclusive: false, isPercentage: true, category: 'sales', isActive: true },
      { id: '2', code: 'IVAP', name: 'Impuesto de Promocion Municipal', shortName: 'IVAP', countryCode: 'PE', defaultRate: 0.02, isInclusive: false, isPercentage: true, category: 'municipal', isActive: true },
      { id: '3', code: 'RENTA', name: 'Impuesto a la Renta', shortName: 'Renta', countryCode: 'PE', defaultRate: 0.295, isInclusive: false, isPercentage: true, category: 'income', isActive: true },
      { id: '4', code: 'EXONERADO', name: 'Exonerado del IGV', shortName: 'Exo', countryCode: 'PE', defaultRate: 0, isInclusive: false, isPercentage: true, category: 'sales', isActive: true },
    ];
  },
  async GetAllSettings() {
    await delay(150);
    return {};
  },
  async GetProfile(): Promise<ProfileDTO> {
    await delay(150);
    return { userId: '00000000-0000-0000-0000-0000000000aa', fullName: 'Administrador del Sistema', username: 'admin', email: 'admin@vfinancy.local', avatarUrl: '', theme: 'system', language: 'es-PE', dateFormat: 'DD/MM/YYYY', numberFormat: 'es-PE', decimalPlaces: 2, timezone: 'America/Lima', lastLoginAt: new Date().toISOString(), isActive: true };
  },
  async UpdateProfile() {
    await delay(200);
  },
  async GetAuditLog(): Promise<AuditLogResult> {
    await delay(150);
    return {
      events: [
        { id: '1', eventType: 'LOGIN', userId: 'admin', description: 'Inicio de sesion exitoso', ipAddress: '127.0.0.1', device: 'Desktop', occurredAt: new Date().toISOString() },
        { id: '2', eventType: 'CONFIG_UPDATE', userId: 'admin', description: 'Actualizacion de preferencia: theme', ipAddress: '127.0.0.1', device: 'Desktop', occurredAt: new Date(Date.now() - 3600000).toISOString() },
      ],
      total: 2,
    };
  },
  async GetConnectionConfig(): Promise<ConnectionConfigDTO> {
    await delay(100);
    return { host: 'localhost', port: 5432, user: 'postgres', password: 'casa123', database: 'vfinancy', sslMode: 'disable' };
  },
  async TestDatabaseConnection(cfg: ConnectionConfigDTO) {
    await delay(400);
    if (!cfg.host || !cfg.database) return 'Host y base de datos son requeridos';
    return '';
  },
  async SaveConnectionConfig() {
    await delay(300);
  },
  async GetAppSettings() {
    await delay(100);
    return { windowTitle: 'vfinancy', width: 1280, height: 800, logLevel: 'info', logFormat: 'json' };
  },
  async SaveAppSettings() {
    await delay(300);
  },
  async GetModules() {
    await delay(100);
    return mockModules;
  },
  async SetModuleEnabled(id, enabled) {
    await delay(150);
    const m = mockModules.find((x) => x.id === id);
    if (m) m.enabled = enabled;
  },
};
