package webhook

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rmalik/sandbox-payments/internal/payment"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNotifyPayment_deliversEvent(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		if payload.Event != "payment.succeeded" {
			t.Errorf("event = %q, want payment.succeeded", payload.Event)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	sender := NewSender(discardLogger())
	sender.NotifyPayment(&payment.Payment{
		ID:         "pay_webhook1",
		Amount:     1000,
		Currency:   "PLN",
		Status:     payment.StatusSucceeded,
		WebhookURL: server.URL,
	})

	waitFor(t, func() bool { return calls.Load() == 1 })
}

func TestNotifyPayment_skipsWithoutURL(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	sender := NewSender(discardLogger())
	sender.NotifyPayment(&payment.Payment{
		ID:       "pay_webhook2",
		Status:   payment.StatusSucceeded,
		WebhookURL: "",
	})

	time.Sleep(100 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatalf("calls = %d, want 0", calls.Load())
	}
}

func TestEventForStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status payment.Status
		want   string
	}{
		{status: payment.StatusSucceeded, want: "payment.succeeded"},
		{status: payment.StatusFailed, want: "payment.failed"},
		{status: payment.StatusCancelled, want: "payment.cancelled"},
		{status: payment.StatusPending, want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			if got := eventForStatus(tt.status); got != tt.want {
				t.Fatalf("eventForStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
