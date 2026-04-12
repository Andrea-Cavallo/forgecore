// Package grpc provides the internal gRPC server for ValidateToken and GetUser.
// Encoding: JSON codec (both server and clients must set the same codec).
// Proto definition: shared/proto/auth.proto
// To regenerate from proto: protoc --go_out=. --go-grpc_out=. shared/proto/auth.proto
package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"

	"github.com/yourorg/golang-modules/services/auth-service/internal/application"
	"github.com/yourorg/golang-modules/services/auth-service/internal/domain"
)

func init() {
	// Override the default proto codec with JSON so no protoc step is required.
	// Replace this with proper proto-generated code when protoc is available.
	encoding.RegisterCodec(jsonCodec{})
}

// jsonCodec serializes gRPC messages as JSON instead of protobuf.
type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error)      { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                               { return "proto" } // override default codec name

// --- Request / Response types (mirrors auth.proto) ---

// ValidateTokenRequest is the gRPC request for ValidateToken.
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse is the gRPC response for ValidateToken.
type ValidateTokenResponse struct {
	Valid        bool     `json:"valid"`
	UserID       string   `json:"user_id"`
	TenantID     string   `json:"tenant_id"`
	Roles        []string `json:"roles"`
	MFAVerified  bool     `json:"mfa_verified"`
}

// GetUserRequest is the gRPC request for GetUser.
type GetUserRequest struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// GetUserResponse is the gRPC response for GetUser.
type GetUserResponse struct {
	UserID        string   `json:"user_id"`
	TenantID      string   `json:"tenant_id"`
	EmailVerified bool     `json:"email_verified"`
	MFAEnabled    bool     `json:"mfa_enabled"`
	Roles         []string `json:"roles"`
	IsDeleted     bool     `json:"is_deleted"`
}

// --- Server ---

// AuthServer implements the auth gRPC service.
type AuthServer struct {
	jwtSvc     *application.JWTService
	tokenStore domain.TokenStore
	userRepo   domain.UserRepository
}

// NewAuthServer creates an AuthServer with the provided dependencies.
func NewAuthServer(jwtSvc *application.JWTService, tokenStore domain.TokenStore, userRepo domain.UserRepository) *AuthServer {
	return &AuthServer{jwtSvc: jwtSvc, tokenStore: tokenStore, userRepo: userRepo}
}

// validateToken validates a JWT and checks it against the blacklist.
func (s *AuthServer) validateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	claims, err := s.jwtSvc.Validate(req.Token)
	if err != nil {
		return &ValidateTokenResponse{Valid: false}, nil
	}
	blacklisted, err := s.tokenStore.IsBlacklisted(ctx, claims.JTI)
	if err != nil || blacklisted {
		return &ValidateTokenResponse{Valid: false}, nil
	}
	return &ValidateTokenResponse{
		Valid:       true,
		UserID:      claims.UserID.String(),
		TenantID:    claims.TenantID.String(),
		Roles:       claims.Roles,
		MFAVerified: claims.MFAVerified,
	}, nil
}

// getUser retrieves basic user metadata.
func (s *AuthServer) getUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user_id non valido")
	}
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id non valido")
	}
	user, err := s.userRepo.GetByID(ctx, userID, tenantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "utente non trovato")
	}
	return &GetUserResponse{
		UserID:        user.ID.String(),
		TenantID:      user.TenantID.String(),
		EmailVerified: user.EmailVerified,
		MFAEnabled:    user.MFAEnabled,
		Roles:         user.Roles,
		IsDeleted:     user.IsDeleted(),
	}, nil
}

// --- Service descriptor (replaces protoc-generated registration) ---

// authServiceHandler is the interface used for gRPC service descriptor type checking.
// AuthServer implements it; the interface ensures RegisterService doesn't panic.
type authServiceHandler interface {
	validateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error)
	getUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
}

func _ValidateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(authServiceHandler).validateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.AuthService/ValidateToken"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(authServiceHandler).validateToken(ctx, req.(*ValidateTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(authServiceHandler).getUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.v1.AuthService/GetUser"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(authServiceHandler).getUser(ctx, req.(*GetUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var authServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.v1.AuthService",
	HandlerType: (*authServiceHandler)(nil), // interface pointer for reflect.Implements check
	Methods: []grpc.MethodDesc{
		{MethodName: "ValidateToken", Handler: _ValidateToken_Handler},
		{MethodName: "GetUser", Handler: _GetUser_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth/auth.proto",
}

// Register adds the auth gRPC service to the server.
func Register(s *grpc.Server, srv *AuthServer) {
	s.RegisterService(&authServiceDesc, srv)
}

// ListenAndServe starts the gRPC server on addr and blocks until an error occurs.
func ListenAndServe(addr string, srv *AuthServer, opts ...grpc.ServerOption) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("ascolto gRPC fallito su %s: %w", addr, err)
	}
	s := grpc.NewServer(opts...)
	Register(s, srv)
	return s.Serve(lis)
}
