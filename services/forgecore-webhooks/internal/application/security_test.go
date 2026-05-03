package application

import "testing"

func TestValidateWebhookURLRejectsSSRFTargets(t *testing.T) {
	cases := []string{
		"http://example.com/webhook",
		"https://127.0.0.1/webhook",
		"https://localhost/webhook",
	}
	for _, raw := range cases {
		if err := validateWebhookURL(raw); err == nil {
			t.Fatalf("URL SSRF accettato: %s", raw)
		}
	}
}
