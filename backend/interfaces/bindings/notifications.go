package bindings

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/features/notifications"
)

// NotificationDTO is the serializable view of a device-local
// notification.
type NotificationDTO struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	RecordType string `json:"recordType"`
	RecordID   string `json:"recordId"`
	IsRead     bool   `json:"isRead"`
	ReadAt     string `json:"readAt"`
	CreatedAt  string `json:"createdAt"`
}

func toNotificationDTO(n *notifications.Notification) *NotificationDTO {
	readAt := ""
	if n.ReadAt != nil {
		readAt = n.ReadAt.Format(time.RFC3339)
	}
	recordID := ""
	if n.RecordID != nil {
		recordID = n.RecordID.String()
	}
	return &NotificationDTO{
		ID:         n.ID.String(),
		Type:       n.Type,
		Title:      n.Title,
		Message:    n.Message,
		RecordType: n.RecordType,
		RecordID:   recordID,
		IsRead:     n.ReadAt != nil,
		ReadAt:     readAt,
		CreatedAt:  n.CreatedAt.Format(time.RFC3339),
	}
}

// ListNotificationsRequest filters the notification feed.
type ListNotificationsRequest struct {
	OnlyUnread bool `json:"onlyUnread"`
	PaginationRequest
}

// ListNotifications returns the notification feed of the active
// company, newest first.
func (a *App) ListNotifications(req ListNotificationsRequest) (PageResult, error) {
	page, err := a.notificationsSvc.List(a.Context(), notifications.ListFilter{
		CompanyID:   a.companyID(),
		UnreadOnly:  req.OnlyUnread,
		PageRequest: req.toPageRequest(),
	})
	if err != nil {
		return PageResult{}, err
	}
	items := make([]*NotificationDTO, 0, len(page.Items))
	for _, n := range page.Items {
		items = append(items, toNotificationDTO(n))
	}
	return PageResult{Items: items, Total: page.Total, Page: page.Offset/page.Limit + 1, PageSize: page.Limit}, nil
}

// UnreadNotificationCount returns how many notifications of the active
// company are still unread (bell badge).
func (a *App) UnreadNotificationCount() (int, error) {
	return a.notificationsSvc.UnreadCount(a.Context(), a.companyID())
}

// MarkNotificationsRead marks the given notifications as read.
func (a *App) MarkNotificationsRead(ids []string) error {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		uuidID, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		parsed = append(parsed, uuidID)
	}
	_, err := a.notificationsSvc.MarkRead(a.Context(), a.companyID(), parsed)
	return err
}

// MarkAllNotificationsRead marks every notification of the active
// company as read.
func (a *App) MarkAllNotificationsRead() error {
	_, err := a.notificationsSvc.MarkAllRead(a.Context(), a.companyID())
	return err
}

// DeleteNotification removes a single notification.
func (a *App) DeleteNotification(id string) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return a.notificationsSvc.Delete(a.Context(), a.companyID(), uuidID)
}

// GenerateClearanceNotifications runs the clearance scan on demand and
// returns how many new notifications were created.
func (a *App) GenerateClearanceNotifications() (int, error) {
	return a.notificationsSvc.Generate(a.Context())
}
