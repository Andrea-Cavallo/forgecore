// Package http exposes the payment-service REST API.
package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/payment-service/internal/application"
	"github.com/yourorg/golang-modules/shared/pagination"
)

const (
	headerTenantID  = "X-Tenant-ID"
	headerUserID    = "X-User-ID"
	headerStripesig = "Stripe-Signature"
	maxWebhookBytes = 1 << 20 // 1 MB
)

// Handler holds all HTTP handlers for the payment service.
type Handler struct {
	createPayment *application.CreatePaymentUseCase
	refund        *application.RefundUseCase
	webhook       *application.HandleStripeWebhookUseCase
}

// NewHandler constructs the payment HTTP handler.
func NewHandler(
	cp *application.CreatePaymentUseCase,
	rf *application.RefundUseCase,
	wh *application.HandleStripeWebhookUseCase,
) *Handler {
	return &Handler{createPayment: cp, refund: rf, webhook: wh}
}

// RegisterRoutes mounts the payment routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/payments", h.createPaymentHandler)
	mux.HandleFunc("POST /v1/payments/{id}/refund", h.refundHandler)
	mux.HandleFunc("GET /v1/payments", h.listPaymentsHandler)
	mux.HandleFunc("POST /v1/webhooks/stripe", h.stripeWebhookHandler)
}

// ---- handlers ----

func (h *Handler) createPaymentHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractIDs(w, r)
	if !ok {
		return
	}

	var body struct {
		Amount     int64  `json:"amount"`
		Currency   string `json:"currency"`
		CustomerID string `json:"customer_id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	out, err := h.createPayment.Execute(r.Context(), application.CreatePaymentInput{
		TenantID:   tenantID,
		UserID:     userID,
		Amount:     body.Amount,
		Currency:   body.Currency,
		CustomerID: body.CustomerID,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out.Payment)
}

func (h *Handler) refundHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := extractIDs(w, r)
	if !ok {
		return
	}

	paymentID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id pagamento non valido")
		return
	}

	var body struct {
		Amount int64 `json:"amount"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	if err := h.refund.Execute(r.Context(), application.RefundInput{
		PaymentID: paymentID,
		TenantID:  tenantID,
		Amount:    body.Amount,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"payments": []any{},
		"cursor":   pagination.Cursor{},
	})
}

func (h *Handler) stripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "lettura body fallita")
		return
	}
	sig := r.Header.Get(headerStripesig)
	if err := h.webhook.Execute(r.Context(), application.HandleWebhookInput{
		RawBody:   body,
		SigHeader: sig,
	}); err != nil {
		slog.Error("webhook stripe fallito", "errore", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---- helpers ----

func extractIDs(w http.ResponseWriter, r *http.Request) (tenantID, userID uuid.UUID, ok bool) {
	tenantID, err := uuid.Parse(r.Header.Get(headerTenantID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "tenant_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err = uuid.Parse(r.Header.Get(headerUserID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, userID, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "corpo richiesta non valido")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
