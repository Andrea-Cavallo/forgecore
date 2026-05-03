package domain

import "github.com/google/uuid"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type TokenClaims struct {
	UserID      uuid.UUID
	TenantID    uuid.UUID
	Roles       []string
	JTI         string
	MFAVerified bool
}
