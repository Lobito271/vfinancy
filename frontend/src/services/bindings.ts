import type { AppBindings } from 'virtual:wails-bindings';

let bindings: AppBindings | null = null;

async function getBindings(): Promise<AppBindings> {
  if (bindings) return bindings;
  try {
    const mod = await import('virtual:wails-bindings');
    bindings = mod.App;
    return bindings;
  } catch {
    return getMockBindings();
  }
}

function getMockBindings(): AppBindings {
  return {
    Login: async (req) => {
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
    Logout: async () => { await delay(200); },
    ChangePassword: async (_cur, _new, _tok) => { await delay(300); },
    ValidateSession: async () => { await delay(100); return true; },
    GetBusinessInfo: async () => {
      await delay(150);
      return { name: 'vfinancy S.A.C.', tradeName: 'vfinancy', taxId: '20600000001', address: 'Lima, Peru', phone: '', email: 'admin@vfinancy.local', logo: '' };
    },
    UpdateBusinessInfo: async () => { await delay(200); },
    GetPreferences: async () => {
      await delay(150);
      return {
        defaultCurrency: 'PEN', defaultTaxCode: 'IGV', expiryAlertDays: 25, defaultCountry: 'PE',
        dateFormat: 'DD/MM/YYYY', numberFormat: 'es-PE', decimalPlaces: 2, language: 'es-PE',
        theme: 'light', timezone: 'America/Lima', fiscalYearStart: 1, backupFolder: '', exportFolder: '', backupFrequency: 'daily',
      };
    },
    UpdatePreference: async () => { await delay(200); },
    GetCurrencies: async () => {
      await delay(100);
      return [
        { code: 'PEN', symbol: 'S/', name: 'Sol peruano', decimalPlaces: 2, type: 'fiat', isActive: true },
        { code: 'USD', symbol: '$', name: 'Dolar estadounidense', decimalPlaces: 2, type: 'fiat', isActive: true },
      ];
    },
    GetTaxes: async () => {
      await delay(100);
      return [
        { id: '1', code: 'IGV', name: 'Impuesto General a las Ventas', shortName: 'IGV', countryCode: 'PE', defaultRate: 0.18, isInclusive: false, isPercentage: true, category: 'sales', isActive: true },
        { id: '2', code: 'IVAP', name: 'Impuesto de Promocion Municipal', shortName: 'IVAP', countryCode: 'PE', defaultRate: 0.02, isInclusive: false, isPercentage: true, category: 'municipal', isActive: true },
        { id: '3', code: 'RENTA', name: 'Impuesto a la Renta', shortName: 'Renta', countryCode: 'PE', defaultRate: 0.295, isInclusive: false, isPercentage: true, category: 'income', isActive: true },
        { id: '4', code: 'EXONERADO', name: 'Exonerado del IGV', shortName: 'Exo', countryCode: 'PE', defaultRate: 0, isInclusive: false, isPercentage: true, category: 'sales', isActive: true },
      ];
    },
    GetAllSettings: async () => { await delay(150); return {}; },
    GetProfile: async () => {
      await delay(150);
      return { userId: '00000000-0000-0000-0000-0000000000aa', fullName: 'Administrador del Sistema', username: 'admin', email: 'admin@vfinancy.local', avatarUrl: '', theme: 'system', language: 'es-PE', dateFormat: 'DD/MM/YYYY', numberFormat: 'es-PE', decimalPlaces: 2, timezone: 'America/Lima', lastLoginAt: new Date().toISOString(), isActive: true };
    },
    UpdateProfile: async () => { await delay(200); },
    GetAuditLog: async () => {
      await delay(150);
      return [[
        { id: '1', eventType: 'LOGIN', userId: 'admin', description: 'Inicio de sesion exitoso', ipAddress: '127.0.0.1', device: 'Desktop', occurredAt: new Date().toISOString() },
        { id: '2', eventType: 'CONFIG_UPDATE', userId: 'admin', description: 'Actualizacion de preferencia: theme', ipAddress: '127.0.0.1', device: 'Desktop', occurredAt: new Date(Date.now() - 3600000).toISOString() },
      ], 2] as [import('virtual:wails-bindings').AuditEventDTO[], number];
    },
  };
}

function delay(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

export const wailsClient = {
  async login(username: string, password: string, remember: boolean) {
    const b = await getBindings();
    return b.Login({ username, password, remember });
  },
  async logout(token: string) {
    const b = await getBindings();
    return b.Logout(token);
  },
  async changePassword(currentPassword: string, newPassword: string, token: string) {
    const b = await getBindings();
    return b.ChangePassword(currentPassword, newPassword, token);
  },
  async validateSession(token: string) {
    const b = await getBindings();
    return b.ValidateSession(token);
  },
  async getBusinessInfo() {
    const b = await getBindings();
    return b.GetBusinessInfo();
  },
  async updateBusinessInfo(info: import('virtual:wails-bindings').BusinessInfoDTO) {
    const b = await getBindings();
    return b.UpdateBusinessInfo(info);
  },
  async getPreferences() {
    const b = await getBindings();
    return b.GetPreferences();
  },
  async updatePreference(key: string, value: string) {
    const b = await getBindings();
    return b.UpdatePreference(key, value);
  },
  async getCurrencies() {
    const b = await getBindings();
    return b.GetCurrencies();
  },
  async getTaxes() {
    const b = await getBindings();
    return b.GetTaxes();
  },
  async getAllSettings() {
    const b = await getBindings();
    return b.GetAllSettings();
  },
  async getProfile(token: string) {
    const b = await getBindings();
    return b.GetProfile(token);
  },
  async updateProfile(token: string, profile: import('virtual:wails-bindings').ProfileDTO) {
    const b = await getBindings();
    return b.UpdateProfile(token, profile);
  },
  async getAuditLog(page: number, pageSize: number, eventType: string) {
    const b = await getBindings();
    return b.GetAuditLog(page, pageSize, eventType);
  },
};
