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
	"golang.org/x/crypto/bcrypt"
)

const (
	resetTokenTTL  int64 = 3600       // 1 ora
	resetKeyPrefix       = "pwreset:"
)

var ErrPasswordResetTokenInvalid = errors.New("token di reset non valido o scaduto")

// ForgotPasswordUseCase sends a password reset email.
// Always returns 200 OK to prevent user enumeration.
type ForgotPasswordUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenStore
	enc    *crypto.PIIEncryptor
	pub    EventPublisher
}

type ForgotPasswordInput struct {
	TenantID uuid.UUID
	Email    string
}

func NewForgotPasswordUseCase(
	users domain.UserRepository,
	tokens domain.TokenStore,
	enc *crypto.PIIEncryptor,
	pub EventPublisher,
) *ForgotPasswordUseCase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &ForgotPasswordUseCase{users: users, tokens: tokens, enc: enc, pub: pub}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, in ForgotPasswordInput) error {
	emailHash := uc.enc.Hash(in.Email)
	user, err := uc.users.GetByEmailHash(ctx, emailHash, in.TenantID)
	if err != nil {
		// Non rivela se l'email esiste o meno
		return nil
	}
	token := uuid.New().String()
	key := resetKeyPrefix + user.ID.String()
	if err := uc.tokens.StoreOneTimeToken(ctx, key, token, resetTokenTTL); err != nil {
		return fmt.Errorf("salva token reset: %w", err)
	}
	resetURL := fmt.Sprintf("/v1/auth/password-reset/confirm?token=%s&user_id=%s", token, user.ID)
	if err := uc.pub.Publish(ctx, events.SubjectPasswordReset, events.PasswordReset{
		TenantID:   in.TenantID,
		UserID:     user.ID,
		Email:      in.Email,
		ResetURL:   resetURL,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("pubblica evento reset: %w", err)
	}
	return nil
}

// ResetPasswordUseCase validates the token and sets a new password.
type ResetPasswordUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenStore
}

type ResetPasswordInput struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Token       string
	NewPassword string
}

func (i ResetPasswordInput) Validate() error {
	if i.Token == "" {
		return fmt.Errorf("token obbligatorio")
	}
	if len(i.NewPassword) < 8 {
		return fmt.Errorf("la password deve avere almeno 8 caratteri")
	}
	return nil
}

func NewResetPasswordUseCase(users domain.UserRepository, tokens domain.TokenStore) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{users: users, tokens: tokens}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, in ResetPasswordInput) error {
	if err := in.Validate(); err != nil {
		return err
	}
	key := resetKeyPrefix + in.UserID.String()
	stored, err := uc.tokens.PopOneTimeToken(ctx, key)
	if err != nil || stored != in.Token {
		return ErrPasswordResetTokenInvalid
	}
	user, err := uc.users.GetByID(ctx, in.UserID, in.TenantID)
	if err != nil {
		return fmt.Errorf("carica utente: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now().UTC()
	return uc.users.Update(ctx, user)
}
