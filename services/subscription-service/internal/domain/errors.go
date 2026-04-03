package domain

import "errors"

var (
	ErrSubscriptionNotFound  = errors.New("abbonamento non trovato")
	ErrPlanNotFound          = errors.New("piano non trovato")
	ErrSubscriptionNotActive = errors.New("abbonamento non attivo")
	ErrAlreadySubscribed     = errors.New("utente già abbonato a questo piano")
	ErrProviderError         = errors.New("errore provider pagamenti")
)
