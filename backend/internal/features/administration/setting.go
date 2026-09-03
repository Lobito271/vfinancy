package administration

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
)

type ApplicationSetting struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	Key         string
	Value       json.RawMessage
	Category    string
	Label       string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
}

func NewApplicationSetting(companyID uuid.UUID, key, category, label, description string, value json.RawMessage, isPublic bool) (*ApplicationSetting, error) {
	if companyID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("company id is required"))
	}
	if key == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("key is required"))
	}
	if category == "" {
		return nil, errors.Wrap(errors.ErrRequired, errField("category is required"))
	}
	now := time.Now()
	return &ApplicationSetting{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Key:         key,
		Value:       value,
		Category:    category,
		Label:       label,
		Description: description,
		IsPublic:    isPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *ApplicationSetting) Update(value json.RawMessage, updatedBy uuid.UUID) {
	s.Value = value
	s.UpdatedAt = time.Now()
	s.UpdatedBy = &updatedBy
}

func (s *ApplicationSetting) StringValue() string {
	var v string
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return ""
	}
	return v
}

func (s *ApplicationSetting) IntValue() int {
	var v int
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return 0
	}
	return v
}

func (s *ApplicationSetting) BoolValue() bool {
	var v bool
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return false
	}
	return v
}

func (s *ApplicationSetting) Float64Value() float64 {
	var v float64
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return 0
	}
	return v
}
