package application

import (
	"context"
	"fmt"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/google/uuid"
)

type LogoutUseCase struct {
	sessions domain.SessionRepository
	tokens   domain.TokenStore
}

func NewLogoutUseCase(sessions domain.SessionRepository, tokens domain.TokenStore) *LogoutUseCase {
	return &LogoutUseCase{sessions: sessions, tokens: tokens}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, tenantID, userID, sessionID uuid.UUID, jti string) error {
	if jti != "" {
		if err := uc.tokens.BlacklistJTI(ctx, jti, accessTokenTTLSeconds); err != nil {
			return fmt.Errorf("blacklist token: %w", err)
		}
	}
	if sessionID == uuid.Nil {
		return uc.sessions.DeleteByUserID(ctx, userID, tenantID)
	}
	return NewRevokeSessionUseCase(uc.sessions).Execute(ctx, tenantID, userID, sessionID)
}
