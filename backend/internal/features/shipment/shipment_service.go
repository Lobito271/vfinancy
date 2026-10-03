// Package shipment implements the business logic for the Envios
// (shipments) module. Every shipment belongs to one sale; the customer
// is derived from it. The code and the security code are assigned at
// creation and cannot be edited later.
package shipment

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/infrastructure/logger"
	"vfinancy/backend/internal/domain/enums"
	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/repositories"
	"vfinancy/backend/internal/features/sales"
)

// saleGetter is the narrow sales contract consumed by the shipment
// slice to resolve the customer of the originating sale.
type saleGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*sales.Sale, error)
}

// ShipmentService owns the shipment slice.
type ShipmentService struct {
	repo  ShipmentRepository
	txm   repositories.TransactionManager
	sales saleGetter
	log   *logger.Logger
}

// New returns a ShipmentService ready for use.
func New(repo ShipmentRepository, txm repositories.TransactionManager, log *logger.Logger) *ShipmentService {
	return &ShipmentService{repo: repo, txm: txm, log: log}
}

// SetSales installs the sales lookup used to derive the customer of the
// originating sale. Without it Create rejects the shipment.
func (s *ShipmentService) SetSales(g saleGetter) {
	if g != nil {
		s.sales = g
	}
}

// CreateInput is the payload for Create.
type CreateInput struct {
	SaleID       uuid.UUID
	SecurityCode string
	Location     string
	ShipmentDate *time.Time
	DeliveredAt  *time.Time
	Description  string
	Notes        string
}

// Create validates the input, assigns the next shipment code and
// persists the shipment in a single transaction. The customer is taken
// from the originating sale.
func (s *ShipmentService) Create(ctx context.Context, in CreateInput) (*Shipment, error) {
	if in.SaleID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("shipment must be linked to a sale"))
	}
	if s.sales == nil {
		return nil, derrors.New("INTERNAL", "sales are not configured")
	}
	sale, err := s.sales.GetByID(ctx, in.SaleID)
	if err != nil {
		return nil, err
	}
	var out *Shipment
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		code, err := s.repo.NextCode(ctx)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		shp := &Shipment{
			ID:           uuid.New(),
			Code:         code,
			SecurityCode: strings.TrimSpace(in.SecurityCode),
			SaleID:       sale.ID,
			CustomerID:   &sale.CustomerID,
			Location:     strings.TrimSpace(in.Location),
			ShipmentDate: in.ShipmentDate,
			DeliveredAt:  in.DeliveredAt,
			Description:  strings.TrimSpace(in.Description),
			Notes:        in.Notes,
			Status:       enums.ShipmentStatusPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := shp.Validate(true); err != nil {
			return err
		}
		if err := s.repo.Create(ctx, shp); err != nil {
			return err
		}
		shp.SaleNumber = sale.Number
		out = shp
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("shipment created",
		"shipment_id", out.ID, "code", out.Code, "sale_id", out.SaleID)
	return out, nil
}

// UpdateInput is the payload for Update. Pointer fields keep the
// current value when nil. Code, SecurityCode and SaleID are immutable
// and are never updated.
type UpdateInput struct {
	ID           uuid.UUID
	Location     *string
	ShipmentDate **time.Time
	DeliveredAt  **time.Time
	Description  *string
	Notes        *string
	Status       *enums.ShipmentStatus
}

// Update applies the requested changes in a single transaction.
func (s *ShipmentService) Update(ctx context.Context, in UpdateInput) (*Shipment, error) {
	if in.ID == uuid.Nil {
		return nil, derrors.Wrap(derrors.ErrRequired, errField("id is required"))
	}
	var out *Shipment
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		shp, err := s.repo.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if in.Location != nil {
			shp.Location = strings.TrimSpace(*in.Location)
		}
		if in.ShipmentDate != nil {
			shp.ShipmentDate = *in.ShipmentDate
		}
		if in.DeliveredAt != nil {
			shp.DeliveredAt = *in.DeliveredAt
		}
		if in.Description != nil {
			shp.Description = strings.TrimSpace(*in.Description)
		}
		if in.Notes != nil {
			shp.Notes = *in.Notes
		}
		if in.Status != nil {
			if !in.Status.Valid() {
				return derrors.Wrap(derrors.ErrInvalidEnum, errField("status is invalid"))
			}
			shp.Status = *in.Status
		}
		if err := shp.Validate(false); err != nil {
			return err
		}
		shp.UpdatedAt = time.Now().UTC()
		if err := s.repo.Update(ctx, shp); err != nil {
			return err
		}
		out = shp
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.log.Info("shipment updated", "shipment_id", in.ID)
	return out, nil
}

// Delete soft-deletes the shipment. Historical records are preserved.
func (s *ShipmentService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.log.Info("shipment deleted", "shipment_id", id)
	return nil
}

// GetByID returns a single shipment.
func (s *ShipmentService) GetByID(ctx context.Context, id uuid.UUID) (*Shipment, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a page of shipments matching the filter.
func (s *ShipmentService) List(ctx context.Context, filter ShipmentFilter) (repositories.Page[*Shipment], error) {
	return s.repo.List(ctx, filter)
}
