package application

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/domain"
)

type DeliverInput struct {
	TenantID  uuid.UUID
	EventType string
	Payload   []byte
}

type DeliverUseCase struct {
	endpoints  domain.EndpointRepository
	deliveries domain.DeliveryRepository
	client     *http.Client
}

func NewDeliverUseCase(endpoints domain.EndpointRepository, deliveries domain.DeliveryRepository) *DeliverUseCase {
	return &DeliverUseCase{
		endpoints:  endpoints,
		deliveries: deliveries,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (uc *DeliverUseCase) Execute(ctx context.Context, input DeliverInput) error {
	endpoints, err := uc.endpoints.ListActiveByEvent(ctx, input.EventType)
	if err != nil {
		return fmt.Errorf("lettura endpoint fallita: %w", err)
	}
	for _, ep := range endpoints {
		delivery := &domain.WebhookDelivery{
			ID:         uuid.New(),
			TenantID:   input.TenantID,
			EndpointID: ep.ID,
			EventType:  input.EventType,
			Payload:    input.Payload,
			Status:     domain.DeliveryStatusPending,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		deliveryErr := uc.deliver(ctx, ep, delivery)
		now := time.Now().UTC()
		if deliveryErr != nil {
			delivery.Status = domain.DeliveryStatusFailed
			delivery.LastError = deliveryErr.Error()
		} else {
			delivery.Status = domain.DeliveryStatusDelivered
			delivery.DeliveredAt = &now
		}
		delivery.Attempts++
		delivery.UpdatedAt = now
		if err := uc.deliveries.Create(ctx, delivery); err != nil {
			return fmt.Errorf("salvataggio consegna webhook fallito: %w", err)
		}
	}
	return nil
}

func (uc *DeliverUseCase) deliver(ctx context.Context, ep *domain.WebhookEndpoint, d *domain.WebhookDelivery) error {
	sig := computeHMAC(ep.Secret, d.Payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(d.Payload))
	if err != nil {
		return fmt.Errorf("creazione richiesta fallita: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("X-Webhook-Event", d.EventType)
	resp, err := uc.client.Do(req)
	if err != nil {
		return fmt.Errorf("consegna webhook fallita: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("endpoint ha risposto con status %d", resp.StatusCode)
	}
	return nil
}

func computeHMAC(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload) // hash.Hash.Write never returns a non-nil error per spec
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
