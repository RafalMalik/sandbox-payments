package payment

import "testing"

func TestIsValidMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		slug string
		want bool
	}{
		{slug: "", want: true},
		{slug: "blik", want: true},
		{slug: "card", want: true},
		{slug: "google_pay", want: true},
		{slug: "bank_transfer", want: true},
		{slug: "paypal", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.slug, func(t *testing.T) {
			t.Parallel()
			if got := IsValidMethod(tt.slug); got != tt.want {
				t.Fatalf("IsValidMethod(%q) = %v, want %v", tt.slug, got, tt.want)
			}
		})
	}
}

func TestAvailableMethods(t *testing.T) {
	t.Parallel()

	methods := AvailableMethods()
	if len(methods) != 4 {
		t.Fatalf("len(methods) = %d, want 4", len(methods))
	}

	seen := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		if m.Slug == "" || m.Name == "" || m.Logo == "" {
			t.Fatalf("incomplete method: %+v", m)
		}
		seen[m.Slug] = struct{}{}
	}

	for _, slug := range []string{"card", "blik", "google_pay", "bank_transfer"} {
		if _, ok := seen[slug]; !ok {
			t.Fatalf("missing method slug %q", slug)
		}
	}
}
