import { wailsClient } from '../bindings';

export type AuditEvent = Awaited<ReturnType<typeof wailsClient.getAuditLog>>['events'][number];
export type UserProfile = Awaited<ReturnType<typeof wailsClient.getProfile>>;

export const administrationService = {
  async getProfile(token: string): Promise<UserProfile> {
    return wailsClient.getProfile(token);
  },
  async updateProfile(token: string, profile: UserProfile): Promise<void> {
    await wailsClient.updateProfile(token, profile);
  },
  async getAuditLog(page: number = 1, pageSize: number = 20, eventType: string = ''): Promise<{ events: AuditEvent[]; total: number }> {
    return wailsClient.getAuditLog(page, pageSize, eventType);
  },
};
