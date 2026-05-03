package domain

import "errors"

var (
	ErrEntryNotFound    = errors.New("voce audit non trovata")
	ErrInvalidActorType = errors.New("tipo attore non valido")
)
