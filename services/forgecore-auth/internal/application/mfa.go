package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
)

const (
	mfaIssuer        = "ForgeCore"
	mfaBackupCodeLen = 8
	mfaBackupCount   = 8
)

var (
	ErrMFAAlreadyEnabled  = errors.New("MFA già abilitato per questo account")
	ErrMFANotEnabled      = errors.New("MFA non abilitato per questo account")
	ErrMFAInvalidCode     = errors.New("codice MFA non valido")
	ErrMFANoBackupCodesLeft = errors.New("nessun codice di backup disponibile")
)

// EnableMFAInput contains the user context for MFA setup.
type EnableMFAInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
}

// EnableMFAOutput contains the TOTP provisioning details.
type EnableMFAOutput struct {
	Secret      string
	QRCodeURL   string
	BackupCodes []string
}

// EnableMFAUseCase generates a TOTP secret and backup codes for a user.
type EnableMFAUseCase struct {
	users  domain.UserRepository
	events EventPublisher
}

func NewEnableMFAUseCase(users domain.UserRepository, events EventPublisher) *EnableMFAUseCase {
	return &EnableMFAUseCase{users: users, events: events}
}

func (uc *EnableMFAUseCase) Execute(ctx context.Context, in EnableMFAInput) (*EnableMFAOutput, error) {
	user, err := uc.users.GetByID(ctx, in.UserID, in.TenantID)
	if err != nil {
		return nil, fmt.Errorf("carica utente: %w", err)
	}
	if user.MFAEnabled {
		return nil, ErrMFAAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      mfaIssuer,
		AccountName: in.UserID.String(),
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("genera chiave TOTP: %w", err)
	}

	backupCodes, err := generateBackupCodes(mfaBackupCount)
	if err != nil {
		return nil, fmt.Errorf("genera backup codes: %w", err)
	}

	// Store secret (unencrypted here; encrypt in production via PIIEncryptor)
	user.MFASecret = []byte(key.Secret())
	user.MFABackupCodes = backupCodes
	user.UpdatedAt = time.Now().UTC()

	if err := uc.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("salva segreto MFA: %w", err)
	}

	return &EnableMFAOutput{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyMFAInput contains the code to verify during MFA setup or login.
type VerifyMFAInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Code     string
}

// VerifyMFAOutput contains the JWT after successful MFA.
type VerifyMFAOutput struct {
	Tokens domain.TokenPair
}

// VerifyMFAUseCase validates the TOTP code and activates MFA (or completes login).
type VerifyMFAUseCase struct {
	users  domain.UserRepository
	issuer TokenIssuer
	events EventPublisher
}

func NewVerifyMFAUseCase(users domain.UserRepository, issuer TokenIssuer, events EventPublisher) *VerifyMFAUseCase {
	return &VerifyMFAUseCase{users: users, issuer: issuer, events: events}
}

func (uc *VerifyMFAUseCase) Execute(ctx context.Context, in VerifyMFAInput) (*VerifyMFAOutput, error) {
	user, err := uc.users.GetByID(ctx, in.UserID, in.TenantID)
	if err != nil {
		return nil, fmt.Errorf("carica utente: %w", err)
	}
	if len(user.MFASecret) == 0 {
		return nil, ErrMFANotEnabled
	}

	valid := totp.Validate(in.Code, string(user.MFASecret))
	if !valid {
		// Try backup code
		used, remaining, ok := consumeBackupCode(user.MFABackupCodes, in.Code)
		if !ok {
			return nil, ErrMFAInvalidCode
		}
		_ = used
		user.MFABackupCodes = remaining
		user.UpdatedAt = time.Now().UTC()
		if err := uc.users.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("aggiorna backup codes: %w", err)
		}
	}

	if !user.MFAEnabled {
		user.MFAEnabled = true
		user.UpdatedAt = time.Now().UTC()
		if err := uc.users.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("attiva MFA: %w", err)
		}
	}

	pair, err := uc.issuer.Issue(domain.TokenClaims{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Roles:       user.Roles,
		JTI:         uuid.New().String(),
		MFAVerified: true,
	})
	if err != nil {
		return nil, fmt.Errorf("emetti token: %w", err)
	}

	return &VerifyMFAOutput{Tokens: pair}, nil
}

// DisableMFAUseCase removes MFA from a user account.
type DisableMFAUseCase struct {
	users domain.UserRepository
}

func NewDisableMFAUseCase(users domain.UserRepository) *DisableMFAUseCase {
	return &DisableMFAUseCase{users: users}
}

func (uc *DisableMFAUseCase) Execute(ctx context.Context, tenantID, userID uuid.UUID) error {
	user, err := uc.users.GetByID(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("carica utente: %w", err)
	}
	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}
	user.MFAEnabled = false
	user.MFASecret = nil
	user.MFABackupCodes = nil
	user.UpdatedAt = time.Now().UTC()
	return uc.users.Update(ctx, user)
}

// generateBackupCodes creates n random hex codes.
func generateBackupCodes(n int) ([]string, error) {
	codes := make([]string, n)
	for i := range codes {
		b := make([]byte, mfaBackupCodeLen)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		codes[i] = hex.EncodeToString(b)
	}
	return codes, nil
}

// consumeBackupCode searches for code in the list and returns the remaining slice.
func consumeBackupCode(codes []string, code string) (used string, remaining []string, ok bool) {
	for i, c := range codes {
		if c == code {
			remaining = append(codes[:i:i], codes[i+1:]...)
			return c, remaining, true
		}
	}
	return "", codes, false
}
