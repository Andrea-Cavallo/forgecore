package domain

// OAuthUser holds the normalized profile returned by an OAuth2 provider.
type OAuthUser struct {
	ProviderID string // provider-specific user ID (stable)
	Provider   string // "google" | "github"
	Email      string
	Name       string
}
