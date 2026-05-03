// Package auth fornisce il client SDK per forgecore-auth.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"

	"github.com/Andrea-Cavallo/golang-modules/sdk/go/clientretry"
	"github.com/Andrea-Cavallo/golang-modules/sdk/go/clienttransport"
)

const defaultTimeout = 5 * time.Second

// Client è il client HTTP per forgecore-auth con retry, circuit breaker e OTEL.
type Client struct {
	baseURL string
	http    *http.Client
	cb      *gobreaker.CircuitBreaker
}

// NewClient crea un Client puntando a baseURL (es. "http://forgecore-auth:8081").
func NewClient(baseURL string) *Client {
	transport := clienttransport.NewOTELTransport(nil)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "forgecore-auth",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
	})
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout, Transport: transport},
		cb:      cb,
	}
}

// ValidateTokenResponse è la risposta di /v1/auth/validate.
type ValidateTokenResponse struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
}

// ValidateToken verifica il JWT presso forgecore-auth.
func (c *Client) ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error) {
	var result ValidateTokenResponse
	err := clientretry.RetryFn(ctx, func() error {
		_, cbErr := c.cb.Execute(func() (any, error) {
			return nil, c.doValidate(ctx, token, &result)
		})
		return cbErr
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateToken: %w", err)
	}
	return &result, nil
}

func (c *Client) doValidate(ctx context.Context, token string, out *ValidateTokenResponse) error {
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return fmt.Errorf("marshal richiesta validazione token: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/auth/validate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	clienttransport.InjectBearer(ctx, req, token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("forgecore-auth risposta %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
