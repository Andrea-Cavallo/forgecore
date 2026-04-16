package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// --- session repo stub ---

type stubSessionRepo struct{ createErr error }

func (r *stubSessionRepo) Create(_ context.Context, _ *domain.Session) error { return r.createErr }
func (r *stubSessionRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*domain.Session, error) {
	return nil, domain.ErrSessionNotFound
}
func (r *stubSessionRepo) ListByUser(_ context.Context, _, _ uuid.UUID) ([]*domain.Session, error) {
	return nil, nil
}
func (r *stubSessionRepo) DeleteByID(_ context.Context, _, _ uuid.UUID) error    { return nil }
func (r *stubSessionRepo) DeleteByUserID(_ context.Context, _, _ uuid.UUID) error { return nil }
func (r *stubSessionRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error    { return nil }

// --- token store stub ---

type stubTokenStore struct{}

func (s *stubTokenStore) StoreRefreshToken(_ context.Context, _, _ string, _ int64) error {
	return nil
}
func (s *stubTokenStore) ValidateRefreshToken(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}
func (s *stubTokenStore) BlacklistJTI(_ context.Context, _ string, _ int64) error { return nil }
func (s *stubTokenStore) IsBlacklisted(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (s *stubTokenStore) IncrBruteForce(_ context.Context, _ string) (int64, error) { return 1, nil }
func (s *stubTokenStore) SetBruteForceLockout(_ context.Context, _ string, _ int64) error {
	return nil
}
func (s *stubTokenStore) GetBruteForceCount(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (s *stubTokenStore) StoreOneTimeToken(_ context.Context, _, _ string, _ int64) error {
	return nil
}
func (s *stubTokenStore) PopOneTimeToken(_ context.Context, _ string) (string, error) {
	return "", nil
}

// --- jwt service stub ---

type stubTokenIssuer struct{}

func (s *stubTokenIssuer) Issue(claims domain.TokenClaims) (domain.TokenPair, error) {
	return domain.TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresIn: 900}, nil
}

// --- tests ---

func buildLoginUser(password string) (*domain.User, uuid.UUID) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 4)
	tenantID := uuid.New()
	return &domain.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		EmailHash:    "hashvalue",
		PasswordHash: string(hash),
		Roles:        []string{"user"},
	}, tenantID
}

func TestLoginUseCase_Success(t *testing.T) {
	const password = "SecurePass123!"
	user, tenantID := buildLoginUser(password)
	enc := testEncryptor()

	repo := newStubUserRepo()
	repo.users[enc.Hash("user@example.com")] = user
	user.EmailHash = enc.Hash("user@example.com")
	user.TenantID = tenantID

	uc := NewLoginUseCase(repo, &stubSessionRepo{}, &stubTokenStore{}, nil, enc, &stubTokenIssuer{})

	out, err := uc.Execute(context.Background(), LoginInput{
		TenantID:  tenantID,
		Email:     "user@example.com",
		Password:  password,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("atteso nessun errore, ottenuto: %v", err)
	}
	if out.Tokens.AccessToken == "" {
		t.Error("AccessToken non deve essere vuoto")
	}
}

func TestLoginUseCase_InvalidPassword(t *testing.T) {
	user, tenantID := buildLoginUser("correctpassword")
	enc := testEncryptor()
	repo := newStubUserRepo()
	repo.users[enc.Hash("user@example.com")] = user
	user.EmailHash = enc.Hash("user@example.com")
	user.TenantID = tenantID

	uc := NewLoginUseCase(repo, &stubSessionRepo{}, &stubTokenStore{}, nil, enc, &stubTokenIssuer{})

	_, err := uc.Execute(context.Background(), LoginInput{
		TenantID:  tenantID,
		Email:     "user@example.com",
		Password:  "wrongpassword",
		IPAddress: "127.0.0.1",
	})
	if err == nil {
		t.Fatal("atteso errore, ottenuto nil")
	}
}

func TestLoginUseCase_LockedAccount(t *testing.T) {
	user, tenantID := buildLoginUser("password123")
	future := time.Now().Add(1 * time.Hour)
	user.LockedUntil = &future
	enc := testEncryptor()
	repo := newStubUserRepo()
	repo.users[enc.Hash("locked@example.com")] = user
	user.EmailHash = enc.Hash("locked@example.com")
	user.TenantID = tenantID

	uc := NewLoginUseCase(repo, &stubSessionRepo{}, &stubTokenStore{}, nil, enc, &stubTokenIssuer{})

	_, err := uc.Execute(context.Background(), LoginInput{
		TenantID:  tenantID,
		Email:     "locked@example.com",
		Password:  "password123",
		IPAddress: "127.0.0.1",
	})
	if err == nil {
		t.Fatal("atteso errore per account bloccato, ottenuto nil")
	}
}

func TestJWTService_IssueAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-that-is-32bytes-long!")

	userID := uuid.New()
	tenantID := uuid.New()
	pair, err := svc.Issue(domain.TokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		Roles:    []string{"user"},
		JTI:      uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("AccessToken vuoto")
	}

	claims, err := svc.Validate(pair.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %v want %v", claims.UserID, userID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("TenantID mismatch: got %v want %v", claims.TenantID, tenantID)
	}
}

func TestJWTService_Validate_InvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret-that-is-32bytes-long!")
	_, err := svc.Validate("invalid.token.here")
	if err == nil {
		t.Error("atteso errore per token non valido, ottenuto nil")
	}
}
