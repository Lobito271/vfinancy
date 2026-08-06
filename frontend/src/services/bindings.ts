import type {
  AppBindings,
  AppSettingsDTO,
  BusinessInfoDTO,
  ConnectionConfigDTO,
  ProfileDTO,
} from './wails-types';
import { mockBindings } from './mock-bindings';

let resolved: AppBindings | null = null;

function isWailsRuntime(): boolean {
  return typeof window !== 'undefined' && Boolean((window as { go?: unknown }).go);
}

async function resolveBindings(): Promise<AppBindings> {
  if (resolved) return resolved;
  if (!isWailsRuntime()) {
    resolved = mockBindings;
  } else {
    const mod = await import('../../wailsjs/go/bindings/App');
    resolved = mod as unknown as AppBindings;
  }
  return resolved;
}

export const wailsClient = {
  async login(username: string, password: string, remember: boolean) {
    const b = await resolveBindings();
    return b.Login({ username, password, remember });
  },
  async logout(token: string) {
    const b = await resolveBindings();
    return b.Logout(token);
  },
  async changePassword(currentPassword: string, newPassword: string, token: string) {
    const b = await resolveBindings();
    return b.ChangePassword(currentPassword, newPassword, token);
  },
  async validateSession(token: string) {
    const b = await resolveBindings();
    return b.ValidateSession(token);
  },
  async getBusinessInfo() {
    const b = await resolveBindings();
    return b.GetBusinessInfo();
  },
  async updateBusinessInfo(info: BusinessInfoDTO) {
    const b = await resolveBindings();
    return b.UpdateBusinessInfo(info);
  },
  async getPreferences() {
    const b = await resolveBindings();
    return b.GetPreferences();
  },
  async updatePreference(key: string, value: string) {
    const b = await resolveBindings();
    return b.UpdatePreference(key, value);
  },
  async getCurrencies() {
    const b = await resolveBindings();
    return b.GetCurrencies();
  },
  async getTaxes() {
    const b = await resolveBindings();
    return b.GetTaxes();
  },
  async getAllSettings() {
    const b = await resolveBindings();
    return b.GetAllSettings();
  },
  async getProfile(token: string) {
    const b = await resolveBindings();
    return b.GetProfile(token);
  },
  async updateProfile(token: string, profile: ProfileDTO) {
    const b = await resolveBindings();
    return b.UpdateProfile(token, profile);
  },
  async getAuditLog(page: number, pageSize: number, eventType: string) {
    const b = await resolveBindings();
    return b.GetAuditLog(page, pageSize, eventType);
  },
  async getConnectionConfig() {
    const b = await resolveBindings();
    return b.GetConnectionConfig();
  },
  async testDatabaseConnection(cfg: ConnectionConfigDTO) {
    const b = await resolveBindings();
    return b.TestDatabaseConnection(cfg);
  },
  async saveConnectionConfig(cfg: ConnectionConfigDTO) {
    const b = await resolveBindings();
    return b.SaveConnectionConfig(cfg);
  },
  async getAppSettings() {
    const b = await resolveBindings();
    return b.GetAppSettings();
  },
  async saveAppSettings(settings: AppSettingsDTO) {
    const b = await resolveBindings();
    return b.SaveAppSettings(settings);
  },
  async getModules() {
    const b = await resolveBindings();
    return b.GetModules();
  },
  async setModuleEnabled(id: string, enabled: boolean) {
    const b = await resolveBindings();
    return b.SetModuleEnabled(id, enabled);
  },
};
