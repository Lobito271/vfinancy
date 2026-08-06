package auth

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/domain/valueobjects"
)

// Branch is a physical location under a Company. Users carry a
// default branch and may be granted branch-scoped roles.
type Branch struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Code        valueobjects.ShortCode
	Name        valueobjects.FullName
	Address     valueobjects.Address
	Phone       valueobjects.Phone
	IsDefault   bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
}

// NewBranchOptions is the input to NewBranch.
type NewBranchOptions struct {
	CompanyID uuid.UUID
	Code      valueobjects.ShortCode
	Name      valueobjects.FullName
	Address   valueobjects.Address  // optional
	Phone     valueobjects.Phone    // optional
	IsDefault bool
}

// NewBranch constructs a Branch. Code is required and unique per
// company (enforced at the database level).
func NewBranch(now time.Time, opts NewBranchOptions) (*Branch, error) {
	if opts.CompanyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	return &Branch{
		ID:        uuid.New(),
		CompanyID: opts.CompanyID,
		Code:      opts.Code,
		Name:      opts.Name,
		Address:   opts.Address,
		Phone:     opts.Phone,
		IsDefault: opts.IsDefault,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Activate / Deactivate.
func (b *Branch) Activate()    { b.IsActive = true }
func (b *Branch) Deactivate()  { b.IsActive = false }

// MarkAsDefault flags this branch as the default for its company.
// At most one branch per company can be the default; the database
// enforces this with a partial unique index, and the application
// layer is expected to coordinate the flip atomically.
func (b *Branch) MarkAsDefault() { b.IsDefault = true }

// ClearDefault removes the default flag.
func (b *Branch) ClearDefault() { b.IsDefault = false }

// UpdateContact updates the address and phone. Either field may be
// the zero value of its value-object to mean "unchanged" / "blank".
func (b *Branch) UpdateContact(addr valueobjects.Address, phone valueobjects.Phone) {
	b.Address = addr
	b.Phone = phone
}

// SoftDelete marks the branch as deleted. The branch is preserved
// for historical reporting.
func (b *Branch) SoftDelete(at time.Time, by uuid.UUID) {
	now := at
	b.DeletedAt = &now
	b.UpdatedAt = at
	b.UpdatedBy = &by
	b.IsActive = false
}

func (b *Branch) IsDeleted() bool { return b.DeletedAt != nil }
