import type { BackupResult, Preferences, SyncConfig } from '../wails-types';
import { wailsClient } from '../bindings';

export const settingsService = {
  async getPreferences(): Promise<Preferences> {
    return wailsClient.getPreferences();
  },
  async updatePreference(key: string, value: number | string): Promise<Preferences> {
    return wailsClient.updatePreference(key, value);
  },
  async getSyncConfig(): Promise<SyncConfig> {
    return wailsClient.getSyncConfig();
  },
  async saveSyncConfig(cfg: SyncConfig): Promise<void> {
    await wailsClient.saveSyncConfig(cfg);
  },
  async testSyncConnection(cfg: SyncConfig): Promise<void> {
    await wailsClient.testSyncConnection(cfg);
  },
  async createBackup(): Promise<BackupResult> {
    return wailsClient.createBackup();
  },
};
