package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
)

// ListSessionsUseCase retrieves active sessions for a user.
type ListSessionsUseCase struct {
	sessions domain.SessionRepository
}

func NewListSessionsUseCase(sessions domain.SessionRepository) *ListSessionsUseCase {
	return &ListSessionsUseCase{sessions: sessions}
}

type ListSessionsOutput struct {
	Sessions []*domain.Session
}

func (uc *ListSessionsUseCase) Execute(ctx context.Context, tenantID, userID uuid.UUID) (*ListSessionsOutput, error) {
	list, err := uc.sessions.ListByUser(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("elenco sessioni: %w", err)
	}
	return &ListSessionsOutput{Sessions: list}, nil
}

// RevokeSessionUseCase deletes a specific session owned by a user.
type RevokeSessionUseCase struct {
	sessions domain.SessionRepository
}

func NewRevokeSessionUseCase(sessions domain.SessionRepository) *RevokeSessionUseCase {
	return &RevokeSessionUseCase{sessions: sessions}
}

func (uc *RevokeSessionUseCase) Execute(ctx context.Context, tenantID, userID, sessionID uuid.UUID) error {
	sess, err := uc.sessions.GetByID(ctx, sessionID, tenantID)
	if err != nil {
		return fmt.Errorf("carica sessione: %w", err)
	}
	if sess.UserID != userID {
		return domain.ErrSessionNotFound
	}
	return uc.sessions.DeleteByID(ctx, sessionID, tenantID)
}

// RevokeAllSessionsUseCase logs out a user from all devices.
type RevokeAllSessionsUseCase struct {
	sessions domain.SessionRepository
}

func NewRevokeAllSessionsUseCase(sessions domain.SessionRepository) *RevokeAllSessionsUseCase {
	return &RevokeAllSessionsUseCase{sessions: sessions}
}

func (uc *RevokeAllSessionsUseCase) Execute(ctx context.Context, tenantID, userID uuid.UUID) error {
	return uc.sessions.DeleteByUserID(ctx, userID, tenantID)
}
