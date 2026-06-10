package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rmalik/sandbox-payments/internal/payment"
)

const maxAttempts = 3

// Payload is the JSON body sent to merchant webhook endpoints.
type Payload struct {
	Event     string `json:"event"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

// Sender delivers payment status events to merchant webhooks.
type Sender struct {
	client *http.Client
	log    *slog.Logger
}

// NewSender creates a webhook sender with sensible defaults.
func NewSender(log *slog.Logger) *Sender {
	return &Sender{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
	}
}

// ITNPayload is sent when the gateway releases the payer without final confirmation.
type ITNPayload struct {
	Event     string `json:"event"`
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
}

// FinalizeFunc runs after a successful ITN delivery (e.g. confirm payment).
type FinalizeFunc func(ctx context.Context, p *payment.Payment) error

// NotifyITN sends an instant transaction notification after an optional delay.
func (s *Sender) NotifyITN(p *payment.Payment, finalize FinalizeFunc) {
	if p.ITNURL == "" {
		return
	}

	payload := ITNPayload{
		Event:     "payment.processing",
		PaymentID: p.ID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    string(payment.StatusProcessing),
	}

	delay := time.Duration(p.ITNDelay) * time.Second
	go s.deliverITN(context.Background(), p.ITNURL, payload, delay, p, finalize)
}

func (s *Sender) deliverITN(ctx context.Context, url string, payload ITNPayload, delay time.Duration, p *payment.Payment, finalize FinalizeFunc) {
	if delay > 0 {
		s.log.Info("itn scheduled", "payment_id", p.ID, "delay", delay)
		time.Sleep(delay)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.log.Error("itn marshal failed", "error", err)
		return
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := s.deliver(ctx, url, body); err == nil {
			s.log.Info("itn delivered", "url", url, "payment_id", p.ID)
			if finalize != nil {
				if err := finalize(ctx, p); err != nil {
					s.log.Error("itn finalize failed", "payment_id", p.ID, "error", err)
				}
			}
			return
		}
		s.log.Warn("itn attempt failed", "url", url, "payment_id", p.ID, "attempt", attempt, "error", err)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	s.log.Error("itn delivery exhausted retries", "url", url, "payment_id", p.ID)
}

// NotifyPayment sends a webhook for a terminal payment status (async, with retries).
func (s *Sender) NotifyPayment(p *payment.Payment) {
	if p.WebhookURL == "" {
		return
	}

	event := eventForStatus(p.Status)
	if event == "" {
		return
	}

	payload := Payload{
		Event:     event,
		PaymentID: p.ID,
		Amount:    p.Amount,
		Currency:  p.Currency,
	}

	go s.deliverWithRetry(context.Background(), p.WebhookURL, payload)
}

func (s *Sender) deliverWithRetry(ctx context.Context, url string, payload Payload) {
	body, err := json.Marshal(payload)
	if err != nil {
		s.log.Error("webhook marshal failed", "error", err)
		return
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := s.deliver(ctx, url, body); err == nil {
			s.log.Info("webhook delivered", "url", url, "event", payload.Event, "payment_id", payload.PaymentID)
			return
		} else {
			s.log.Warn("webhook attempt failed",
				"url", url,
				"event", payload.Event,
				"payment_id", payload.PaymentID,
				"attempt", attempt,
				"error", err,
			)
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	s.log.Error("webhook delivery exhausted retries",
		"url", url,
		"event", payload.Event,
		"payment_id", payload.PaymentID,
	)
}

func (s *Sender) deliver(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "sandbox-payments/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func eventForStatus(status payment.Status) string {
	switch status {
	case payment.StatusSucceeded:
		return "payment.succeeded"
	case payment.StatusFailed:
		return "payment.failed"
	case payment.StatusCancelled:
		return "payment.cancelled"
	default:
		return ""
	}
}
