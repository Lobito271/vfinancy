import { wailsClient } from '../bindings';

type BusinessInfo = Awaited<ReturnType<typeof wailsClient.getBusinessInfo>>;
type Preferences = Awaited<ReturnType<typeof wailsClient.getPreferences>>;
type Currency = Awaited<ReturnType<typeof wailsClient.getCurrencies>>[number];
type Tax = Awaited<ReturnType<typeof wailsClient.getTaxes>>[number];

export const settingsService = {
  async getBusinessInfo(): Promise<BusinessInfo> {
    return wailsClient.getBusinessInfo();
  },
  async updateBusinessInfo(info: BusinessInfo): Promise<void> {
    await wailsClient.updateBusinessInfo(info);
  },
  async getPreferences(): Promise<Preferences> {
    return wailsClient.getPreferences();
  },
  async updatePreference(key: string, value: string): Promise<void> {
    await wailsClient.updatePreference(key, value);
  },
  async getCurrencies(): Promise<Currency[]> {
    return wailsClient.getCurrencies();
  },
  async getTaxes(): Promise<Tax[]> {
    return wailsClient.getTaxes();
  },
  async getAllSettings(): Promise<Record<string, unknown>> {
    return wailsClient.getAllSettings();
  },
};
