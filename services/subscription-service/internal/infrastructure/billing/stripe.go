// Package billing implements domain.BillingProvider using the Stripe REST API.
package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	stripeAPIBase   = "https://api.stripe.com/v1"
	billingTimeout  = 15 * time.Second
)

// StripeProvider implements domain.BillingProvider using Stripe Subscriptions API.
type StripeProvider struct {
	secretKey  string
	httpClient *http.Client
}

// NewStripeProvider constructs a StripeProvider.
func NewStripeProvider(secretKey string) *StripeProvider {
	return &StripeProvider{
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: billingTimeout},
	}
}

// CreateSubscription creates a Stripe Subscription and returns its ID.
func (p *StripeProvider) CreateSubscription(ctx context.Context, customerID, planProviderID string) (string, error) {
	body := url.Values{
		"customer":           {customerID},
		"items[0][price]":    {planProviderID},
	}
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.post(ctx, "/subscriptions", body, &result); err != nil {
		return "", fmt.Errorf("creazione abbonamento stripe: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("stripe: %s", result.Error.Message)
	}
	return result.ID, nil
}

// CancelSubscription cancels a Stripe Subscription at period end.
func (p *StripeProvider) CancelSubscription(ctx context.Context, providerID string) error {
	body := url.Values{"cancel_at_period_end": {"true"}}
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.post(ctx, "/subscriptions/"+providerID, body, &result); err != nil {
		return fmt.Errorf("cancellazione abbonamento stripe: %w", err)
	}
	if result.Error != nil {
		return fmt.Errorf("stripe: %s", result.Error.Message)
	}
	return nil
}

// ChangeSubscriptionPlan updates the subscription to a new price/plan.
func (p *StripeProvider) ChangeSubscriptionPlan(ctx context.Context, providerID, newPlanID string) error {
	// Retrieve current subscription items first.
	var sub struct {
		Items struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"items"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.get(ctx, "/subscriptions/"+providerID, &sub); err != nil {
		return fmt.Errorf("lettura abbonamento stripe: %w", err)
	}
	if sub.Error != nil {
		return fmt.Errorf("stripe: %s", sub.Error.Message)
	}
	if len(sub.Items.Data) == 0 {
		return fmt.Errorf("nessun item nell'abbonamento")
	}
	body := url.Values{
		"items[0][id]":    {sub.Items.Data[0].ID},
		"items[0][price]": {newPlanID},
	}
	var result struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.post(ctx, "/subscriptions/"+providerID, body, &result); err != nil {
		return fmt.Errorf("cambio piano stripe: %w", err)
	}
	if result.Error != nil {
		return fmt.Errorf("stripe: %s", result.Error.Message)
	}
	return nil
}

func (p *StripeProvider) post(ctx context.Context, path string, body url.Values, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, stripeAPIBase+path, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (p *StripeProvider) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stripeAPIBase+path, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.secretKey, "")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}
