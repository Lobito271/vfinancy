/// <reference types="vite/client" />

declare module 'virtual:wails-bindings' {
  export interface LoginRequest {
    username: string;
    password: string;
    remember: boolean;
  }

  export interface LoginResponse {
    token: string;
    expiresAt: string;
    userId: string;
    fullName: string;
    email: string;
    username: string;
    roles: string[];
    companyId: string;
    mustChangePassword: boolean;
  }

  export interface BusinessInfoDTO {
    name: string;
    tradeName: string;
    taxId: string;
    address: string;
    phone: string;
    email: string;
    logo: string;
  }

  export interface PreferencesDTO {
    defaultCurrency: string;
    defaultTaxCode: string;
    expiryAlertDays: number;
    defaultCountry: string;
    dateFormat: string;
    numberFormat: string;
    decimalPlaces: number;
    language: string;
    theme: string;
    timezone: string;
    fiscalYearStart: number;
    backupFolder: string;
    exportFolder: string;
    backupFrequency: string;
  }

  export interface CurrencyDTO {
    code: string;
    symbol: string;
    name: string;
    decimalPlaces: number;
    type: string;
    isActive: boolean;
  }

  export interface TaxDTO {
    id: string;
    code: string;
    name: string;
    shortName: string;
    countryCode: string;
    defaultRate: number;
    isInclusive: boolean;
    isPercentage: boolean;
    category: string;
    isActive: boolean;
  }

  export interface ProfileDTO {
    userId: string;
    fullName: string;
    username: string;
    email: string;
    avatarUrl: string;
    theme: string;
    language: string;
    dateFormat: string;
    numberFormat: string;
    decimalPlaces: number;
    timezone: string;
    lastLoginAt: string;
    isActive: boolean;
  }

  export interface AuditEventDTO {
    id: string;
    eventType: string;
    userId: string;
    description: string;
    ipAddress: string;
    device: string;
    occurredAt: string;
  }

  export interface AppBindings {
    Login(req: LoginRequest): Promise<LoginResponse>;
    Logout(sessionToken: string): Promise<void>;
    ChangePassword(currentPassword: string, newPassword: string, sessionToken: string): Promise<void>;
    ValidateSession(sessionToken: string): Promise<boolean>;

    GetBusinessInfo(): Promise<BusinessInfoDTO>;
    UpdateBusinessInfo(info: BusinessInfoDTO): Promise<void>;
    GetPreferences(): Promise<PreferencesDTO>;
    UpdatePreference(key: string, value: string): Promise<void>;
    GetCurrencies(): Promise<CurrencyDTO[]>;
    GetTaxes(): Promise<TaxDTO[]>;
    GetAllSettings(): Promise<Record<string, unknown>>;

    GetProfile(sessionToken: string): Promise<ProfileDTO>;
    UpdateProfile(sessionToken: string, profile: ProfileDTO): Promise<void>;
    GetAuditLog(page: number, pageSize: number, eventType: string): Promise<[AuditEventDTO[], number]>;
  }

  export const App: AppBindings;
}

declare module '*.svg' {
  const src: string;
  export default src;
}
declare module '*.png' {
  const src: string;
  export default src;
}
