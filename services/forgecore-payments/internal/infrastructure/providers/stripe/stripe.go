// Package stripe implements domain.PaymentProvider using the Stripe REST API directly.
// No official SDK required — uses net/http with Basic auth (secret key as username).
package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/domain"
)

const (
	stripeAPIBase        = "https://api.stripe.com/v1"
	webhookToleranceSecs = 300
)

// Provider implements domain.PaymentProvider against the Stripe REST API.
type Provider struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
}

// NewProvider creates a Stripe provider with the given API credentials.
func NewProvider(secretKey, webhookSecret string) *Provider {
	return &Provider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Charge creates a Stripe PaymentIntent and immediately confirms it.
// Returns the PaymentIntent ID as the provider ID.
func (p *Provider) Charge(ctx context.Context, amount int64, currency, customerID string) (string, error) {
	body := url.Values{
		"amount":                             {strconv.FormatInt(amount, 10)},
		"currency":                           {currency},
		"customer":                           {customerID},
		"confirm":                            {"true"},
		"automatic_payment_methods[enabled]": {"true"},
		"automatic_payment_methods[allow_redirects]": {"never"},
	}
	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.post(ctx, "/payment_intents", body, &result); err != nil {
		return "", fmt.Errorf("creazione payment intent: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("stripe: %s", result.Error.Message)
	}
	return result.ID, nil
}

// Refund creates a Stripe Refund for the given PaymentIntent.
func (p *Provider) Refund(ctx context.Context, providerID string, amount int64) error {
	body := url.Values{
		"payment_intent": {providerID},
		"amount":         {strconv.FormatInt(amount, 10)},
	}
	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.post(ctx, "/refunds", body, &result); err != nil {
		return fmt.Errorf("rimborso stripe: %w", err)
	}
	if result.Error != nil {
		return fmt.Errorf("stripe: %s", result.Error.Message)
	}
	return nil
}

// VerifyWebhook validates the Stripe-Signature header and returns the parsed event.
func (p *Provider) VerifyWebhook(rawBody []byte, sigHeader string) (*domain.PaymentWebhookEvent, error) {
	ts, sigs, err := parseSignatureHeader(sigHeader)
	if err != nil {
		return nil, fmt.Errorf("parsing firma webhook: %w", err)
	}
	if time.Now().Unix()-ts > webhookToleranceSecs {
		return nil, fmt.Errorf("webhook scaduto")
	}
	expected := p.computeSignature(ts, rawBody)
	if !signaturesMatch(expected, sigs) {
		return nil, fmt.Errorf("firma webhook non valida")
	}
	var event stripeEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return nil, fmt.Errorf("parsing evento stripe: %w", err)
	}
	return &domain.PaymentWebhookEvent{
		ID:                event.ID,
		Type:              event.Type,
		ProviderPaymentID: event.Data.Object.ID,
		Status:            event.Data.Object.Status,
		Amount:            event.Data.Object.Amount,
	}, nil
}

type stripeEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Amount int64  `json:"amount"`
		} `json:"object"`
	} `json:"data"`
}

func (p *Provider) post(ctx context.Context, path string, body url.Values, result interface{}) error {
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

func (p *Provider) computeSignature(ts int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(p.webhookSecret))
	_, _ = fmt.Fprintf(mac, "%d.", ts)
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func parseSignatureHeader(header string) (int64, []string, error) {
	var ts int64
	var sigs []string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			n, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("timestamp firma non valido")
			}
			ts = n
		case "v1":
			sigs = append(sigs, kv[1])
		}
	}
	if ts == 0 || len(sigs) == 0 {
		return 0, nil, fmt.Errorf("header firma incompleto")
	}
	return ts, sigs, nil
}

func signaturesMatch(expected string, sigs []string) bool {
	for _, s := range sigs {
		if hmac.Equal([]byte(expected), []byte(s)) {
			return true
		}
	}
	return false
}
