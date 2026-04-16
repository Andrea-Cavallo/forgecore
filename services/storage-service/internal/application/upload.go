package application

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/domain"
)

type UploadInput struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Bucket      string
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

func (i UploadInput) Validate() error {
	if i.Bucket == "" {
		return fmt.Errorf("bucket obbligatorio")
	}
	if i.Filename == "" {
		return fmt.Errorf("nome file obbligatorio")
	}
	if i.Size > domain.MaxFileSizeBytes {
		return fmt.Errorf("file supera %d bytes", domain.MaxFileSizeBytes)
	}
	return nil
}

type UploadOutput struct {
	File *domain.File
}

type UploadUseCase struct {
	files   domain.FileRepository
	storage domain.StorageProvider
}

func NewUploadUseCase(files domain.FileRepository, storage domain.StorageProvider) *UploadUseCase {
	return &UploadUseCase{files: files, storage: storage}
}

func (uc *UploadUseCase) Execute(ctx context.Context, input UploadInput) (*UploadOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	fileID := uuid.New()
	key := fmt.Sprintf("%s/%s/%s", input.TenantID, input.UserID, fileID)
	if err := uc.storage.Upload(ctx, input.Bucket, key, input.Reader, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("upload fallito: %w", err)
	}
	f := &domain.File{
		ID:          fileID,
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Bucket:      input.Bucket,
		Key:         key,
		Filename:    input.Filename,
		ContentType: input.ContentType,
		Size:        input.Size,
		CreatedAt:   time.Now().UTC(),
	}
	if err := uc.files.Save(ctx, f); err != nil {
		return nil, fmt.Errorf("salvataggio metadati file fallito: %w", err)
	}
	return &UploadOutput{File: f}, nil
}
