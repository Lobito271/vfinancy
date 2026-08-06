package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	derrors "vfinancy/backend/internal/domain/errors"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultParams() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func HashPassword(password string, params *Argon2Params) (string, error) {
	if params == nil {
		params = DefaultParams()
	}

	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory,
		params.Iterations,
		params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("auth: invalid hash format")
	}
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("auth: unsupported algorithm %q", parts[1])
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("auth: parsing params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("auth: decoding salt: %w", err)
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("auth: decoding hash: %w", err)
	}

	keyLen := uint32(len(expected))
	computed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLen)

	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return derrors.Wrap(ErrPasswordWeak, errors.New("password must be at least 8 characters"))
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	if !hasUpper {
		return derrors.Wrap(ErrPasswordWeak, errors.New("password must contain at least one uppercase letter"))
	}
	if !hasLower {
		return derrors.Wrap(ErrPasswordWeak, errors.New("password must contain at least one lowercase letter"))
	}
	if !hasDigit {
		return derrors.Wrap(ErrPasswordWeak, errors.New("password must contain at least one digit"))
	}

	return nil
}
