package auth

import (
	"time"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/errors"
)

type UserProfile struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	AvatarURL     string
	Theme         string
	Language      string
	DateFormat    string
	NumberFormat  string
	DecimalPlaces int
	Timezone      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUserProfile(userID uuid.UUID) (*UserProfile, error) {
	if userID == uuid.Nil {
		return nil, errors.Wrap(errors.ErrRequired, errField("user id is required"))
	}
	now := time.Now()
	return &UserProfile{
		ID:            uuid.New(),
		UserID:        userID,
		Theme:         "system",
		Language:      "es-PE",
		DateFormat:    "DD/MM/YYYY",
		NumberFormat:  "es-PE",
		DecimalPlaces: 2,
		Timezone:      "America/Lima",
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (p *UserProfile) ChangeTheme(theme string) error {
	switch theme {
	case "light", "dark", "system":
		p.Theme = theme
		return nil
	default:
		return errors.Wrap(errors.ErrInvalidEnum, errField("theme must be light, dark, or system"))
	}
}

func (p *UserProfile) ChangeLanguage(lang string) error {
	if lang == "" {
		return errors.Wrap(errors.ErrRequired, errField("language is required"))
	}
	p.Language = lang
	return nil
}

func (p *UserProfile) ChangeDateFormat(format string) error {
	if format == "" {
		return errors.Wrap(errors.ErrRequired, errField("date format is required"))
	}
	p.DateFormat = format
	return nil
}

func (p *UserProfile) ChangeNumberFormat(format string) error {
	if format == "" {
		return errors.Wrap(errors.ErrRequired, errField("number format is required"))
	}
	p.NumberFormat = format
	return nil
}

func (p *UserProfile) ChangeDecimalPlaces(places int) error {
	if places < 0 || places > 6 {
		return errors.Wrap(errors.ErrOutOfRange, errField("decimal places must be 0..6"))
	}
	p.DecimalPlaces = places
	return nil
}

func (p *UserProfile) ChangeTimezone(tz string) error {
	if tz == "" {
		return errors.Wrap(errors.ErrRequired, errField("timezone is required"))
	}
	p.Timezone = tz
	return nil
}

func (p *UserProfile) SetAvatar(url string) {
	p.AvatarURL = url
}
