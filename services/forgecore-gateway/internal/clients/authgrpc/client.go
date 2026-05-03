// Package authgrpc provides a gRPC client for the forgecore-auth ValidateToken RPC.
// It uses the same JSON codec override as the server — no protoc step required.
package authgrpc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

func init() {
	// Mirror the codec registered on the forgecore-auth gRPC server.
	encoding.RegisterCodec(jsonCodec{})
}

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error)      { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                               { return "proto" }

// --- Request / Response types (must match forgecore-auth server.go) ---

type validateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse holds the result of a ValidateToken RPC call.
type ValidateTokenResponse struct {
	Valid       bool     `json:"valid"`
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Roles       []string `json:"roles"`
	MFAVerified bool     `json:"mfa_verified"`
}

// Client wraps a gRPC connection to forgecore-auth.
type Client struct {
	conn *grpc.ClientConn
}

// NewClient dials the forgecore-auth gRPC endpoint at addr.
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connessione gRPC auth fallita su %s: %w", addr, err)
	}
	return &Client{conn: conn}, nil
}

// Close shuts down the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ValidateToken calls forgecore-auth.ValidateToken and returns the claims.
func (c *Client) ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error) {
	var resp ValidateTokenResponse
	err := c.conn.Invoke(ctx, "/auth.v1.AuthService/ValidateToken",
		&validateTokenRequest{Token: token}, &resp)
	if err != nil {
		return nil, fmt.Errorf("validazione token: %w", err)
	}
	return &resp, nil
}
