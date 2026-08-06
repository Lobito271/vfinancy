package masterdata

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Warehouse is a physical storage location. Each warehouse has a
// default address and a manager (user).
type Warehouse struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	BranchID    uuid.UUID
	Code        valueobjects.ShortCode
	Name        valueobjects.FullName
	Address     valueobjects.Address
	ManagerID   *uuid.UUID
	IsDefault   bool
	AllowsClearance bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewWarehouseOptions is the input to NewWarehouse.
type NewWarehouseOptions struct {
	CompanyID       uuid.UUID
	BranchID        uuid.UUID
	Code            valueobjects.ShortCode
	Name            valueobjects.FullName
	Address         valueobjects.Address
	ManagerID       *uuid.UUID
	IsDefault       bool
	AllowsClearance bool
}

// NewWarehouse validates and constructs a warehouse.
func NewWarehouse(now time.Time, opts NewWarehouseOptions) (*Warehouse, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if opts.BranchID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("branch id is required"))
	}
	return &Warehouse{
		ID:              uuid.New(),
		CompanyID:       opts.CompanyID,
		BranchID:        opts.BranchID,
		Code:            opts.Code,
		Name:            opts.Name,
		Address:         opts.Address,
		ManagerID:       opts.ManagerID,
		IsDefault:       opts.IsDefault,
		AllowsClearance: opts.AllowsClearance,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Activate / Deactivate.
func (w *Warehouse) Activate()   { w.IsActive = true }
func (w *Warehouse) Deactivate() { w.IsActive = false }

// SetManager assigns or clears the manager. Pass nil to clear.
func (w *Warehouse) SetManager(userID *uuid.UUID) { w.ManagerID = userID }

// Rename updates the display name.
func (w *Warehouse) Rename(name valueobjects.FullName) { w.Name = name }

// SetAllowsClearance toggles whether the warehouse accepts clearance
// (post-25-day) batches for sale.
func (w *Warehouse) SetAllowsClearance(v bool) { w.AllowsClearance = v }

// DefaultCategoryLabel is used in reports; kept here for tests.
const DefaultCategoryLabel = "Sin categoría"
