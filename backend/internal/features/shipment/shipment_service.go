// Package shipment implements the business logic for the Envios
// (shipments) module. The service owns the CRUD; the four-digit code
// is assigned at creation and cannot be edited later.
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
)

// ShipmentService owns the shipment slice.
type ShipmentService struct {
	repo ShipmentRepository
	txm  repositories.TransactionManager
	log  *logger.Logger
}

// New returns a ShipmentService ready for use.
func New(repo ShipmentRepository, txm repositories.TransactionManager, log *logger.Logger) *ShipmentService {
	return &ShipmentService{repo: repo, txm: txm, log: log}
}

// CreateInput is the payload for Create.
type CreateInput struct {
	CustomerID  *uuid.UUID
	SaleID      *uuid.UUID
	Description string
	Notes       string
}

// Create validates the input, assigns the next four-digit code and
// persists the shipment in a single transaction.
func (s *ShipmentService) Create(ctx context.Context, in CreateInput) (*Shipment, error) {
	code, err := s.repo.NextCode(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	shp := &Shipment{
		ID:          uuid.New(),
		Code:        code,
		CustomerID:  in.CustomerID,
		SaleID:      in.SaleID,
		Description: strings.TrimSpace(in.Description),
		Notes:       in.Notes,
		Status:      enums.ShipmentStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := shp.Validate(true); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, shp); err != nil {
		return nil, err
	}
	s.log.Info("shipment created", "shipment_id", shp.ID, "code", shp.Code)
	return shp, nil
}

// UpdateInput is the payload for Update. Pointer fields keep the
// current value when nil; the code is immutable and is never updated.
type UpdateInput struct {
	ID          uuid.UUID
	CustomerID  *uuid.UUID
	SaleID      *uuid.UUID
	Description *string
	Notes       *string
	Status      *enums.ShipmentStatus
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
		if in.CustomerID != nil {
			shp.CustomerID = in.CustomerID
		}
		if in.SaleID != nil {
			shp.SaleID = in.SaleID
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
