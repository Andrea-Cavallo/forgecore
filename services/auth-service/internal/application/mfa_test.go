package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/application"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

// mfaUserRepo is an in-memory stub that supports GetByID for MFA tests.
type mfaUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMFAUserRepo() *mfaUserRepo {
	return &mfaUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (r *mfaUserRepo) Create(_ context.Context, u *domain.User) error {
	r.users[u.ID] = u
	return nil
}

func (r *mfaUserRepo) GetByID(_ context.Context, id, _ uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *mfaUserRepo) GetByEmailHash(_ context.Context, hash string, tenantID uuid.UUID) (*domain.User, error) {
	for _, u := range r.users {
		if u.EmailHash == hash && u.TenantID == tenantID {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *mfaUserRepo) Update(_ context.Context, u *domain.User) error {
	r.users[u.ID] = u
	return nil
}

func (r *mfaUserRepo) Delete(_ context.Context, id, _ uuid.UUID) error {
	delete(r.users, id)
	return nil
}

func (r *mfaUserRepo) GetByOAuthProvider(_ context.Context, _, _ string, _ uuid.UUID) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (r *mfaUserRepo) ListByTenant(_ context.Context, _ uuid.UUID, _ pagination.Cursor) ([]*domain.User, error) {
	return nil, nil
}

// testPublisher satisfies application.EventPublisher.
type testPublisher struct{}

func (testPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

// testTokenIssuer returns fixed tokens.
type testTokenIssuer struct{}

func (testTokenIssuer) Issue(_ domain.TokenClaims) (domain.TokenPair, error) {
	return domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}, nil
}

func newMFATestUser(tenantID uuid.UUID) *domain.User {
	return &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmailHash:     "hash@test.com",
		EmailEnc:      []byte("enc@test.com"),
		PasswordHash:  "$2a$12$testhash",
		Roles:         []string{"user"},
		EmailVerified: true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func TestEnableMFA_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	_ = repo.Create(ctx, user)

	uc := application.NewEnableMFAUseCase(repo, testPublisher{})
	out, err := uc.Execute(ctx, application.EnableMFAInput{TenantID: tenantID, UserID: user.ID})
	if err != nil {
		t.Fatalf("atteso nessun errore, ricevuto: %v", err)
	}
	if out.Secret == "" {
		t.Error("secret non deve essere vuoto")
	}
	if out.QRCodeURL == "" {
		t.Error("QR code URL non deve essere vuoto")
	}
	if len(out.BackupCodes) != 8 {
		t.Errorf("attesi 8 backup codes, ricevuti %d", len(out.BackupCodes))
	}
}

func TestEnableMFA_AlreadyEnabled(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	user.MFAEnabled = true
	_ = repo.Create(ctx, user)

	uc := application.NewEnableMFAUseCase(repo, testPublisher{})
	_, err := uc.Execute(ctx, application.EnableMFAInput{TenantID: tenantID, UserID: user.ID})
	if err != application.ErrMFAAlreadyEnabled {
		t.Errorf("atteso ErrMFAAlreadyEnabled, ricevuto: %v", err)
	}
}

func TestVerifyMFA_ValidCode(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	_ = repo.Create(ctx, user)

	// Abilita prima MFA per ottenere il secret
	enableUC := application.NewEnableMFAUseCase(repo, testPublisher{})
	enableOut, err := enableUC.Execute(ctx, application.EnableMFAInput{TenantID: tenantID, UserID: user.ID})
	if err != nil {
		t.Fatalf("Enable MFA: %v", err)
	}

	// Genera codice TOTP valido con il secret appena creato
	code, err := totp.GenerateCode(enableOut.Secret, time.Now())
	if err != nil {
		t.Fatalf("genera codice TOTP: %v", err)
	}

	verifyUC := application.NewVerifyMFAUseCase(repo, testTokenIssuer{}, testPublisher{})
	out, err := verifyUC.Execute(ctx, application.VerifyMFAInput{
		TenantID: tenantID,
		UserID:   user.ID,
		Code:     code,
	})
	if err != nil {
		t.Fatalf("atteso nessun errore, ricevuto: %v", err)
	}
	if out.Tokens.AccessToken != "access" {
		t.Errorf("atteso token 'access', ricevuto: %s", out.Tokens.AccessToken)
	}
}

func TestVerifyMFA_InvalidCode(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	user.MFASecret = []byte("JBSWY3DPEHPK3PXP")
	_ = repo.Create(ctx, user)

	verifyUC := application.NewVerifyMFAUseCase(repo, testTokenIssuer{}, testPublisher{})
	_, err := verifyUC.Execute(ctx, application.VerifyMFAInput{
		TenantID: tenantID,
		UserID:   user.ID,
		Code:     "000000",
	})
	if err != application.ErrMFAInvalidCode {
		t.Errorf("atteso ErrMFAInvalidCode, ricevuto: %v", err)
	}
}

func TestVerifyMFA_BackupCode(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	user.MFASecret = []byte("JBSWY3DPEHPK3PXP")
	backupCode := "aabbccddeeff0011"
	user.MFABackupCodes = []string{backupCode, "other1122334455"}
	_ = repo.Create(ctx, user)

	verifyUC := application.NewVerifyMFAUseCase(repo, testTokenIssuer{}, testPublisher{})
	out, err := verifyUC.Execute(ctx, application.VerifyMFAInput{
		TenantID: tenantID,
		UserID:   user.ID,
		Code:     backupCode,
	})
	if err != nil {
		t.Fatalf("atteso nessun errore con backup code, ricevuto: %v", err)
	}
	if out.Tokens.AccessToken != "access" {
		t.Errorf("atteso token 'access', ricevuto: %s", out.Tokens.AccessToken)
	}

	// Verifica che il backup code sia stato consumato
	updated, _ := repo.GetByID(ctx, user.ID, tenantID)
	if len(updated.MFABackupCodes) != 1 {
		t.Errorf("atteso 1 backup code rimanente, ricevuti %d", len(updated.MFABackupCodes))
	}
}

func TestDisableMFA_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMFAUserRepo()
	tenantID := uuid.New()
	user := newMFATestUser(tenantID)
	user.MFAEnabled = true
	user.MFASecret = []byte("JBSWY3DPEHPK3PXP")
	_ = repo.Create(ctx, user)

	uc := application.NewDisableMFAUseCase(repo)
	if err := uc.Execute(ctx, tenantID, user.ID); err != nil {
		t.Fatalf("atteso nessun errore, ricevuto: %v", err)
	}

	updated, _ := repo.GetByID(ctx, user.ID, tenantID)
	if updated.MFAEnabled {
		t.Error("atteso MFAEnabled=false dopo disable")
	}
	if len(updated.MFASecret) != 0 {
		t.Error("atteso MFASecret vuoto dopo disable")
	}
}
