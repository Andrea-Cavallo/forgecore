package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/auth-service/internal/domain"
	"github.com/yourorg/golang-modules/shared/crypto"
	"github.com/yourorg/golang-modules/shared/events"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	IPAddress string
	UserAgent string
	DeviceID  string
}

type LoginOutput struct {
	Tokens domain.TokenPair
}

type LoginUseCase struct {
	users     domain.UserRepository
	sessions  domain.SessionRepository
	tokens    domain.TokenStore
	publisher *events.Publisher
	encryptor *crypto.PIIEncryptor
	jwtIssuer TokenIssuer
}

type TokenIssuer interface {
	Issue(claims domain.TokenClaims) (domain.TokenPair, error)
}

func NewLoginUseCase(users domain.UserRepository, sessions domain.SessionRepository, tokens domain.TokenStore, pub *events.Publisher, enc *crypto.PIIEncryptor, issuer TokenIssuer) *LoginUseCase {
	return &LoginUseCase{users: users, sessions: sessions, tokens: tokens, publisher: pub, encryptor: enc, jwtIssuer: issuer}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	emailHash := crypto.Hash(input.Email)
	user, err := uc.users.GetByEmailHash(ctx, emailHash, input.TenantID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.IsDeleted() || user.IsLocked(time.Now().UTC()) {
		return nil, domain.ErrAccountLocked
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	pair, err := uc.jwtIssuer.Issue(domain.TokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Roles:    user.Roles,
		JTI:      uuid.New().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("emissione token fallita: %w", err)
	}
	session := &domain.Session{
		ID:         uuid.New(),
		TenantID:   input.TenantID,
		UserID:     user.ID,
		DeviceID:   input.DeviceID,
		UserAgent:  input.UserAgent,
		IPAddress:  input.IPAddress,
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(7 * 24 * time.Hour),
	}
	_ = uc.sessions.Create(ctx, session)
	_ = uc.publisher.Publish(ctx, events.SubjectUserLogin, events.UserLogin{
		TenantID:   user.TenantID,
		UserID:     user.ID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: time.Now().UTC(),
	})
	return &LoginOutput{Tokens: pair}, nil
}
