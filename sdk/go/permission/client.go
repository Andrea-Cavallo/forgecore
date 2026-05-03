// Package permission fornisce il client SDK per forgecore-permissions.
package permission

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

// Client è il client HTTP per forgecore-permissions con retry, circuit breaker e OTEL.
type Client struct {
	baseURL string
	http    *http.Client
	cb      *gobreaker.CircuitBreaker
}

// NewClient crea un Client puntando a baseURL (es. "http://forgecore-permissions:8087").
func NewClient(baseURL string) *Client {
	transport := clienttransport.NewOTELTransport(nil)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "forgecore-permissions",
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

// CheckRequest è il payload per /v1/permissions/check.
type CheckRequest struct {
	UserID       string  `json:"user_id"`
	TenantID     string  `json:"tenant_id"`
	ResourceType string  `json:"resource_type"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Action       string  `json:"action"`
}

// CheckResponse è la risposta di /v1/permissions/check.
type CheckResponse struct {
	Allowed bool `json:"allowed"`
}

// Check verifica se un utente può eseguire un'azione su una risorsa.
func (c *Client) Check(ctx context.Context, token string, req CheckRequest) (bool, error) {
	var result CheckResponse
	err := clientretry.RetryFn(ctx, func() error {
		_, cbErr := c.cb.Execute(func() (any, error) {
			return nil, c.doCheck(ctx, token, req, &result)
		})
		return cbErr
	})
	if err != nil {
		return false, fmt.Errorf("Check: %w", err)
	}
	return result.Allowed, nil
}

func (c *Client) doCheck(ctx context.Context, token string, req CheckRequest, out *CheckResponse) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal richiesta: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/permissions/check", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	clienttransport.InjectBearer(ctx, httpReq, token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("forgecore-permissions risposta %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// SeedRoles chiama forgecore-permissions per creare i ruoli predefiniti per un tenant.
func (c *Client) SeedRoles(ctx context.Context, token, tenantID string) error {
	err := clientretry.RetryFn(ctx, func() error {
		_, cbErr := c.cb.Execute(func() (any, error) {
			return nil, c.doSeedRoles(ctx, token, tenantID)
		})
		return cbErr
	})
	if err != nil {
		return fmt.Errorf("SeedRoles: %w", err)
	}
	return nil
}

func (c *Client) doSeedRoles(ctx context.Context, token, tenantID string) error {
	body, err := json.Marshal(map[string]string{"tenant_id": tenantID})
	if err != nil {
		return fmt.Errorf("marshal richiesta seed ruoli: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/permissions/seed-roles", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("forgecore-permissions seed-roles risposta %d", resp.StatusCode)
	}
	return nil
}
