package bindings

import (
	"time"

	"vfinancy/backend/internal/domain/repositories"
)

type ProfileDTO struct {
	UserID        string `json:"userId"`
	FullName      string `json:"fullName"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	AvatarURL     string `json:"avatarUrl"`
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	DateFormat    string `json:"dateFormat"`
	NumberFormat  string `json:"numberFormat"`
	DecimalPlaces int    `json:"decimalPlaces"`
	Timezone      string `json:"timezone"`
	LastLoginAt   string `json:"lastLoginAt"`
	IsActive      bool   `json:"isActive"`
}

type AuditEventDTO struct {
	ID          string `json:"id"`
	EventType   string `json:"eventType"`
	UserID      string `json:"userId"`
	Description string `json:"description"`
	IPAddress   string `json:"ipAddress"`
	Device      string `json:"device"`
	OccurredAt  string `json:"occurredAt"`
}

func (a *App) GetProfile(sessionToken string) (*ProfileDTO, error) {
	ctx := a.Context()

	session, err := a.sessionSvc.Validate(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	user, err := a.repos.Users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	profile, err := a.profileSvc.GetByUserID(ctx, session.UserID)
	if err != nil {
		profile, err = a.profileSvc.EnsureProfile(ctx, session.UserID)
		if err != nil {
			return nil, err
		}
	}

	dto := &ProfileDTO{
		UserID:        user.ID.String(),
		FullName:      user.FullName.String(),
		Username:      user.Username.String(),
		Email:         user.Email.String(),
		AvatarURL:     profile.AvatarURL,
		Theme:         profile.Theme,
		Language:      profile.Language,
		DateFormat:    profile.DateFormat,
		NumberFormat:  profile.NumberFormat,
		DecimalPlaces: profile.DecimalPlaces,
		Timezone:      profile.Timezone,
		IsActive:      user.IsActive,
	}

	if user.LastLoginAt != nil {
		dto.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	return dto, nil
}

func (a *App) UpdateProfile(sessionToken string, dto ProfileDTO) error {
	ctx := a.Context()

	session, err := a.sessionSvc.Validate(ctx, sessionToken)
	if err != nil {
		return err
	}

	profile, err := a.profileSvc.GetByUserID(ctx, session.UserID)
	if err != nil {
		return err
	}

	if dto.Theme != "" {
		if err := profile.ChangeTheme(dto.Theme); err != nil {
			return err
		}
	}
	if dto.Language != "" {
		if err := profile.ChangeLanguage(dto.Language); err != nil {
			return err
		}
	}
	if dto.DateFormat != "" {
		if err := profile.ChangeDateFormat(dto.DateFormat); err != nil {
			return err
		}
	}
	if dto.NumberFormat != "" {
		if err := profile.ChangeNumberFormat(dto.NumberFormat); err != nil {
			return err
		}
	}
	if dto.DecimalPlaces > 0 {
		if err := profile.ChangeDecimalPlaces(dto.DecimalPlaces); err != nil {
			return err
		}
	}
	if dto.Timezone != "" {
		if err := profile.ChangeTimezone(dto.Timezone); err != nil {
			return err
		}
	}
	if dto.AvatarURL != "" {
		profile.SetAvatar(dto.AvatarURL)
	}

	return a.profileSvc.Update(ctx, profile)
}

func (a *App) GetAuditLog(page, pageSize int, eventType string) ([]AuditEventDTO, int, error) {
	ctx := a.Context()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filter := repositories.AuditEventFilter{
		EventType: eventType,
		PageRequest: repositories.PageRequest{
			Limit:  pageSize,
			Offset: offset,
		},
	}

	events, total, err := a.auditSvc.List(ctx, demoCompanyID, filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]AuditEventDTO, len(events))
	for i, e := range events {
		dto := AuditEventDTO{
			ID:          e.ID.String(),
			EventType:   string(e.EventType),
			Description: e.Description,
			IPAddress:   e.IPAddress,
			Device:      e.Device,
			OccurredAt:  e.OccurredAt.Format(time.RFC3339),
		}
		if e.UserID != nil {
			dto.UserID = e.UserID.String()
		}
		result[i] = dto
	}

	return result, total, nil
}
