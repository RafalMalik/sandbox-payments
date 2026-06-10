package api

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rmalik/sandbox-payments/internal/payment"
	"github.com/rmalik/sandbox-payments/internal/storage"
	"github.com/rmalik/sandbox-payments/internal/version"
	"github.com/rmalik/sandbox-payments/internal/webhook"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	root := projectRoot(t)
	store, err := storage.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	tmpl, err := template.ParseGlob(filepath.Join(root, "web/templates", "*.html"))
	if err != nil {
		t.Fatalf("ParseGlob() error = %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := payment.NewService(store, "http://localhost:8080")
	wh := webhook.NewSender(log)
	handler := NewHandler(
		svc, wh, tmpl,
		"http://localhost:8080",
		filepath.Join(root, "docs"),
		filepath.Join(root, "CHANGELOG.md"),
		log,
	)
	return NewServer(handler, filepath.Join(root, "web/static"), log)
}

func doJSON(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	t.Parallel()

	rec := doJSON(t, newTestHandler(t), http.MethodGet, "/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["status"] != "ok" || payload["version"] != version.Version {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestListPaymentMethods(t *testing.T) {
	t.Parallel()

	rec := doJSON(t, newTestHandler(t), http.MethodGet, "/api/payment-methods", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var methods []payment.Method
	if err := json.Unmarshal(rec.Body.Bytes(), &methods); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(methods) != 4 {
		t.Fatalf("len(methods) = %d, want 4", len(methods))
	}
}

func TestCreateAndGetPayment(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	body := map[string]any{
		"amount":         1999,
		"currency":       "PLN",
		"description":    "Test order",
		"payment_method": "blik",
		"success_url":    "https://merchant.example/success",
		"cancel_url":     "https://merchant.example/cancel",
		"failed_url":     "https://merchant.example/failed",
		"metadata":       map[string]string{"order_id": "99"},
	}

	createRec := doJSON(t, handler, http.MethodPost, "/api/payments", body)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	paymentID, _ := created["payment_id"].(string)
	if paymentID == "" {
		t.Fatalf("missing payment_id in %#v", created)
	}

	getRec := doJSON(t, handler, http.MethodGet, "/api/payments/"+paymentID, nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getRec.Code, getRec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal get: %v", err)
	}
	if got["payment_id"] != paymentID || got["status"] != "pending" {
		t.Fatalf("payment = %#v", got)
	}
}

func TestCreatePayment_validation(t *testing.T) {
	t.Parallel()

	rec := doJSON(t, newTestHandler(t), http.MethodPost, "/api/payments", map[string]any{
		"amount":      -1,
		"currency":    "PLN",
		"success_url": "https://merchant.example/success",
		"cancel_url":  "https://merchant.example/cancel",
		"failed_url":  "https://merchant.example/failed",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetPayment_notFound(t *testing.T) {
	t.Parallel()

	rec := doJSON(t, newTestHandler(t), http.MethodGet, "/api/payments/pay_missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestPaymentPage_andApproveRedirect(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	createRec := doJSON(t, handler, http.MethodPost, "/api/payments", map[string]any{
		"amount":         2500,
		"currency":       "PLN",
		"payment_method": "card",
		"success_url":    "https://merchant.example/success",
		"cancel_url":     "https://merchant.example/cancel",
		"failed_url":     "https://merchant.example/failed",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d", createRec.Code)
	}

	var created map[string]string
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	paymentID := created["payment_id"]

	pageReq := httptest.NewRequest(http.MethodGet, "/pay/"+paymentID, nil)
	pageRec := httptest.NewRecorder()
	handler.ServeHTTP(pageRec, pageReq)
	if pageRec.Code != http.StatusOK {
		t.Fatalf("page status = %d", pageRec.Code)
	}
	if !strings.Contains(pageRec.Body.String(), "PLN 25.00") {
		t.Fatal("expected formatted amount on payment page")
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/pay/"+paymentID+"/approve", nil)
	approveRec := httptest.NewRecorder()
	handler.ServeHTTP(approveRec, approveReq)
	if approveRec.Code != http.StatusSeeOther {
		t.Fatalf("approve status = %d, want 303", approveRec.Code)
	}

	location := approveRec.Header().Get("Location")
	if !strings.HasPrefix(location, "https://merchant.example/success") {
		t.Fatalf("location = %q", location)
	}
	if !strings.Contains(location, "status=succeeded") {
		t.Fatalf("location = %q, expected succeeded status", location)
	}
}

func TestSelectMethodRedirect(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	createRec := doJSON(t, handler, http.MethodPost, "/api/payments", map[string]any{
		"amount":       1000,
		"currency":     "PLN",
		"success_url":  "https://merchant.example/success",
		"cancel_url":   "https://merchant.example/cancel",
		"failed_url":   "https://merchant.example/failed",
	})
	var created map[string]string
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)

	form := "payment_method=blik"
	req := httptest.NewRequest(http.MethodPost, "/pay/"+created["payment_id"]+"/select", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if rec.Header().Get("Location") != "/pay/"+created["payment_id"] {
		t.Fatalf("location = %q", rec.Header().Get("Location"))
	}
}

func TestOpenAPISpec(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	newTestHandler(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi:") {
		t.Fatal("expected openapi spec content")
	}
}

func TestFormatAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount   int64
		currency string
		want     string
	}{
		{amount: 1999, currency: "pln", want: "PLN 19.99"},
		{amount: 2500, currency: "eur", want: "EUR 25.00"},
		{amount: 99, currency: "usd", want: "USD 0.99"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := formatAmount(tt.amount, tt.currency); got != tt.want {
				t.Fatalf("formatAmount() = %q, want %q", got, tt.want)
			}
		})
	}
}
