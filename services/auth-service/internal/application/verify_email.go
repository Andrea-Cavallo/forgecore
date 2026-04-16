package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/crypto"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

const (
	verifyTokenTTL  int64 = 86400      // 24 ore
	verifyKeyPrefix       = "emailver:"
)

var (
	ErrEmailVerifyTokenInvalid = errors.New("token di verifica non valido o scaduto")
	ErrEmailAlreadyVerified    = errors.New("email già verificata")
)

// VerifyEmailUseCase confirms the user's email address.
type VerifyEmailUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenStore
}

type VerifyEmailInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Token    string
}

func NewVerifyEmailUseCase(users domain.UserRepository, tokens domain.TokenStore) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{users: users, tokens: tokens}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, in VerifyEmailInput) error {
	if in.Token == "" {
		return fmt.Errorf("token obbligatorio")
	}
	key := verifyKeyPrefix + in.UserID.String()
	stored, err := uc.tokens.PopOneTimeToken(ctx, key)
	if err != nil || stored != in.Token {
		return ErrEmailVerifyTokenInvalid
	}
	user, err := uc.users.GetByID(ctx, in.UserID, in.TenantID)
	if err != nil {
		return fmt.Errorf("carica utente: %w", err)
	}
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}
	user.EmailVerified = true
	user.UpdatedAt = time.Now().UTC()
	return uc.users.Update(ctx, user)
}

// ResendVerificationUseCase sends a new verification email.
type ResendVerificationUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenStore
	enc    *crypto.PIIEncryptor
	pub    EventPublisher
}

type ResendVerificationInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
}

func NewResendVerificationUseCase(
	users domain.UserRepository,
	tokens domain.TokenStore,
	enc *crypto.PIIEncryptor,
	pub EventPublisher,
) *ResendVerificationUseCase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &ResendVerificationUseCase{users: users, tokens: tokens, enc: enc, pub: pub}
}

func (uc *ResendVerificationUseCase) Execute(ctx context.Context, in ResendVerificationInput) error {
	user, err := uc.users.GetByID(ctx, in.UserID, in.TenantID)
	if err != nil {
		return fmt.Errorf("carica utente: %w", err)
	}
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}
	token := uuid.New().String()
	key := verifyKeyPrefix + user.ID.String()
	if err := uc.tokens.StoreOneTimeToken(ctx, key, token, verifyTokenTTL); err != nil {
		return fmt.Errorf("salva token verifica: %w", err)
	}
	verifyURL := fmt.Sprintf("/v1/auth/email/verify?token=%s&user_id=%s", token, user.ID)
	decrypted, decErr := uc.enc.Decrypt(string(user.EmailEnc))
	if decErr != nil {
		decrypted = "" // non blocca — l'evento porta comunque la URL
	}
	if err := uc.pub.Publish(ctx, events.SubjectEmailVerified, events.UserRegistered{
		TenantID:   in.TenantID,
		UserID:     user.ID,
		Email:      decrypted,
		VerifyURL:  verifyURL,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("pubblica evento verifica: %w", err)
	}
	return nil
}
