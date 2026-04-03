package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/shared/pagination"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*User, error)
	GetByEmailHash(ctx context.Context, emailHash string, tenantID uuid.UUID) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Session, error)
	DeleteByUserID(ctx context.Context, userID, tenantID uuid.UUID) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
}

type TokenStore interface {
	StoreRefreshToken(ctx context.Context, key, token string, ttlSeconds int64) error
	ValidateRefreshToken(ctx context.Context, key, token string) (bool, error)
	BlacklistJTI(ctx context.Context, jti string, ttlSeconds int64) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
	IncrBruteForce(ctx context.Context, key string) (int64, error)
	SetBruteForceLockout(ctx context.Context, key string, ttlSeconds int64) error
	GetBruteForceCount(ctx context.Context, key string) (int64, error)
}
