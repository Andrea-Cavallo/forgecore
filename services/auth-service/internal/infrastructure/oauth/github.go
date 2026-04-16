package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
)

const (
	githubAuthURL  = "https://github.com/login/oauth/authorize"
	githubTokenURL = "https://github.com/login/oauth/access_token"
	githubUserURL  = "https://api.github.com/user"
	githubEmailURL = "https://api.github.com/user/emails"
)

// GitHubProvider implements the OAuth2 flow for GitHub Sign-In.
type GitHubProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// NewGitHubProvider creates a GitHubProvider with the given credentials.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the GitHub OAuth consent URL for the given CSRF state.
func (g *GitHubProvider) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":    {g.clientID},
		"redirect_uri": {g.redirectURL},
		"scope":        {"read:user user:email"},
		"state":        {state},
	}
	return githubAuthURL + "?" + params.Encode()
}

// Exchange trades an authorization code for a normalized OAuthUser.
func (g *GitHubProvider) Exchange(ctx context.Context, code string) (*domain.OAuthUser, error) {
	token, err := g.exchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("scambio codice github: %w", err)
	}
	return g.fetchUserInfo(ctx, token)
}

func (g *GitHubProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"code":          {code},
		"redirect_uri":  {g.redirectURL},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("creazione richiesta token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("chiamata token github: %w", err)
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
		return "", fmt.Errorf("errore provider github: %s", result.Error)
	}
	return result.AccessToken, nil
}

func (g *GitHubProvider) fetchUserInfo(ctx context.Context, accessToken string) (*domain.OAuthUser, error) {
	user, err := g.fetchProfile(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if user.Email == "" {
		email, err := g.fetchPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		user.Email = email
	}
	if user.Email == "" {
		return nil, domain.ErrOAuthEmailMissing
	}
	return user, nil
}

func (g *GitHubProvider) fetchProfile(ctx context.Context, accessToken string) (*domain.OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creazione richiesta profilo: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chiamata profilo github: %w", err)
	}
	defer resp.Body.Close()
	var info struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode profilo github: %w", err)
	}
	name := info.Name
	if name == "" {
		name = info.Login
	}
	return &domain.OAuthUser{
		ProviderID: fmt.Sprintf("%d", info.ID),
		Provider:   "github",
		Email:      info.Email,
		Name:       name,
	}, nil
}

func (g *GitHubProvider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubEmailURL, nil)
	if err != nil {
		return "", fmt.Errorf("creazione richiesta email: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("chiamata email github: %w", err)
	}
	defer resp.Body.Close()
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
		Verify  bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode email github: %w", err)
	}
	for _, e := range emails {
		if e.Primary && e.Verify {
			return e.Email, nil
		}
	}
	return "", nil
}
