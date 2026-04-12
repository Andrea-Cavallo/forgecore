package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourorg/golang-modules/services/auth-service/internal/domain"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleProvider implements the OAuth2 flow for Google Sign-In.
type GoogleProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// NewGoogleProvider creates a GoogleProvider with the given credentials.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the Google consent screen URL for the given CSRF state.
func (g *GoogleProvider) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {g.clientID},
		"redirect_uri":  {g.redirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"select_account"},
	}
	return googleAuthURL + "?" + params.Encode()
}

// Exchange trades an authorization code for a normalized OAuthUser.
func (g *GoogleProvider) Exchange(ctx context.Context, code string) (*domain.OAuthUser, error) {
	token, err := g.exchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("scambio codice google: %w", err)
	}
	return g.fetchUserInfo(ctx, token)
}

func (g *GoogleProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"code":          {code},
		"redirect_uri":  {g.redirectURL},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("creazione richiesta token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("chiamata token google: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode risposta token: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("errore provider google: %s", result.Error)
	}
	return result.AccessToken, nil
}

func (g *GoogleProvider) fetchUserInfo(ctx context.Context, accessToken string) (*domain.OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creazione richiesta userinfo: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chiamata userinfo google: %w", err)
	}
	defer resp.Body.Close()
	var info struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode userinfo google: %w", err)
	}
	if info.Email == "" {
		return nil, fmt.Errorf("email non disponibile da google")
	}
	return &domain.OAuthUser{
		ProviderID: info.Sub,
		Provider:   "google",
		Email:      info.Email,
		Name:       info.Name,
	}, nil
}
