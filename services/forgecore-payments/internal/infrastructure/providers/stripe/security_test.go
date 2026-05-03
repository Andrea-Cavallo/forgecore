package stripe

import "testing"

func TestVerifyWebhookRejectsInvalidSignature(t *testing.T) {
	provider := NewProvider("sk_test", "whsec_test")
	_, err := provider.VerifyWebhook([]byte(`{"id":"evt_1"}`), "t=1,v1=invalid")
	if err == nil {
		t.Fatalf("firma webhook non valida accettata")
	}
}
