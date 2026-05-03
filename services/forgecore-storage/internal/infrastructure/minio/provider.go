// Package minio implements domain.StorageProvider using the MinIO Go SDK.
package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/domain"
)

// Provider implements domain.StorageProvider backed by MinIO.
type Provider struct {
	client *miniogo.Client
}

// NewProvider constructs a MinIO storage provider.
func NewProvider(endpoint, accessKey, secretKey string, useSSL bool) (*Provider, error) {
	client, err := miniogo.New(endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &Provider{client: client}, nil
}

// Upload streams data to MinIO.
func (p *Provider) Upload(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) error {
	_, err := p.client.PutObject(ctx, bucket, key, r, size, miniogo.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio upload: %w", err)
	}
	return nil
}

// Delete removes an object from MinIO.
func (p *Provider) Delete(ctx context.Context, bucket, key string) error {
	if err := p.client.RemoveObject(ctx, bucket, key, miniogo.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio delete: %w", err)
	}
	return nil
}

// Presign generates a pre-signed GET URL with the given TTL.
func (p *Provider) Presign(ctx context.Context, bucket, key string, ttlSeconds int64) (*domain.PresignedURL, error) {
	expiry := time.Duration(ttlSeconds) * time.Second
	u, err := p.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return nil, fmt.Errorf("minio presign: %w", err)
	}
	return &domain.PresignedURL{
		URL:       u.String(),
		ExpiresAt: time.Now().UTC().Add(expiry),
	}, nil
}
