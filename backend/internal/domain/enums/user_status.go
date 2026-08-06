package enums

// UserStatus is the lifecycle state of a system user.
type UserStatus string

const (
	// UserStatusActive — can authenticate, no lockout.
	UserStatusActive UserStatus = "active"
	// UserStatusInactive — administratively disabled.
	UserStatusInactive UserStatus = "inactive"
	// UserStatusLocked — temporarily locked out by failed-attempt policy.
	UserStatusLocked UserStatus = "locked"
)

func (u UserStatus) Valid() bool {
	switch u {
	case UserStatusActive, UserStatusInactive, UserStatusLocked:
		return true
	}
	return false
}

func (u UserStatus) String() string { return string(u) }
