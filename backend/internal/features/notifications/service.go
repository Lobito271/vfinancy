package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/domain/valueobjects"
	"vfinancy/backend/internal/features/inventory"
	"vfinancy/backend/internal/shared/logger"
)

// NotificationsService owns the device-local notification feed. Its
// only generator today is the inventory clearance rule, but the feed
// (type column + dedup) is ready for future alert kinds (low stock,
// payments) without schema changes.
type NotificationsService struct {
	repo          Repository
	clearance     ClearanceSource
	clearanceDays func(context.Context, uuid.UUID) int
	productInfo   func(context.Context, uuid.UUID) (description, sku string, err error)
	activeCompany func() uuid.UUID
	log           *logger.Logger
}

// ClearanceSource yields the batches currently on clearance for a
// company. The inventory service implements it.
type ClearanceSource interface {
	GenerateClearanceCandidates(ctx context.Context, companyID uuid.UUID, at time.Time) ([]*inventory.InventoryBatch, error)
}

func New(repo Repository, log *logger.Logger) *NotificationsService {
	return &NotificationsService{repo: repo, log: log}
}

// SetClearanceSource wires the generator to the inventory clearance
// candidates.
func (s *NotificationsService) SetClearanceSource(src ClearanceSource) {
	s.clearance = src
}

// SetClearanceDays wires the maximum-sale-date window used to build
// the notification text (defaults to the clearance rule when unset).
func (s *NotificationsService) SetClearanceDays(days func(context.Context, uuid.UUID) int) {
	s.clearanceDays = days
}

// SetProductInfo wires the product name/sku resolver used to render
// the clearance message.
func (s *NotificationsService) SetProductInfo(resolve func(context.Context, uuid.UUID) (description, sku string, err error)) {
	s.productInfo = resolve
}

// SetActiveCompany wires the active-company resolver (uuid.Nil when
// the workspace is not configured). Generation is a no-op then.
func (s *NotificationsService) SetActiveCompany(resolve func() uuid.UUID) {
	s.activeCompany = resolve
}

func (s *NotificationsService) clearanceWindow(ctx context.Context, companyID uuid.UUID) int {
	if s.clearanceDays != nil {
		if days := s.clearanceDays(ctx, companyID); days > 0 {
			return days
		}
	}
	return inventory.ClearanceDays
}

// Generate scans the inventory for clearance batches and appends a
// notification per batch (idempotent: the (company_id, type,
// dedup_key) key suppresses duplicates). Notifications for batches
// that left clearance since the last scan are soft-deleted so the
// feed stays truthful. Returns the number of new notifications.
func (s *NotificationsService) Generate(ctx context.Context) (int, error) {
	companyID := s.activeCompany()
	if companyID == uuid.Nil {
		return 0, nil
	}
	now := time.Now().UTC()
	today := valueobjects.Date(now)

	batches, err := s.clearance.GenerateClearanceCandidates(ctx, companyID, now)
	if err != nil {
		return 0, err
	}

	keep := make([]string, 0, len(batches))
	notes := make([]*Notification, 0, len(batches))
	days := s.clearanceWindow(ctx, companyID)
	names := map[uuid.UUID][2]string{}
	for _, b := range batches {
		keep = append(keep, b.ID.String())
		info, ok := names[b.ProductID]
		if !ok {
			description, sku, err := s.productInfo(ctx, b.ProductID)
			if err != nil {
				return 0, err
			}
			info = [2]string{description, sku}
			names[b.ProductID] = info
		}
		notes = append(notes, NewClearanceNotification(companyID, b, days, today, info[0], info[1]))
	}

	created, err := s.repo.CreateBatch(ctx, notes)
	if err != nil {
		return 0, err
	}
	if _, err := s.repo.DeleteStaleClearance(ctx, companyID, keep); err != nil {
		return 0, err
	}
	return created, nil
}

// List returns the notification feed, newest first.
func (s *NotificationsService) List(ctx context.Context, filter ListFilter) (repositories.Page[*Notification], error) {
	return s.repo.List(ctx, filter)
}

// UnreadCount returns the number of unread notifications.
func (s *NotificationsService) UnreadCount(ctx context.Context, companyID uuid.UUID) (int, error) {
	return s.repo.CountUnread(ctx, companyID)
}

// MarkRead marks the given notifications read.
func (s *NotificationsService) MarkRead(ctx context.Context, companyID uuid.UUID, ids []uuid.UUID) (int, error) {
	if companyID == uuid.Nil || len(ids) == 0 {
		return 0, nil
	}
	return s.repo.MarkRead(ctx, companyID, ids)
}

// MarkAllRead marks every notification of the company read.
func (s *NotificationsService) MarkAllRead(ctx context.Context, companyID uuid.UUID) (int, error) {
	if companyID == uuid.Nil {
		return 0, nil
	}
	return s.repo.MarkAllRead(ctx, companyID)
}

// Delete soft-deletes a single notification.
func (s *NotificationsService) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	if companyID == uuid.Nil {
		return nil
	}
	return s.repo.Delete(ctx, companyID, id)
}

// NewClearanceNotification renders the alert for a batch that passed
// its maximum sale date. The message is data (like audit entries), so
// it is built here in Spanish rather than through the i18n helper.
func NewClearanceNotification(companyID uuid.UUID, b *inventory.InventoryBatch, days int, today valueobjects.Date, description, sku string) *Notification {
	const layout = "2006-01-02"
	max := b.MaximumSaleDateAfter(days)
	msg := fmt.Sprintf(
		"%s (%s) lleva %d días en el almacén; ingresó el %s y debía venderse antes del %s.",
		description, sku, b.DaysInStock(today), b.ArrivalDate.Format(layout), max.Format(layout),
	)
	return &Notification{
		CompanyID:  companyID,
		Type:       TypeClearance,
		Title:      "Producto en remate",
		Message:    msg,
		RecordType: RecordTypeBatch,
		RecordID:   &b.ID,
		DedupKey:   b.ID.String(),
	}
}
