package payment

import "testing"

func TestRedirectURL(t *testing.T) {
	t.Parallel()

	base := Payment{
		ID:         "pay_test123",
		SuccessURL: "https://merchant.example/success",
		CancelURL:  "https://merchant.example/cancel",
		FailedURL:  "https://merchant.example/failed",
	}

	tests := []struct {
		name    string
		status  Status
		wantURL string
		wantErr bool
	}{
		{
			name:    "succeeded",
			status:  StatusSucceeded,
			wantURL: "https://merchant.example/success?payment_id=pay_test123&status=succeeded",
		},
		{
			name:    "processing uses success url",
			status:  StatusProcessing,
			wantURL: "https://merchant.example/success?payment_id=pay_test123&status=processing",
		},
		{
			name:    "failed",
			status:  StatusFailed,
			wantURL: "https://merchant.example/failed?payment_id=pay_test123&status=failed",
		},
		{
			name:    "cancelled",
			status:  StatusCancelled,
			wantURL: "https://merchant.example/cancel?payment_id=pay_test123&status=cancelled",
		},
		{
			name:    "pending not redirectable",
			status:  StatusPending,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := base
			p.Status = tt.status

			got, err := RedirectURL(&p)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("RedirectURL() error = %v", err)
			}
			if got != tt.wantURL {
				t.Fatalf("RedirectURL() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}
