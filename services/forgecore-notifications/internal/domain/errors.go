package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notifica non trovata")
	ErrProviderUnavailable  = errors.New("provider notifiche non disponibile")
	ErrInvalidChannel       = errors.New("canale notifica non valido")
	ErrMaxAttemptsReached   = errors.New("numero massimo di tentativi raggiunto")
)
