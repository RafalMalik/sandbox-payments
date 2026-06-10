package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmalik/sandbox-payments/internal/payment"
	"github.com/rmalik/sandbox-payments/internal/version"
	"github.com/rmalik/sandbox-payments/internal/webhook"
)

// Handler exposes HTTP endpoints for the sandbox payment gateway.
type Handler struct {
	payments *payment.Service
	webhooks *webhook.Sender
	tmpl     *template.Template
	baseURL       string
	docsDir       string
	changelogPath string
	log           *slog.Logger
}

// NewHandler wires dependencies for HTTP handling.
func NewHandler(payments *payment.Service, webhooks *webhook.Sender, tmpl *template.Template, baseURL, docsDir, changelogPath string, log *slog.Logger) *Handler {
	return &Handler{
		payments: payments, webhooks: webhooks, tmpl: tmpl,
		baseURL: baseURL, docsDir: docsDir, changelogPath: changelogPath, log: log,
	}
}

// RegisterRoutes attaches all application routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /{$}", h.homePage)
	mux.HandleFunc("GET /changelog", h.changelogPage)
	mux.HandleFunc("GET /docs/api", h.apiDocsPage)
	mux.HandleFunc("GET /docs/api/", h.apiDocsPage)
	mux.HandleFunc("GET /openapi.yaml", h.openAPISpec)
	mux.HandleFunc("GET /docs", h.apiDocsRedirect)
	mux.HandleFunc("GET /docs/", h.apiDocsRedirect)
	mux.HandleFunc("GET /docs/openapi.yaml", h.openAPIRedirect)
	mux.HandleFunc("GET /doc", h.apiDocsRedirect)
	mux.HandleFunc("GET /api/payment-methods", h.listMethods)
	mux.HandleFunc("POST /api/payments", h.createPayment)
	mux.HandleFunc("GET /api/payments/{id}", h.getPayment)
	mux.HandleFunc("GET /pay/{id}", h.paymentPage)
	mux.HandleFunc("POST /pay/{id}/select", h.selectMethod)
	mux.HandleFunc("POST /pay/{id}/approve", h.approvePayment)
	mux.HandleFunc("POST /pay/{id}/fail", h.failPayment)
	mux.HandleFunc("POST /pay/{id}/cancel", h.cancelPayment)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": version.Version,
	})
}

func (h *Handler) listMethods(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.payments.ListMethods())
}

func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req payment.CreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.payments.Create(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	p, err := h.payments.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paymentResponse(p))
}

type payPageData struct {
	Payment *payment.Payment
	Methods []payment.Method
	Amount  string
}

