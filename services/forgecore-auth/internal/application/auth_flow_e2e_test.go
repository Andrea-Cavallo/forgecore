package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
	"github.com/google/uuid"
)

type authFlowUserRepo struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func newAuthFlowUserRepo() *authFlowUserRepo {
	return &authFlowUserRepo{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}}
}

func (r *authFlowUserRepo) Create(_ context.Context, u *domain.User) error {
	r.byEmail[u.EmailHash] = u
	r.byID[u.ID] = u
	return nil
}

func (r *authFlowUserRepo) GetByID(_ context.Context, id, tenantID uuid.UUID) (*domain.User, error) {
	user, ok := r.byID[id]
	if !ok || user.TenantID != tenantID {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *authFlowUserRepo) GetByEmailHash(_ context.Context, hash string, tenantID uuid.UUID) (*domain.User, error) {
	user, ok := r.byEmail[hash]
	if !ok || user.TenantID != tenantID {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *authFlowUserRepo) GetByOAuthProvider(context.Context, string, string, uuid.UUID) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *authFlowUserRepo) Update(_ context.Context, u *domain.User) error {
	r.byID[u.ID] = u
	r.byEmail[u.EmailHash] = u
	return nil
}

func (r *authFlowUserRepo) Delete(_ context.Context, id, tenantID uuid.UUID) error {
	user, ok := r.byID[id]
	if !ok || user.TenantID != tenantID {
		return domain.ErrUserNotFound
	}
	delete(r.byID, id)
	delete(r.byEmail, user.EmailHash)
	return nil
}

func (r *authFlowUserRepo) ListByTenant(context.Context, uuid.UUID, pagination.Cursor) ([]*domain.User, error) {
	return nil, nil
}

type authFlowSessionRepo struct {
	byID map[uuid.UUID]*domain.Session
}

func newAuthFlowSessionRepo() *authFlowSessionRepo {
	return &authFlowSessionRepo{byID: map[uuid.UUID]*domain.Session{}}
}

func (r *authFlowSessionRepo) Create(_ context.Context, s *domain.Session) error {
	r.byID[s.ID] = s
	return nil
}

func (r *authFlowSessionRepo) GetByID(_ context.Context, id, tenantID uuid.UUID) (*domain.Session, error) {
	session, ok := r.byID[id]
	if !ok || session.TenantID != tenantID {
		return nil, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *authFlowSessionRepo) ListByUser(_ context.Context, userID, tenantID uuid.UUID) ([]*domain.Session, error) {
	var out []*domain.Session
	for _, session := range r.byID {
		if session.UserID == userID && session.TenantID == tenantID {
			out = append(out, session)
		}
	}
	return out, nil
}

func (r *authFlowSessionRepo) DeleteByID(_ context.Context, id, tenantID uuid.UUID) error {
	session, ok := r.byID[id]
	if !ok || session.TenantID != tenantID {
		return domain.ErrSessionNotFound
	}
	delete(r.byID, id)
	return nil
}

func (r *authFlowSessionRepo) DeleteByUserID(_ context.Context, userID, tenantID uuid.UUID) error {
	for id, session := range r.byID {
		if session.UserID == userID && session.TenantID == tenantID {
			delete(r.byID, id)
		}
	}
	return nil
}

func (r *authFlowSessionRepo) UpdateLastSeen(context.Context, uuid.UUID) error {
	return nil
}

type authFlowTokenStore struct {
	refresh  map[string]string
	blackJTI map[string]struct{}
}

func newAuthFlowTokenStore() *authFlowTokenStore {
	return &authFlowTokenStore{refresh: map[string]string{}, blackJTI: map[string]struct{}{}}
}

func (s *authFlowTokenStore) StoreRefreshToken(_ context.Context, key, token string, _ int64) error {
	s.refresh[key] = token
	return nil
}

func (s *authFlowTokenStore) ValidateRefreshToken(_ context.Context, key, token string) (bool, error) {
	return s.refresh[key] == token, nil
}

func (s *authFlowTokenStore) BlacklistJTI(_ context.Context, jti string, _ int64) error {
	s.blackJTI[jti] = struct{}{}
	return nil
}

func (s *authFlowTokenStore) IsBlacklisted(_ context.Context, jti string) (bool, error) {
	_, ok := s.blackJTI[jti]
	return ok, nil
}

func (s *authFlowTokenStore) IncrBruteForce(context.Context, string) (int64, error) {
	return 1, nil
}

func (s *authFlowTokenStore) SetBruteForceLockout(context.Context, string, int64) error {
	return nil
}

func (s *authFlowTokenStore) GetBruteForceCount(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *authFlowTokenStore) StoreOneTimeToken(context.Context, string, string, int64) error {
	return nil
}

func (s *authFlowTokenStore) PopOneTimeToken(context.Context, string) (string, error) {
	return "", errors.New("token non trovato")
}

func TestAuthFrontendFlowE2E_RegisterLoginRefreshLogoutMeProtected(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	users := newAuthFlowUserRepo()
	sessions := newAuthFlowSessionRepo()
	tokens := newAuthFlowTokenStore()
	enc := testEncryptor()
	jwtSvc := NewRotatingJWTService("kid-2026-05", "current-secret-that-is-long-enough", map[string]string{
		"kid-2026-04": "previous-secret-that-is-long-enough",
	})

	registerOut, err := NewRegisterUseCase(users, nil, enc).Execute(ctx, RegisterInput{
		TenantID: tenantID,
		Email:    "frontend@example.com",
		Password: "ChangeMe123!",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	loginOut, err := NewLoginUseCase(users, sessions, tokens, nil, enc, jwtSvc).Execute(ctx, LoginInput{
		TenantID:  tenantID,
		Email:     "frontend@example.com",
		Password:  "ChangeMe123!",
		IPAddress: "127.0.0.1",
		UserAgent: "forgecore-e2e",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	claims, err := jwtSvc.Validate(loginOut.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("validate protected token: %v", err)
	}
	if claims.UserID != registerOut.UserID {
		t.Fatalf("protected user mismatch: got %s want %s", claims.UserID, registerOut.UserID)
	}

	meOut, err := NewMeUseCase(users).Execute(ctx, tenantID, claims.UserID)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if meOut.TenantID != tenantID {
		t.Fatalf("tenant mismatch: got %s want %s", meOut.TenantID, tenantID)
	}

	refreshOut, err := NewRefreshTokenUseCase(tokens, jwtSvc).Execute(ctx, RefreshTokenInput{
		TenantID:     tenantID,
		UserID:       claims.UserID,
		Roles:        claims.Roles,
		RefreshJTI:   claims.JTI,
		RefreshToken: loginOut.Tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshOut.Tokens.AccessToken == loginOut.Tokens.AccessToken {
		t.Fatal("refresh deve emettere un nuovo access token")
	}

	if err := NewLogoutUseCase(sessions, tokens).Execute(ctx, tenantID, claims.UserID, uuid.Nil, claims.JTI); err != nil {
		t.Fatalf("logout: %v", err)
	}
	blacklisted, err := tokens.IsBlacklisted(ctx, claims.JTI)
	if err != nil {
		t.Fatalf("blacklist check: %v", err)
	}
	if !blacklisted {
		t.Fatal("token precedente non inserito in blacklist")
	}
}
