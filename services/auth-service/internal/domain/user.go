package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	EmailEnc      []byte
	EmailHash     string
	PasswordHash  string
	Roles         []string
	MFAEnabled    bool
	MFASecret     []byte
	EmailVerified bool
	LockedUntil   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && now.Before(*u.LockedUntil)
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}
