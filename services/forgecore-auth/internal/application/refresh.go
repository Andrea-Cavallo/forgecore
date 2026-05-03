package application

import (
	"context"
	"fmt"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/google/uuid"
)

type RefreshTokenInput struct {
	TenantID     uuid.UUID
	UserID       uuid.UUID
	Roles        []string
	RefreshToken string
	RefreshJTI   string
}

type RefreshTokenOutput struct {
	Tokens domain.TokenPair
	JTI    string
}

type RefreshTokenUseCase struct {
	tokens domain.TokenStore
	issuer TokenIssuer
}

func NewRefreshTokenUseCase(tokens domain.TokenStore, issuer TokenIssuer) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{tokens: tokens, issuer: issuer}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	valid, err := uc.tokens.ValidateRefreshToken(ctx, input.RefreshJTI, input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("validazione refresh token: %w", err)
	}
	if !valid {
		return nil, domain.ErrTokenInvalid
	}
	jti := uuid.New().String()
	pair, err := uc.issuer.Issue(domain.TokenClaims{
		UserID:   input.UserID,
		TenantID: input.TenantID,
		Roles:    input.Roles,
		JTI:      jti,
	})
	if err != nil {
		return nil, fmt.Errorf("emissione token refresh: %w", err)
	}
	if err := uc.tokens.StoreRefreshToken(ctx, jti, pair.RefreshToken, refreshTokenTTLSeconds); err != nil {
		return nil, fmt.Errorf("salvataggio refresh token: %w", err)
	}
	return &RefreshTokenOutput{Tokens: pair, JTI: jti}, nil
}
