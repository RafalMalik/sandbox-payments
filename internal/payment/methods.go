package payment

// Method describes an available payment option shown to the payer.
type Method struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// AvailableMethods returns the sandbox payment methods supported by the gateway.
func AvailableMethods() []Method {
	return []Method{
		{Slug: "card", Name: "Card", Logo: "/logos/card.svg"},
		{Slug: "blik", Name: "BLIK", Logo: "/logos/blik.svg"},
		{Slug: "google_pay", Name: "Google Pay", Logo: "/logos/google-pay.svg"},
		{Slug: "bank_transfer", Name: "Bank Transfer", Logo: "/logos/bank.svg"},
	}
}

// IsValidMethod reports whether slug matches a known payment method.
func IsValidMethod(slug string) bool {
	if slug == "" {
		return true
	}
	for _, m := range AvailableMethods() {
		if m.Slug == slug {
			return true
		}
	}
	return false
}
