package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
)

// ExportDataOutput holds all user data returned for a GDPR export request.
type ExportDataOutput struct {
	User     *domain.User      `json:"user"`
	Sessions []*domain.Session `json:"sessions"`
}

// ExportDataUseCase assembles all personal data held for a user (GDPR Art. 20).
type ExportDataUseCase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
}

func NewExportDataUseCase(users domain.UserRepository, sessions domain.SessionRepository) *ExportDataUseCase {
	return &ExportDataUseCase{users: users, sessions: sessions}
}

// Execute returns the full data export for the given user.
func (uc *ExportDataUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) (*ExportDataOutput, error) {
	user, err := uc.users.GetByID(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("recupero utente per export: %w", err)
	}
	sessions, err := uc.sessions.ListByUser(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("recupero sessioni per export: %w", err)
	}
	return &ExportDataOutput{User: user, Sessions: sessions}, nil
}

// DeleteAccountUseCase soft-deletes a user and zeroes out all PII (GDPR Art. 17).
type DeleteAccountUseCase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
}

func NewDeleteAccountUseCase(users domain.UserRepository, sessions domain.SessionRepository) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{users: users, sessions: sessions}
}

// Execute soft-deletes the account, zeroes PII fields, and revokes all sessions.
func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) error {
	user, err := uc.users.GetByID(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("recupero utente per eliminazione: %w", err)
	}
	now := time.Now().UTC()
	user.DeletedAt = &now
	user.EmailEnc = nil
	user.EmailHash = ""
	user.PasswordHash = ""
	user.MFASecret = nil
	user.MFABackupCodes = nil
	user.OAuthProvider = ""
	user.OAuthProviderID = ""
	user.UpdatedAt = now
	if err := uc.users.Update(ctx, user); err != nil {
		return fmt.Errorf("eliminazione account fallita: %w", err)
	}
	if err := uc.sessions.DeleteByUserID(ctx, userID, tenantID); err != nil {
		return fmt.Errorf("revoca sessioni fallita: %w", err)
	}
	return nil
}
