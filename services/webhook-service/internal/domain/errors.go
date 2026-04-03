package domain

import "errors"

var (
	ErrEndpointNotFound   = errors.New("endpoint webhook non trovato")
	ErrDeliveryNotFound   = errors.New("consegna webhook non trovata")
	ErrEndpointInactive   = errors.New("endpoint webhook non attivo")
	ErrMaxAttemptsReached = errors.New("numero massimo tentativi raggiunto")
)
