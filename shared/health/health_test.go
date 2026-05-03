package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerReturnsReady(t *testing.T) {
	handler := Handler("forgecore-test", map[string]Check{
		"db": func(context.Context) error { return nil },
	})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"service":"forgecore-test"`) {
		t.Fatalf("body inatteso: %s", rec.Body.String())
	}
}

func TestHandlerReturnsUnavailable(t *testing.T) {
	handler := Handler("forgecore-test", map[string]Check{
		"db": func(context.Context) error { return errors.New("database non disponibile") },
	})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"fail"`) {
		t.Fatalf("body inatteso: %s", rec.Body.String())
	}
}
