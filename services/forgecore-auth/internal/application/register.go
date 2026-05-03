package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/crypto"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type RegisterInput struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (i RegisterInput) Validate() error {
	if i.Email == "" {
		return fmt.Errorf("email obbligatoria")
	}
	if len(i.Password) < 8 {
		return fmt.Errorf("password minimo 8 caratteri")
	}
	return nil
}

type RegisterOutput struct {
	UserID uuid.UUID
}

type RegisterUseCase struct {
	users     domain.UserRepository
	publisher EventPublisher
	encryptor *crypto.PIIEncryptor
}

func NewRegisterUseCase(users domain.UserRepository, pub EventPublisher, enc *crypto.PIIEncryptor) *RegisterUseCase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &RegisterUseCase{users: users, publisher: pub, encryptor: enc}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	emailHash := uc.encryptor.Hash(input.Email)
	existing, err := uc.users.GetByEmailHash(ctx, emailHash, input.TenantID)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("verifica email esistente fallita: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	emailEnc, err := uc.encryptor.Encrypt(input.Email)
	if err != nil {
		return nil, fmt.Errorf("cifratura email fallita: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password fallito: %w", err)
	}
	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     input.TenantID,
		EmailEnc:     []byte(emailEnc),
		EmailHash:    emailHash,
		PasswordHash: string(hash),
		Roles:        []string{"user"},
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creazione utente fallita: %w", err)
	}
	_ = uc.publisher.Publish(ctx, events.SubjectUserRegistered, events.UserRegistered{
		TenantID:   user.TenantID,
		UserID:     user.ID,
		Email:      input.Email,
		OccurredAt: time.Now().UTC(),
	})
	return &RegisterOutput{UserID: user.ID}, nil
}
