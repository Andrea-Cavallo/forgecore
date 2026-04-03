package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("utente non trovato")
	ErrEmailAlreadyExists = errors.New("email già in uso per questo tenant")
	ErrInvalidCredentials = errors.New("credenziali non valide")
	ErrAccountLocked      = errors.New("account temporaneamente bloccato")
	ErrTokenInvalid       = errors.New("token non valido")
	ErrTokenExpired       = errors.New("token scaduto")
	ErrMFARequired        = errors.New("autenticazione MFA richiesta")
	ErrMFAInvalid         = errors.New("codice MFA non valido")
	ErrSessionNotFound    = errors.New("sessione non trovata")
)
