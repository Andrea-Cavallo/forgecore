package domain

import "errors"

var (
	ErrFileNotFound    = errors.New("file non trovato")
	ErrFileTooLarge    = errors.New("file supera dimensione massima consentita")
	ErrInvalidBucket   = errors.New("bucket non valido")
	ErrStorageProvider = errors.New("errore provider storage")
)
