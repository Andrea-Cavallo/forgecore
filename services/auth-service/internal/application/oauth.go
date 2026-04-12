package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/auth-service/internal/domain"
	"github.com/yourorg/golang-modules/shared/crypto"
	"github.com/yourorg/golang-modules/shared/events"
	"golang.org/x/crypto/bcrypt"
)

// OAuthProvider is the interface implemented by Google / GitHub adapters.
type OAuthProvider interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*domain.OAuthUser, error)
}

// OAuthAuthorizeUseCase generates an OAuth redirect URL and stores CSRF state.
type OAuthAuthorizeUseCase struct {
	tokens domain.TokenStore
}

func NewOAuthAuthorizeUseCase(tokens domain.TokenStore) *OAuthAuthorizeUseCase {
	return &OAuthAuthorizeUseCase{tokens: tokens}
}

const oAuthStateTTL = 300 // 5 minuti

// Execute returns the provider auth URL and the CSRF state token.
func (uc *OAuthAuthorizeUseCase) Execute(ctx context.Context, provider OAuthProvider) (authURL, state string, err error) {
	state = uuid.New().String()
	key := "oauth:state:" + state
	if err = uc.tokens.StoreOneTimeToken(ctx, key, state, oAuthStateTTL); err != nil {
		return "", "", fmt.Errorf("salvataggio stato oauth: %w", err)
	}
	return provider.GetAuthURL(state), state, nil
}

// OAuthCallbackInput holds the data received from the OAuth provider callback.
type OAuthCallbackInput struct {
	TenantID  uuid.UUID
	Code      string
	State     string
	IPAddress string
	UserAgent string
	DeviceID  string
}

// OAuthCallbackUseCase handles the OAuth2 callback: verifies state, upserts user, issues tokens.
type OAuthCallbackUseCase struct {
	users     domain.UserRepository
	sessions  domain.SessionRepository
	tokens    domain.TokenStore
	issuer    TokenIssuer
	publisher EventPublisher
	encryptor *crypto.PIIEncryptor
}

func NewOAuthCallbackUseCase(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	tokens domain.TokenStore,
	issuer TokenIssuer,
	pub EventPublisher,
	enc *crypto.PIIEncryptor,
) *OAuthCallbackUseCase {
	if pub == nil {
		pub = noopPublisher{}
	}
	return &OAuthCallbackUseCase{
		users: users, sessions: sessions, tokens: tokens,
		issuer: issuer, publisher: pub, encryptor: enc,
	}
}

// Execute processes the OAuth callback and returns a JWT token pair.
func (uc *OAuthCallbackUseCase) Execute(ctx context.Context, input OAuthCallbackInput, provider OAuthProvider) (*LoginOutput, error) {
	if err := uc.verifyState(ctx, input.State); err != nil {
		return nil, err
	}
	oauthUser, err := provider.Exchange(ctx, input.Code)
	if err != nil {
		return nil, fmt.Errorf("scambio codice oauth: %w", err)
	}
	user, err := uc.upsertUser(ctx, oauthUser, input.TenantID)
	if err != nil {
		return nil, err
	}
	return uc.issueTokensAndSession(ctx, user, input)
}

func (uc *OAuthCallbackUseCase) verifyState(ctx context.Context, state string) error {
	key := "oauth:state:" + state
	stored, err := uc.tokens.PopOneTimeToken(ctx, key)
	if err != nil || stored != state {
		return domain.ErrOAuthStateMismatch
	}
	return nil
}

func (uc *OAuthCallbackUseCase) upsertUser(ctx context.Context, oauthUser *domain.OAuthUser, tenantID uuid.UUID) (*domain.User, error) {
	// 1. Look up by OAuth identity
	user, err := uc.users.GetByOAuthProvider(ctx, oauthUser.Provider, oauthUser.ProviderID, tenantID)
	if err == nil {
		return user, nil // existing OAuth user
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("ricerca utente oauth: %w", err)
	}
	// 2. Look up by email — link if exists
	emailHash := uc.encryptor.Hash(oauthUser.Email)
	user, err = uc.users.GetByEmailHash(ctx, emailHash, tenantID)
	if err == nil {
		return uc.linkOAuthToExisting(ctx, user, oauthUser)
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("ricerca utente per email: %w", err)
	}
	// 3. Create new user
	return uc.createOAuthUser(ctx, oauthUser, tenantID, emailHash)
}

func (uc *OAuthCallbackUseCase) linkOAuthToExisting(ctx context.Context, user *domain.User, oauthUser *domain.OAuthUser) (*domain.User, error) {
	user.OAuthProvider = oauthUser.Provider
	user.OAuthProviderID = oauthUser.ProviderID
	user.UpdatedAt = time.Now().UTC()
	if err := uc.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("collegamento oauth: %w", err)
	}
	return user, nil
}

func (uc *OAuthCallbackUseCase) createOAuthUser(ctx context.Context, oauthUser *domain.OAuthUser, tenantID uuid.UUID, emailHash string) (*domain.User, error) {
	emailEnc, err := uc.encryptor.Encrypt(oauthUser.Email)
	if err != nil {
		return nil, fmt.Errorf("cifratura email: %w", err)
	}
	// Password-less user: store a random unusable hash
	unusable, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	now := time.Now().UTC()
	user := &domain.User{
		ID:              uuid.New(),
		TenantID:        tenantID,
		EmailEnc:        []byte(emailEnc),
		EmailHash:       emailHash,
		PasswordHash:    string(unusable),
		Roles:           []string{"user"},
		EmailVerified:   true, // OAuth providers verify email
		OAuthProvider:   oauthUser.Provider,
		OAuthProviderID: oauthUser.ProviderID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creazione utente oauth: %w", err)
	}
	return user, nil
}

func (uc *OAuthCallbackUseCase) issueTokensAndSession(ctx context.Context, user *domain.User, input OAuthCallbackInput) (*LoginOutput, error) {
	pair, err := uc.issuer.Issue(domain.TokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Roles:    user.Roles,
		JTI:      uuid.New().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("emissione token: %w", err)
	}
	session := &domain.Session{
		ID:         uuid.New(),
		TenantID:   input.TenantID,
		UserID:     user.ID,
		DeviceID:   input.DeviceID,
		UserAgent:  input.UserAgent,
		IPAddress:  input.IPAddress,
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(sessionDuration),
	}
	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creazione sessione oauth: %w", err)
	}
	_ = uc.publisher.Publish(ctx, events.SubjectUserLogin, events.UserLogin{
		TenantID:   user.TenantID,
		UserID:     user.ID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		OccurredAt: time.Now().UTC(),
	})
	return &LoginOutput{Tokens: pair}, nil
}
