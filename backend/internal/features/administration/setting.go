package administration

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ApplicationSetting is one device-local preference row stored as a
// JSON value under a unique key.
type ApplicationSetting struct {
	ID        uuid.UUID
	Key       string
	Value     json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewApplicationSetting builds a setting with a fresh ID and UTC timestamps.
func NewApplicationSetting(key string, value json.RawMessage) *ApplicationSetting {
	now := time.Now().UTC()
	return &ApplicationSetting{
		ID:        uuid.New(),
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Touch stamps UpdatedAt with the current UTC time.
func (s *ApplicationSetting) Touch() {
	s.UpdatedAt = time.Now().UTC()
}

// StringValue decodes the stored JSON value as a string, or "" when it
// is not a JSON string.
func (s *ApplicationSetting) StringValue() string {
	var v string
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return ""
	}
	return v
}

// IntValue decodes the stored JSON value as an integer, or 0 when it
// is not a JSON number.
func (s *ApplicationSetting) IntValue() int {
	var v int
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return 0
	}
	return v
}

// Float64Value decodes the stored JSON value as a float, or 0 when it
// is not a JSON number.
func (s *ApplicationSetting) Float64Value() float64 {
	var v float64
	if err := json.Unmarshal(s.Value, &v); err != nil {
		return 0
	}
	return v
}
