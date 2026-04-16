package application

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/webhook-service/internal/domain"
)

type RegisterEndpointInput struct {
	TenantID uuid.UUID
	URL      string
	Secret   string
	Events   []string
}

func (i RegisterEndpointInput) Validate() error {
	if i.URL == "" {
		return fmt.Errorf("URL obbligatorio")
	}
	if err := validateWebhookURL(i.URL); err != nil {
		return err
	}
	if len(i.Events) == 0 {
		return fmt.Errorf("almeno un evento richiesto")
	}
	return nil
}

// validateWebhookURL rejects non-HTTPS schemes and private/loopback IP targets (SSRF prevention).
func validateWebhookURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("URL non valido: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("sono ammessi solo URL HTTPS")
	}
	host := u.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil {
		// Allow resolution failures at registration time; block on known bad hosts.
		return nil
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("URL punta a indirizzo IP privato o loopback non consentito")
		}
	}
	return nil
}

type RegisterEndpointOutput struct {
	Endpoint *domain.WebhookEndpoint
}

type RegisterEndpointUseCase struct {
	endpoints domain.EndpointRepository
}

func NewRegisterEndpointUseCase(endpoints domain.EndpointRepository) *RegisterEndpointUseCase {
	return &RegisterEndpointUseCase{endpoints: endpoints}
}

func (uc *RegisterEndpointUseCase) Execute(ctx context.Context, input RegisterEndpointInput) (*RegisterEndpointOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	endpoint := &domain.WebhookEndpoint{
		ID:        uuid.New(),
		TenantID:  input.TenantID,
		URL:       input.URL,
		Secret:    input.Secret,
		Events:    input.Events,
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := uc.endpoints.Create(ctx, endpoint); err != nil {
		return nil, fmt.Errorf("registrazione endpoint fallita: %w", err)
	}
	return &RegisterEndpointOutput{Endpoint: endpoint}, nil
}
