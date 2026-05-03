package domain

import "errors"

var (
	ErrPaymentNotFound      = errors.New("pagamento non trovato")
	ErrPaymentNotRefundable = errors.New("pagamento non rimborsabile")
	ErrProviderError        = errors.New("errore provider pagamento")
	ErrInvalidAmount        = errors.New("importo non valido")
)
