package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/domain"
)

const presignTTLSeconds = 3600

type GeneratePresignedInput struct {
	FileID   uuid.UUID
	TenantID uuid.UUID
}

type GeneratePresignedUseCase struct {
	files   domain.FileRepository
	storage domain.StorageProvider
}

func NewGeneratePresignedUseCase(files domain.FileRepository, storage domain.StorageProvider) *GeneratePresignedUseCase {
	return &GeneratePresignedUseCase{files: files, storage: storage}
}

func (uc *GeneratePresignedUseCase) Execute(ctx context.Context, input GeneratePresignedInput) (*domain.PresignedURL, error) {
	f, err := uc.files.GetByID(ctx, input.FileID, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("file non trovato: %w", err)
	}
	url, err := uc.storage.Presign(ctx, f.Bucket, f.Key, presignTTLSeconds)
	if err != nil {
		return nil, fmt.Errorf("generazione URL presigned fallita: %w", err)
	}
	return url, nil
}
