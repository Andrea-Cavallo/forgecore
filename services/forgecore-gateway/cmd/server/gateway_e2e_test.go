package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/clients/authgrpc"
	"google.golang.org/grpc"
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

func TestGatewaySecurityE2E_ForbiddenOrigin(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()

	handler := newTestGatewayHandler(t, upstream.URL)
	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestGatewaySecurityE2E_MissingTokenUsesEnvelope(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()

	handler := newTestGatewayHandler(t, upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	req.Header.Set("X-Request-ID", "req-test")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"missing_token"`) {
		t.Fatalf("body inatteso: %s", rec.Body.String())
	}
}

func TestGatewaySecurityE2E_RBACDeniesUnauthorizedRole(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("upstream non deve essere chiamato senza ruolo autorizzato")
	}))
	defer upstream.Close()

	authServer := startTestAuthGRPC(t, []string{"user"})
	authClient, err := authgrpc.NewClient(authServer)
	if err != nil {
		t.Fatalf("client auth test: %v", err)
	}
	t.Cleanup(func() {
		if err := authClient.Close(); err != nil {
			t.Fatalf("chiusura client auth test: %v", err)
		}
	})

	handler, err := buildHandler(testConfig(upstream.URL), authClient)
	if err != nil {
		t.Fatalf("costruzione gateway test: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/00000000-0000-0000-0000-000000000001/disable", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-Request-ID", "req-rbac")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"forbidden"`) {
		t.Fatalf("body inatteso: %s", rec.Body.String())
	}
}

func TestGatewaySecurityE2E_RBACAllowsAuthorizedRole(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	authServer := startTestAuthGRPC(t, []string{"admin"})
	authClient, err := authgrpc.NewClient(authServer)
	if err != nil {
		t.Fatalf("client auth test: %v", err)
	}
	t.Cleanup(func() {
		if err := authClient.Close(); err != nil {
			t.Fatalf("chiusura client auth test: %v", err)
		}
	})

	handler, err := buildHandler(testConfig(upstream.URL), authClient)
	if err != nil {
		t.Fatalf("costruzione gateway test: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/users/00000000-0000-0000-0000-000000000001/disable", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
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

func startTestAuthGRPC(t *testing.T, roles []string) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listener auth test: %v", err)
	}
	server := grpc.NewServer()
	registerTestAuthService(server, func(_ context.Context, token string) (*authgrpc.ValidateTokenResponse, error) {
		if token != "valid-token" {
			return &authgrpc.ValidateTokenResponse{Valid: false}, nil
		}
		return &authgrpc.ValidateTokenResponse{
			Valid:    true,
			UserID:   "00000000-0000-0000-0000-000000000010",
			TenantID: "00000000-0000-0000-0000-000000000020",
			Roles:    roles,
		}, nil
	})
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(server.Stop)
	return lis.Addr().String()
}

type testAuthValidateFunc func(context.Context, string) (*authgrpc.ValidateTokenResponse, error)

type testAuthService struct {
	validate testAuthValidateFunc
}

type testAuthServiceServer interface {
	ValidateToken(context.Context, *testValidateTokenRequest) (*authgrpc.ValidateTokenResponse, error)
}

type testValidateTokenRequest struct {
	Token string `json:"token"`
}

func (s *testAuthService) ValidateToken(ctx context.Context, req *testValidateTokenRequest) (*authgrpc.ValidateTokenResponse, error) {
	return s.validate(ctx, req.Token)
}

func registerTestAuthService(server *grpc.Server, validate testAuthValidateFunc) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "auth.v1.AuthService",
		HandlerType: (*testAuthServiceServer)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "ValidateToken",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				var req testValidateTokenRequest
				if err := dec(&req); err != nil {
					return nil, err
				}
				return srv.(testAuthServiceServer).ValidateToken(ctx, &req)
			},
		}},
	}, &testAuthService{validate: validate})
}
