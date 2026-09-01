import type { NotificationDTO } from '../wails-types';
import { wailsClient } from '../bindings';

export interface AppNotification {
  id: string;
  type: string;
  title: string;
  message: string;
  recordType: string;
  recordId: string;
  isRead: boolean;
  createdAt: string;
}

function toNotification(dto: NotificationDTO): AppNotification {
  return {
    id: dto.id,
    type: dto.type,
    title: dto.title,
    message: dto.message,
    recordType: dto.recordType,
    recordId: dto.recordId,
    isRead: dto.isRead,
    createdAt: dto.createdAt,
  };
}

export const notificationsService = {
  async list(onlyUnread = false): Promise<AppNotification[]> {
    const res = await wailsClient.listNotifications({ onlyUnread, page: 1, pageSize: 50 });
    return res.items.map(toNotification);
  },
  async unreadCount(): Promise<number> {
    return wailsClient.unreadNotificationCount();
  },
  async markRead(ids: string[]): Promise<void> {
    await wailsClient.markNotificationsRead(ids);
  },
  async markAllRead(): Promise<void> {
    await wailsClient.markAllNotificationsRead();
  },
  async remove(id: string): Promise<void> {
    await wailsClient.deleteNotification(id);
  },
  async generateClearance(): Promise<number> {
    return wailsClient.generateClearanceNotifications();
  },
};
