// Package sms provides SMS delivery via the Twilio REST API.
package sms

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	twilioAPIBase  = "https://api.twilio.com/2010-04-01/Accounts"
	twilioTimeout  = 10 * time.Second
)

// TwilioProvider implements domain.SMSProvider using the Twilio Messages API.
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// NewTwilioProvider constructs a TwilioProvider.
func NewTwilioProvider(accountSID, authToken, fromNumber string) *TwilioProvider {
	return &TwilioProvider{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{Timeout: twilioTimeout},
	}
}

// Send delivers an SMS. The template parameter is ignored; vars["body"] is the message text.
func (p *TwilioProvider) Send(ctx context.Context, to, _ string, vars map[string]string) error {
	endpoint := fmt.Sprintf("%s/%s/Messages.json", twilioAPIBase, p.accountSID)
	body := url.Values{
		"From": {p.fromNumber},
		"To":   {to},
		"Body": {vars["body"]},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("creazione richiesta twilio: %w", err)
	}
	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("invio SMS twilio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio risposta non-ok: %d", resp.StatusCode)
	}
	return nil
}
