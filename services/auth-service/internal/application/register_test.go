package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/auth-service/internal/domain"
	"github.com/yourorg/golang-modules/shared/crypto"
	"github.com/yourorg/golang-modules/shared/pagination"
)

// --- stubs ---

type stubUserRepo struct {
	users     map[string]*domain.User
	createErr error
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[string]*domain.User)}
}

func (r *stubUserRepo) Create(_ context.Context, u *domain.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.users[u.EmailHash] = u
	return nil
}

func (r *stubUserRepo) GetByEmailHash(_ context.Context, hash string, _ uuid.UUID) (*domain.User, error) {
	u, ok := r.users[hash]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *stubUserRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (r *stubUserRepo) Update(_ context.Context, _ *domain.User) error { return nil }
func (r *stubUserRepo) Delete(_ context.Context, _, _ uuid.UUID) error  { return nil }
func (r *stubUserRepo) ListByTenant(_ context.Context, _ uuid.UUID, _ pagination.Cursor) ([]*domain.User, error) {
	return nil, nil
}

func testEncryptor() *crypto.PIIEncryptor {
	return crypto.NewPIIEncryptor(make([]byte, 32), make([]byte, 32))
}

// --- tests ---

func TestRegisterUseCase_Success(t *testing.T) {
	repo := newStubUserRepo()
	uc := NewRegisterUseCase(repo, nil, testEncryptor())

	out, err := uc.Execute(context.Background(), RegisterInput{
		TenantID:  uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Mario",
		LastName:  "Rossi",
	})
	if err != nil {
		t.Fatalf("atteso nessun errore, ottenuto: %v", err)
	}
	if out.UserID == uuid.Nil {
		t.Error("UserID non deve essere nil")
	}
}

func TestRegisterUseCase_EmailAlreadyExists(t *testing.T) {
	repo := newStubUserRepo()
	enc := testEncryptor()
	emailHash := enc.Hash("existing@example.com")
	repo.users[emailHash] = &domain.User{ID: uuid.New(), EmailHash: emailHash}

	uc := NewRegisterUseCase(repo, nil, enc)

	_, err := uc.Execute(context.Background(), RegisterInput{
		TenantID: uuid.New(),
		Email:    "existing@example.com",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Errorf("atteso ErrEmailAlreadyExists, ottenuto: %v", err)
	}
}

func TestRegisterUseCase_InvalidInput(t *testing.T) {
	uc := NewRegisterUseCase(newStubUserRepo(), nil, testEncryptor())

	tests := []struct {
		name  string
		input RegisterInput
	}{
		{"email vuota", RegisterInput{TenantID: uuid.New(), Email: "", Password: "pass1234"}},
		{"password corta", RegisterInput{TenantID: uuid.New(), Email: "a@b.com", Password: "short"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), tc.input)
			if err == nil {
				t.Error("atteso errore di validazione, ottenuto nil")
			}
		})
	}
}

func TestRegisterUseCase_DBCreateError(t *testing.T) {
	repo := newStubUserRepo()
	repo.createErr = errors.New("db connection refused")

	uc := NewRegisterUseCase(repo, nil, testEncryptor())

	_, err := uc.Execute(context.Background(), RegisterInput{
		TenantID: uuid.New(),
		Email:    "new@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Error("atteso errore DB, ottenuto nil")
	}
}