func (h *Handler) paymentPage(w http.ResponseWriter, r *http.Request) {
	p, err := h.payments.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := payPageData{
		Payment: p,
		Methods: h.payments.ListMethods(),
		Amount:  formatAmount(p.Amount, p.Currency),
	}

	if err := h.tmpl.ExecuteTemplate(w, "pay.html", data); err != nil {
		h.log.Error("render payment page", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) selectMethod(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	method := r.FormValue("payment_method")

	if _, err := h.payments.SelectMethod(r.Context(), id, method); err != nil {
		h.handlePageError(w, r, err)
		return
	}

	http.Redirect(w, r, "/pay/"+id, http.StatusSeeOther)
}

func (h *Handler) approvePayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.payments.Get(r.Context(), id)
	if err != nil {
		h.handlePageError(w, r, err)
		return
	}

	if p.ITNURL != "" {
		p, err = h.payments.MarkProcessing(r.Context(), id)
		if err != nil {
			h.handlePageError(w, r, err)
			return
		}

		h.webhooks.NotifyITN(p, func(ctx context.Context, _ *payment.Payment) error {
			confirmed, err := h.payments.ConfirmAfterITN(ctx, id)
			if err != nil {
				return err
			}
			h.webhooks.NotifyPayment(confirmed)
			return nil
		})

		target, err := payment.RedirectURL(p)
		if err != nil {
			h.log.Error("build redirect url", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	h.transition(w, r, payment.StatusSucceeded)
}

func (h *Handler) failPayment(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, payment.StatusFailed)
}

func (h *Handler) cancelPayment(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, payment.StatusCancelled)
}

func (h *Handler) transition(w http.ResponseWriter, r *http.Request, status payment.Status) {
	p, err := h.payments.Transition(r.Context(), r.PathValue("id"), status)
	if err != nil {
		h.handlePageError(w, r, err)
		return
	}

	h.webhooks.NotifyPayment(p)

	target, err := payment.RedirectURL(p)
	if err != nil {
		h.log.Error("build redirect url", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, target, http.StatusSeeOther)
}

type sitePageData struct {
	BaseURL         string
	Version         string
	OpenAPISpecData template.URL
}

func (h *Handler) homePage(w http.ResponseWriter, r *http.Request) {
	data := sitePageData{
		BaseURL: h.baseURL,
		Version: version.Version,
	}
	if err := h.tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		h.log.Error("render home page", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

type changelogPageData struct {
	Version string
	Content string
}

func (h *Handler) changelogPage(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(h.changelogPath)
	if err != nil {
		h.log.Error("read changelog", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := changelogPageData{
		Version: version.Version,
		Content: string(content),
	}
	if err := h.tmpl.ExecuteTemplate(w, "changelog.html", data); err != nil {
		h.log.Error("render changelog page", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) apiDocsPage(w http.ResponseWriter, r *http.Request) {
	specData, err := h.loadOpenAPISpecData()
	if err != nil {
		h.log.Error("read openapi spec", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := sitePageData{
		BaseURL:         h.baseURL,
		Version:         version.Version,
		OpenAPISpecData: specData,
	}
	if err := h.tmpl.ExecuteTemplate(w, "api-docs.html", data); err != nil {
		h.log.Error("render api docs page", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) apiDocsRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/docs/api", http.StatusMovedPermanently)
}

func (h *Handler) loadOpenAPISpecData() (template.URL, error) {
	spec, err := os.ReadFile(filepath.Join(h.docsDir, "openapi.yaml"))
	if err != nil {
		return "", err
	}
	return template.URL("data:application/yaml;base64," + base64.StdEncoding.EncodeToString(spec)), nil
}

func (h *Handler) openAPIRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/openapi.yaml", http.StatusMovedPermanently)
}

func (h *Handler) openAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, filepath.Join(h.docsDir, "openapi.yaml"))
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payment.ErrNotFound):
		writeError(w, http.StatusNotFound, "payment not found")
	case errors.Is(err, payment.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, payment.ErrInvalidStatus):
		writeError(w, http.StatusConflict, err.Error())
	default:
		h.log.Error("unexpected service error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *Handler) handlePageError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, payment.ErrNotFound):
		http.NotFound(w, r)
	case errors.Is(err, payment.ErrInvalidRequest), errors.Is(err, payment.ErrInvalidStatus):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		h.log.Error("unexpected page error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

type paymentAPIResponse struct {
	PaymentID     string          `json:"payment_id"`
	Amount        int64           `json:"amount"`
	Currency      string          `json:"currency"`
	Description   string          `json:"description"`
	Status        payment.Status  `json:"status"`
	PaymentMethod string          `json:"payment_method,omitempty"`
	ITNURL        string          `json:"itn_url,omitempty"`
	ITNDelay      int             `json:"itn_delay_seconds,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

func paymentResponse(p *payment.Payment) paymentAPIResponse {
	return paymentAPIResponse{
		PaymentID:     p.ID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Description:   p.Description,
		Status:        p.Status,
		PaymentMethod: p.PaymentMethod,
		ITNURL:        p.ITNURL,
		ITNDelay:      p.ITNDelay,
		Metadata:      json.RawMessage(p.Metadata),
		CreatedAt:     p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty request body")
	}
	return json.Unmarshal(body, dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func formatAmount(amount int64, currency string) string {
	whole := amount / 100
	frac := amount % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%s %d.%02d", strings.ToUpper(currency), whole, frac)
}
