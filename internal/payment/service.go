package payment

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound       = errors.New("payment not found")
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidStatus  = errors.New("invalid payment status transition")
)

// CreateRequest holds input for initiating a new payment.
type CreateRequest struct {
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	Description   string            `json:"description"`
	PaymentMethod string            `json:"payment_method"`
	SuccessURL    string            `json:"success_url"`
	CancelURL     string            `json:"cancel_url"`
	FailedURL     string            `json:"failed_url"`
	WebhookURL    string            `json:"webhook_url"`
	ITNURL        string            `json:"itn_url"`
	ITNDelay      int               `json:"itn_delay_seconds"`
	Metadata      map[string]string `json:"metadata"`
}

// CreateResponse is returned after a payment is created.
type CreateResponse struct {
	PaymentID   string `json:"payment_id"`
	Status      Status `json:"status"`
	RedirectURL string `json:"redirect_url"`
}

// Service encapsulates payment business logic.
type Service struct {
	store   Store
	baseURL string
}

// NewService constructs a payment service.
func NewService(store Store, baseURL string) *Service {
	return &Service{store: store, baseURL: strings.TrimRight(baseURL, "/")}
}

// ListMethods returns supported payment methods.
func (s *Service) ListMethods() []Method {
	return AvailableMethods()
}

// Create validates input and persists a pending payment.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*CreateResponse, error) {
	if err := validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	metadata := "{}"
	if len(req.Metadata) > 0 {
		raw, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid metadata", ErrInvalidRequest)
		}
		metadata = string(raw)
	}

	now := time.Now().UTC()
	p := &Payment{
		ID:            newPaymentID(),
		Amount:        req.Amount,
		Currency:      strings.ToUpper(req.Currency),
		Description:   req.Description,
		Status:        StatusPending,
		PaymentMethod: req.PaymentMethod,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
		FailedURL:     req.FailedURL,
		WebhookURL:    req.WebhookURL,
		ITNURL:        req.ITNURL,
		ITNDelay:      req.ITNDelay,
		Metadata:      metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.store.Create(ctx, p); err != nil {
		return nil, err
	}

	return &CreateResponse{
		PaymentID:   p.ID,
		Status:      p.Status,
		RedirectURL: fmt.Sprintf("%s/pay/%s", s.baseURL, p.ID),
	}, nil
}

// Get returns a payment by ID.
func (s *Service) Get(ctx context.Context, id string) (*Payment, error) {
	p, err := s.store.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// SelectMethod assigns a payment method to a pending payment.
func (s *Service) SelectMethod(ctx context.Context, id, method string) (*Payment, error) {
	if !IsValidMethod(method) || method == "" {
		return nil, fmt.Errorf("%w: unknown payment method", ErrInvalidRequest)
	}

	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusPending {
		return nil, ErrInvalidStatus
	}

	p.PaymentMethod = method
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// MarkProcessing moves a pending payment to processing (gateway released, no confirmation yet).
func (s *Service) MarkProcessing(ctx context.Context, id string) (*Payment, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusPending {
		return nil, ErrInvalidStatus
	}
	if p.PaymentMethod == "" {
		return nil, fmt.Errorf("%w: payment method required", ErrInvalidRequest)
	}

	p.Status = StatusProcessing
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ConfirmAfterITN finalizes a payment after ITN delivery.
func (s *Service) ConfirmAfterITN(ctx context.Context, id string) (*Payment, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusProcessing {
		return nil, ErrInvalidStatus
	}

	p.Status = StatusSucceeded
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Transition moves a pending payment to a terminal status.
func (s *Service) Transition(ctx context.Context, id string, status Status) (*Payment, error) {
	if status != StatusSucceeded && status != StatusFailed && status != StatusCancelled {
		return nil, fmt.Errorf("%w: unsupported status", ErrInvalidRequest)
	}

	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != StatusPending {
		return nil, ErrInvalidStatus
	}
	if p.PaymentMethod == "" {
		return nil, fmt.Errorf("%w: payment method required", ErrInvalidRequest)
	}

	p.Status = status
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// RedirectURL builds the merchant redirect URL for a completed action.
func RedirectURL(p *Payment) (string, error) {
	var raw string
	switch p.Status {
	case StatusSucceeded, StatusProcessing:
		raw = p.SuccessURL
	case StatusFailed:
		raw = p.FailedURL
	case StatusCancelled:
		raw = p.CancelURL
	default:
		return "", fmt.Errorf("payment is not redirectable")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("payment_id", p.ID)
	q.Set("status", string(p.Status))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func validateCreateRequest(req CreateRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	if req.SuccessURL == "" || req.CancelURL == "" || req.FailedURL == "" {
		return errors.New("success_url, cancel_url and failed_url are required")
	}
	for _, u := range []string{req.SuccessURL, req.CancelURL, req.FailedURL, req.WebhookURL, req.ITNURL} {
		if u == "" {
			continue
		}
		if _, err := url.ParseRequestURI(u); err != nil {
			return fmt.Errorf("invalid url: %s", u)
		}
	}
	if req.ITNDelay < 0 || req.ITNDelay > 300 {
		return errors.New("itn_delay_seconds must be between 0 and 300")
	}
	if req.ITNDelay > 0 && req.ITNURL == "" {
		return errors.New("itn_url is required when itn_delay_seconds is set")
	}
	if !IsValidMethod(req.PaymentMethod) {
		return errors.New("unknown payment method")
	}
	return nil
}

func newPaymentID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "pay_" + hex.EncodeToString(b)
}
