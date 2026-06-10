package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/rmalik/sandbox-payments/internal/payment"
)

func openTestStore(t *testing.T) *SQLiteStore {
	t.Helper()

	store, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func samplePayment(id string) *payment.Payment {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	return &payment.Payment{
		ID:            id,
		Amount:        1999,
		Currency:      "PLN",
		Description:   "Order",
		Status:        payment.StatusPending,
		PaymentMethod: "blik",
		SuccessURL:    "https://merchant.example/success",
		CancelURL:     "https://merchant.example/cancel",
		FailedURL:     "https://merchant.example/failed",
		WebhookURL:    "https://merchant.example/webhook",
		ITNURL:        "https://merchant.example/itn",
		ITNDelay:      10,
		Metadata:      `{"order_id":"1"}`,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestSQLiteStore_CreateGetUpdate(t *testing.T) {
	t.Parallel()

	store := openTestStore(t)
	ctx := context.Background()
	p := samplePayment("pay_storage1")

	if err := store.Create(ctx, p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Amount != p.Amount || got.ITNDelay != p.ITNDelay || got.Metadata != p.Metadata {
		t.Fatalf("GetByID() = %+v, want %+v", got, p)
	}

	got.Status = payment.StatusSucceeded
	got.UpdatedAt = got.UpdatedAt.Add(time.Minute)
	if err := store.Update(ctx, got); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := store.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() after update error = %v", err)
	}
	if updated.Status != payment.StatusSucceeded {
		t.Fatalf("status = %q, want succeeded", updated.Status)
	}
}

func TestSQLiteStore_GetByID_notFound(t *testing.T) {
	t.Parallel()

	store := openTestStore(t)
	_, err := store.GetByID(context.Background(), "pay_missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetByID() error = %v, want sql.ErrNoRows", err)
	}
}
