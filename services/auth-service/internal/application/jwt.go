package application

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// JWTService implements TokenIssuer using HMAC-SHA256.
type JWTService struct{ secret []byte }

// NewJWTService creates a JWTService with the given HMAC secret.
func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	TenantID    string   `json:"tenant_id"`
	Roles       []string `json:"roles"`
	MFAVerified bool     `json:"mfa_verified"`
}

// Issue mints an access+refresh token pair from the provided claims.
func (s *JWTService) Issue(claims domain.TokenClaims) (domain.TokenPair, error) {
	now := time.Now().UTC()
	jti := claims.JTI
	if jti == "" {
		jti = uuid.New().String()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
		TenantID:    claims.TenantID.String(),
		Roles:       claims.Roles,
		MFAVerified: claims.MFAVerified,
	})
	accessStr, err := token.SignedString(s.secret)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("firma access token: %w", err)
	}
	return domain.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: uuid.New().String(),
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

// Validate parses and validates a JWT access token. Returns claims and JTI.
func (s *JWTService) Validate(tokenStr string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo di firma inatteso: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, domain.ErrTokenExpired
	}
	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}
	userID, err := uuid.Parse(c.Subject)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	tenantID, err := uuid.Parse(c.TenantID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	return &domain.TokenClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       c.Roles,
		JTI:         c.ID,
		MFAVerified: c.MFAVerified,
	}, nil
}
