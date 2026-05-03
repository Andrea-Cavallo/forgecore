package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/clients/authgrpc"
)

const testOrigin = "https://app.forgecore.local"

func TestGatewayFrontendE2E_PreflightAndPublicProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			t.Fatalf("path upstream inatteso: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"token":"test-token"}`)); err != nil {
			t.Fatalf("scrittura risposta upstream: %v", err)
		}
	}))
	defer upstream.Close()

	handler := newTestGatewayHandler(t, upstream.URL)
	assertPreflight(t, handler)
	assertPublicProxy(t, handler)
}

func TestGatewayFrontendE2E_HealthAndReadiness(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()

	handler := newTestGatewayHandler(t, upstream.URL)
	assertStatus(t, handler, "/healthz", http.StatusOK)
	assertStatus(t, handler, "/readyz", http.StatusOK)
}

func newTestGatewayHandler(t *testing.T, upstreamURL string) http.Handler {
	t.Helper()
	authClient, err := authgrpc.NewClient("127.0.0.1:1")
	if err != nil {
		t.Fatalf("client auth test: %v", err)
	}
	t.Cleanup(func() {
		if err := authClient.Close(); err != nil {
			t.Fatalf("chiusura client auth test: %v", err)
		}
	})

	handler, err := buildHandler(testConfig(upstreamURL), authClient)
	if err != nil {
		t.Fatalf("costruzione gateway test: %v", err)
	}
	return handler
}

func testConfig(upstreamURL string) config {
	return config{
		addr:                 ":0",
		authGRPCAddr:         "127.0.0.1:1",
		authServiceURL:       upstreamURL,
		paymentServiceURL:    upstreamURL,
		notifServiceURL:      upstreamURL,
		permissionServiceURL: upstreamURL,
		configServiceURL:     upstreamURL,
		webhookServiceURL:    upstreamURL,
		storageServiceURL:    upstreamURL,
		subsServiceURL:       upstreamURL,
		adminServiceURL:      upstreamURL,
		auditServiceURL:      upstreamURL,
		corsOrigin:           testOrigin,
		rateLimit:            100,
	}
}

func assertPreflight(t *testing.T, handler http.Handler) {
	t.Helper()
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", testOrigin)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d", rec.Code)
	}
	assertHeader(t, rec, "Access-Control-Allow-Origin", testOrigin)
	assertHeader(t, rec, "X-Frame-Options", "DENY")
	assertHeader(t, rec, "X-Content-Type-Options", "nosniff")
}

func assertPublicProxy(t *testing.T, handler http.Handler) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.Header.Set("Origin", testOrigin)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("proxy status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertHeader(t, rec, "Access-Control-Allow-Origin", testOrigin)
}

func assertStatus(t *testing.T, handler http.Handler, path string, want int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != want {
		t.Fatalf("%s status = %d", path, rec.Code)
	}
}

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, key string, want string) {
	t.Helper()
	if got := rec.Header().Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
