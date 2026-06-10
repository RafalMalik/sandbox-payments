# Sandbox Payment System (Go) - MVP Spec

## Overview

Minimalistyczny sandbox system płatności imitujący Stripe/PayU.

---

## Features

- GET payment methods
- Create payment with redirect flow
- Hosted payment page
- Payment status handling
- Redirect URLs (success/cancel/failed)
- Optional webhook support

---

## Payment Methods

### GET /api/payment-methods

Response:

```json
[
  { "slug": "card", "name": "Card", "logo": "/logos/card.svg" },
  { "slug": "blik", "name": "BLIK", "logo": "/logos/blik.svg" },
  { "slug": "google_pay", "name": "Google Pay", "logo": "/logos/google-pay.svg" },
  { "slug": "bank_transfer", "name": "Bank Transfer", "logo": "/logos/bank.svg" }
]
```

---

## Create Payment

### POST /api/payments

Request:

```json
{
  "amount": 1999,
  "currency": "PLN",
  "description": "Test order",
  "payment_method": "blik",
  "success_url": "https://app.com/success",
  "cancel_url": "https://app.com/cancel",
  "failed_url": "https://app.com/failed",
  "metadata": {
    "order_id": "123"
  }
}
```

Response:

```json
{
  "payment_id": "pay_123",
  "status": "pending",
  "redirect_url": "https://sandbox.local/pay/pay_123"
}
```

---

## Hosted Payment Page

### /pay/:id

Flow:

### 1. Method preselected

- shows method
- approve / fail / cancel buttons

### 2. No method selected

- show list of methods
- continue

---

## Payment Statuses

- pending
- succeeded
- failed
- cancelled

---

## Redirect Flow

After action:

```
https://app.com/success?payment_id=pay_123&status=succeeded
```

---

## Webhooks (optional)

POST:

```
payment.succeeded
payment.failed
payment.cancelled
```

Payload:

```json
{
  "event": "payment.succeeded",
  "payment_id": "pay_123",
  "amount": 1999,
  "currency": "PLN"
}
```

---

## Database Model

### Payment

```go
type Payment struct {
    ID string

    Amount int64
    Currency string
    Description string

    Status string
    PaymentMethod string

    SuccessURL string
    CancelURL string
    FailedURL string

    WebhookURL string

    Metadata string

    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

## Project Structure (Go)

```
cmd/api
internal/api
internal/payment
internal/storage
web/templates
web/static
```

---

## Implementation Tasks

### Phase 1 - Core API

1. Setup Go project
2. SQLite storage
3. Payment methods endpoint
4. Create payment endpoint
5. Get payment endpoint

---

### Phase 2 - Hosted Page

6. Payment page UI
7. Method selection
8. Action buttons (approve/fail/cancel)
9. Redirect handling

---

### Phase 3 - Webhooks

10. Webhook sender
11. Retry mechanism (3x)
12. Webhook logs

---

### Phase 4 - Dev Experience

13. Swagger / OpenAPI
14. Seed data
15. Docker setup

---

## Optional Improvements

### Test card simulation

- 4111 1111 1111 1111 → success
- 4000 0000 0000 0002 → declined

---

### API Keys

- pk_test_xxx
- sk_test_xxx

Header:

```
Authorization: Bearer sk_test_xxx
```

---

## Goal

Minimal sandbox payment gateway similar to Stripe but simplified.
