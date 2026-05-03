package domain

type PaymentWebhookEvent struct {
	ID                string
	Type              string
	ProviderPaymentID string
	Status            string
	Amount            int64
}
