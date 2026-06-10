package payment

import "context"

// Store abstracts persistence for payments.
type Store interface {
	Create(ctx context.Context, p *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	Update(ctx context.Context, p *Payment) error
}
