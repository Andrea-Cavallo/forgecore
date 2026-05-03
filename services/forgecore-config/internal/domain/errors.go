package domain

import "errors"

var (
	ErrConfigNotFound = errors.New("configurazione non trovata")
	ErrInvalidKey     = errors.New("chiave configurazione non valida")
	ErrInvalidValue   = errors.New("valore configurazione non valido")
)
