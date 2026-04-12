// Package email provides email delivery via the SendGrid REST API.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	sendGridAPIBase = "https://api.sendgrid.com/v3/mail/send"
	defaultTimeout  = 10 * time.Second
)

// SendGridProvider implements domain.EmailProvider using the SendGrid v3 REST API.
type SendGridProvider struct {
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient *http.Client
}

// NewSendGridProvider constructs a SendGrid email provider.
func NewSendGridProvider(apiKey, fromEmail, fromName string) *SendGridProvider {
	return &SendGridProvider{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		fromName:   fromName,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

// Send delivers an email via SendGrid. The template parameter is the email subject;
// vars must contain a "body" key with the plain-text / HTML body.
func (p *SendGridProvider) Send(ctx context.Context, to, template string, vars map[string]string) error {
	body := vars["body"]
	payload := buildPayload(p.fromEmail, p.fromName, to, template, body)
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendGridAPIBase, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creazione richiesta sendgrid: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("invio email sendgrid: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid risposta non-ok: %d", resp.StatusCode)
	}
	return nil
}

// ---- internal DTO ----

type sgPayload struct {
	Personalizations []sgPersonalization `json:"personalizations"`
	From             sgAddress            `json:"from"`
	Subject          string               `json:"subject"`
	Content          []sgContent          `json:"content"`
}

type sgPersonalization struct {
	To []sgAddress `json:"to"`
}

type sgAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sgContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func buildPayload(fromEmail, fromName, toEmail, subject, body string) sgPayload {
	return sgPayload{
		From:    sgAddress{Email: fromEmail, Name: fromName},
		Subject: subject,
		Personalizations: []sgPersonalization{
			{To: []sgAddress{{Email: toEmail}}},
		},
		Content: []sgContent{
			{Type: "text/plain", Value: body},
		},
	}
}
