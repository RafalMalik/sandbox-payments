package payment

import "time"

// Status represents the lifecycle state of a payment.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

// Payment is the core domain entity persisted in storage.
type Payment struct {
	ID            string
	Amount        int64
	Currency      string
	Description   string
	Status        Status
	PaymentMethod string

	SuccessURL string
	CancelURL  string
	FailedURL  string
	WebhookURL string
	ITNURL     string
	ITNDelay   int

	Metadata string

	CreatedAt time.Time
	UpdatedAt time.Time
}
