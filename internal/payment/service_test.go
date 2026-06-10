package payment

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
)

type memStore struct {
	mu       sync.Mutex
	payments map[string]*Payment
}

func newMemStore() *memStore {
	return &memStore{payments: make(map[string]*Payment)}
}

func (m *memStore) Create(_ context.Context, p *Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.payments[p.ID] = &cp
	return nil
}

func (m *memStore) GetByID(_ context.Context, id string) (*Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.payments[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (m *memStore) Update(_ context.Context, p *Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.payments[p.ID]; !ok {
		return sql.ErrNoRows
	}
	cp := *p
	m.payments[p.ID] = &cp
	return nil
}

func validCreateRequest() CreateRequest {
	return CreateRequest{
		Amount:        1999,
		Currency:      "pln",
		Description:   "Test order",
		PaymentMethod: "blik",
		SuccessURL:    "https://merchant.example/success",
		CancelURL:     "https://merchant.example/cancel",
		FailedURL:     "https://merchant.example/failed",
		Metadata:      map[string]string{"order_id": "42"},
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemStore(), "http://localhost:8080")
	ctx := context.Background()

	resp, err := svc.Create(ctx, validCreateRequest())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if resp.PaymentID == "" {
		t.Fatal("expected non-empty payment_id")
	}
	if resp.Status != StatusPending {
		t.Fatalf("status = %q, want %q", resp.Status, StatusPending)
	}
	if resp.RedirectURL != "http://localhost:8080/pay/"+resp.PaymentID {
		t.Fatalf("redirect_url = %q", resp.RedirectURL)
	}

	p, err := svc.Get(ctx, resp.PaymentID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if p.Currency != "PLN" {
		t.Fatalf("currency = %q, want PLN", p.Currency)
	}
	if p.Metadata != `{"order_id":"42"}` {
		t.Fatalf("metadata = %q", p.Metadata)
	}
}

func TestService_Create_validation(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemStore(), "http://localhost:8080")
	ctx := context.Background()

	tests := []struct {
		name    string
		mutate  func(*CreateRequest)
		wantErr error
	}{
		{
			name: "zero amount",
			mutate: func(r *CreateRequest) {
				r.Amount = 0
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "missing currency",
			mutate: func(r *CreateRequest) {
				r.Currency = ""
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "missing redirect urls",
			mutate: func(r *CreateRequest) {
				r.SuccessURL = ""
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "invalid success url",
			mutate: func(r *CreateRequest) {
				r.SuccessURL = "not-a-url"
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "unknown payment method",
			mutate: func(r *CreateRequest) {
				r.PaymentMethod = "crypto"
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "itn delay without url",
			mutate: func(r *CreateRequest) {
				r.ITNDelay = 5
			},
			wantErr: ErrInvalidRequest,
		},
		{
			name: "itn delay out of range",
			mutate: func(r *CreateRequest) {
				r.ITNURL = "https://merchant.example/itn"
				r.ITNDelay = 301
			},
			wantErr: ErrInvalidRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := validCreateRequest()
			tt.mutate(&req)

			_, err := svc.Create(ctx, req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Get_notFound(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemStore(), "http://localhost:8080")
	_, err := svc.Get(context.Background(), "pay_missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, ErrNotFound)
	}
}

func TestService_SelectMethod(t *testing.T) {
	t.Parallel()

	store := newMemStore()
	svc := NewService(store, "http://localhost:8080")
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Amount:      500,
		Currency:    "EUR",
		SuccessURL:  "https://merchant.example/success",
		CancelURL:   "https://merchant.example/cancel",
		FailedURL:   "https://merchant.example/failed",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	p, err := svc.SelectMethod(ctx, created.PaymentID, "card")
	if err != nil {
		t.Fatalf("SelectMethod() error = %v", err)
	}
	if p.PaymentMethod != "card" {
		t.Fatalf("payment_method = %q, want card", p.PaymentMethod)
	}

	_, err = svc.SelectMethod(ctx, created.PaymentID, "unknown")
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("SelectMethod() error = %v, want %v", err, ErrInvalidRequest)
	}
}

func TestService_Transition(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemStore(), "http://localhost:8080")
	ctx := context.Background()

	req := validCreateRequest()
	req.PaymentMethod = ""

	created, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = svc.Transition(ctx, created.PaymentID, StatusSucceeded)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Transition() without method error = %v, want %v", err, ErrInvalidRequest)
	}

	if _, err := svc.SelectMethod(ctx, created.PaymentID, "blik"); err != nil {
		t.Fatalf("SelectMethod() error = %v", err)
	}

	p, err := svc.Transition(ctx, created.PaymentID, StatusSucceeded)
	if err != nil {
		t.Fatalf("Transition() error = %v", err)
	}
	if p.Status != StatusSucceeded {
		t.Fatalf("status = %q, want succeeded", p.Status)
	}

	_, err = svc.Transition(ctx, created.PaymentID, StatusFailed)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("second Transition() error = %v, want %v", err, ErrInvalidStatus)
	}
}

func TestService_ITNFlow(t *testing.T) {
	t.Parallel()

	svc := NewService(newMemStore(), "http://localhost:8080")
	ctx := context.Background()

	req := validCreateRequest()
	req.ITNURL = "https://merchant.example/itn"
	req.ITNDelay = 0

	created, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	processing, err := svc.MarkProcessing(ctx, created.PaymentID)
	if err != nil {
		t.Fatalf("MarkProcessing() error = %v", err)
	}
	if processing.Status != StatusProcessing {
		t.Fatalf("status = %q, want processing", processing.Status)
	}

	redirect, err := RedirectURL(processing)
	if err != nil {
		t.Fatalf("RedirectURL() error = %v", err)
	}
	if redirect == "" {
		t.Fatal("expected redirect url for processing status")
	}

	confirmed, err := svc.ConfirmAfterITN(ctx, created.PaymentID)
	if err != nil {
		t.Fatalf("ConfirmAfterITN() error = %v", err)
	}
	if confirmed.Status != StatusSucceeded {
		t.Fatalf("status = %q, want succeeded", confirmed.Status)
	}
}
