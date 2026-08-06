package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	derrors "vfinancy/backend/internal/domain/errors"
	"vfinancy/backend/internal/shared/logger"
)

type ProfileService struct {
	profiles ProfileRepository
	users    UserRepository
	log      *logger.Logger
}

func NewProfileService(
	profiles ProfileRepository,
	users UserRepository,
	log *logger.Logger,
) *ProfileService {
	if profiles == nil {
		panic("auth: nil profiles repository")
	}
	if users == nil {
		panic("auth: nil users repository")
	}
	if log == nil {
		panic("auth: nil logger")
	}
	return &ProfileService{
		profiles: profiles,
		users:    users,
		log:      log,
	}
}

func (s *ProfileService) GetByUserID(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	if userID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "user id is required")
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

func (s *ProfileService) Create(ctx context.Context, profile *UserProfile) error {
	if profile == nil {
		return derrors.New("REQUIRED", "profile is required")
	}

	if err := s.profiles.Create(ctx, profile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile created", "user_id", profile.UserID, "profile_id", profile.ID)
	return nil
}

func (s *ProfileService) Update(ctx context.Context, profile *UserProfile) error {
	if profile == nil {
		return derrors.New("REQUIRED", "profile is required")
	}

	if err := s.profiles.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile updated", "user_id", profile.UserID, "profile_id", profile.ID)
	return nil
}

func (s *ProfileService) ChangeTheme(ctx context.Context, userID uuid.UUID, theme string) error {
	if userID == uuid.Nil {
		return derrors.New("REQUIRED", "user id is required")
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.ChangeTheme(theme); err != nil {
		return fmt.Errorf("failed to change theme: %w", err)
	}

	if err := s.profiles.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.log.InfoContext(ctx, "theme changed", "user_id", userID, "theme", theme)
	return nil
}

func (s *ProfileService) ChangeLanguage(ctx context.Context, userID uuid.UUID, language string) error {
	if userID == uuid.Nil {
		return derrors.New("REQUIRED", "user id is required")
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.ChangeLanguage(language); err != nil {
		return fmt.Errorf("failed to change language: %w", err)
	}

	if err := s.profiles.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.log.InfoContext(ctx, "language changed", "user_id", userID, "language", language)
	return nil
}

func (s *ProfileService) EnsureProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	if userID == uuid.Nil {
		return nil, derrors.New("REQUIRED", "user id is required")
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err == nil {
		return profile, nil
	}

	exists, err := s.users.Exists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, derrors.New("NOT_FOUND", "user does not exist")
	}

	profile, err = NewUserProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	if err := s.profiles.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to persist profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile created with defaults", "user_id", userID, "profile_id", profile.ID)
	return profile, nil
}
