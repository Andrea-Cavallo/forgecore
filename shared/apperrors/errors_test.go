package apperrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsCode(t *testing.T) {
	err := Wrap(CodeNotFound, "risorsa non trovata", errors.New("missing"))
	if !IsCode(err, CodeNotFound) {
		t.Fatal("codice applicativo non rilevato")
	}
}

func TestHTTPStatus(t *testing.T) {
	if got := HTTPStatus(CodeConflict); got != http.StatusConflict {
		t.Fatalf("status inatteso: %d", got)
	}
	if got := HTTPStatus(Code("custom")); got != http.StatusInternalServerError {
		t.Fatalf("fallback inatteso: %d", got)
	}
}
