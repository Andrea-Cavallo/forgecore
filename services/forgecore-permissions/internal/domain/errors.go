package domain

import "errors"

var (
	ErrPermissionDenied   = errors.New("permesso negato")
	ErrRoleNotFound       = errors.New("ruolo non trovato")
	ErrPermissionNotFound = errors.New("permesso non trovato")
	ErrRoleAlreadyExists  = errors.New("ruolo già esistente")
)
