package administration

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"vfinancy/backend/internal/application/services/common"
	"vfinancy/backend/internal/domain/entities/identity"
	"vfinancy/backend/internal/domain/repositories"
)

type ProfileService struct {
	profiles repositories.ProfileRepository
	users    repositories.UserRepository
	log      *common.Logger
}

func NewProfileService(
	profiles repositories.ProfileRepository,
	users repositories.UserRepository,
	log *common.Logger,
) *ProfileService {
	if profiles == nil {
		panic("administration: nil profiles repository")
	}
	if users == nil {
		panic("administration: nil users repository")
	}
	if log == nil {
		panic("administration: nil logger")
	}
	return &ProfileService{
		profiles: profiles,
		users:    users,
		log:      log,
	}
}

func (s *ProfileService) GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.UserProfile, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("REQUIRED: user id is required")
	}

	profile, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

func (s *ProfileService) Create(ctx context.Context, profile *identity.UserProfile) error {
	if profile == nil {
		return fmt.Errorf("REQUIRED: profile is required")
	}

	if err := s.profiles.Create(ctx, profile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile created", "user_id", profile.UserID, "profile_id", profile.ID)
	return nil
}

func (s *ProfileService) Update(ctx context.Context, profile *identity.UserProfile) error {
	if profile == nil {
		return fmt.Errorf("REQUIRED: profile is required")
	}

	if err := s.profiles.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile updated", "user_id", profile.UserID, "profile_id", profile.ID)
	return nil
}

func (s *ProfileService) ChangeTheme(ctx context.Context, userID uuid.UUID, theme string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("REQUIRED: user id is required")
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
		return fmt.Errorf("REQUIRED: user id is required")
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

func (s *ProfileService) EnsureProfile(ctx context.Context, userID uuid.UUID) (*identity.UserProfile, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("REQUIRED: user id is required")
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
		return nil, fmt.Errorf("NOT_FOUND: user does not exist")
	}

	profile, err = identity.NewUserProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	if err := s.profiles.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to persist profile: %w", err)
	}

	s.log.InfoContext(ctx, "profile created with defaults", "user_id", userID, "profile_id", profile.ID)
	return profile, nil
}
