# Sandbox Payments

Minimal payment gateway sandbox in Go — a simplified Stripe/PayU-like service for local integration testing and demos.

## Problem

Building and testing payment integrations against real providers (Stripe, PayU, Przelewy24) is slow, costly, and brittle:

- Sandbox accounts require setup and often have rate limits.
- Redirect flows, webhooks, and ITN callbacks are hard to exercise in unit tests.
- Teams need a predictable, self-hosted environment to develop checkout flows without touching production systems.

## Solution

Sandbox Payments is a lightweight, self-contained payment gateway that mimics the core merchant flow:

1. Create a payment via REST API.
2. Redirect the payer to a hosted payment page.
3. Simulate approve / fail / cancel outcomes.
4. Redirect back to merchant URLs with `payment_id` and `status`.
5. Optionally deliver webhooks and ITN notifications.

No external services, no CGO, no frameworks — just Go standard library and SQLite.

## Features

- REST API: list methods, create payment, get payment by ID
- Hosted payment page with method selection and action buttons
- Redirect flow with merchant callback URLs (`success`, `cancel`, `failed`)
- Optional webhooks with 3 retry attempts
- ITN (`itn_url`) with configurable delay (`itn_delay_seconds`)
- SQLite persistence (pure Go driver, no CGO)
- Docker and Docker Compose setup
- Home page, interactive API docs (Swagger UI), OpenAPI spec

## Tech stack

| Layer        | Choice                                      |
|--------------|---------------------------------------------|
| Language     | Go 1.22                                     |
| HTTP         | `net/http` (stdlib, Go 1.22+ routing)       |
| Templates    | `html/template`                             |
| Logging      | `log/slog`                                  |
| Database     | SQLite via `modernc.org/sqlite` (no CGO)    |
| Deployment   | Docker multi-stage build                    |

No web frameworks, ORM, or extra HTTP libraries.

## Architecture

```text
┌─────────────┐     REST / HTML      ┌──────────────────┐
│   Merchant  │ ◄──────────────────► │   HTTP (api)     │
│     app     │                      │  handlers + mux  │
└─────────────┘                      └────────┬─────────┘
                                            │
                         ┌──────────────────┼──────────────────┐
                         ▼                  ▼                  ▼
                  ┌────────────┐    ┌────────────┐    ┌────────────┐
                  │  payment   │    │  webhook   │    │   SQLite   │
                  │  service   │    │   sender   │    │   storage  │
                  └────────────┘    └────────────┘    └────────────┘
```

Domain logic lives in `internal/payment` behind a `Store` interface — storage and HTTP layers depend on abstractions, not concrete implementations.

## Quick start

### Docker (recommended)

```bash
cp .env.example .env
docker compose up --build
```

Port and public URL come from `.env` (default `8080`):

```env
PORT=8080
BASE_URL=http://localhost:8080
```

- Home: [http://localhost:8080/](http://localhost:8080/)
- API docs: [http://localhost:8080/docs/api](http://localhost:8080/docs/api)

### Local

```bash
go mod download
go run ./cmd/api
```

| Variable         | Default                 | Description                 |
|------------------|-------------------------|-----------------------------|
| `PORT`           | `8080`                  | HTTP port                   |
| `BASE_URL`       | `http://localhost:8080` | Base URL for `redirect_url` |
| `DATABASE_PATH`  | `./data/payments.db`    | SQLite file path            |
| `STATIC_DIR`     | `web/static`            | Static assets               |
| `TEMPLATE_DIR`   | `web/templates`         | HTML templates              |
| `DOCS_DIR`       | `docs`                  | OpenAPI spec directory      |
| `CHANGELOG_PATH` | `CHANGELOG.md`          | Changelog file for `/changelog` page |

## Testing

```bash
go test ./... -race -count=1
```

Test coverage includes:

- Payment domain logic (validation, transitions, redirect URLs)
- SQLite persistence (in-memory)
- HTTP API endpoints (integration via `httptest`)
- Webhook delivery (async with retry)

## Documentation

- **Home** (`/`) — overview, payment flow, statuses, ITN, webhooks
- **API docs** (`/docs/api`) — interactive Swagger UI
- **OpenAPI** (`/openapi.yaml`) — machine-readable spec

Release history is maintained in [CHANGELOG.md](CHANGELOG.md) (separate from this README).

## API overview

### List payment methods

```bash
curl http://localhost:8080/api/payment-methods
```

### Create payment

```bash
curl -X POST http://localhost:8080/api/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1999,
    "currency": "PLN",
    "description": "Test order",
    "payment_method": "blik",
    "success_url": "https://app.com/success",
    "cancel_url": "https://app.com/cancel",
    "failed_url": "https://app.com/failed",
    "metadata": {"order_id": "123"}
  }'
```

Response:

```json
{
  "payment_id": "pay_abc123",
  "status": "pending",
  "redirect_url": "http://localhost:8080/pay/pay_abc123"
}
```

### Get payment

```bash
curl http://localhost:8080/api/payments/pay_abc123
```

## Hosted payment page

1. Open `redirect_url` in a browser.
2. Select a method if needed, then click **Continue**.
3. Click **Approve**, **Simulate failure**, or **Cancel**.
4. Redirect to merchant URL:

```
https://app.com/success?payment_id=pay_abc123&status=succeeded
```

## ITN (optional)

Set `itn_url` to simulate releasing the payer before final confirmation:

1. Approve → redirect to `success_url?status=processing`
2. After `itn_delay_seconds` (0–300) → POST to `itn_url`
3. After successful ITN → status `succeeded` + webhook (if configured)

## Webhooks (optional)

Set `webhook_url` on payment creation. A POST is sent on final status (`succeeded`, `failed`, `cancelled`) with 3 retry attempts.

## Payment methods

| Slug            | Name          |
|-----------------|---------------|
| `card`          | Card          |
| `blik`          | BLIK          |
| `google_pay`    | Google Pay    |
| `bank_transfer` | Bank Transfer |

## Health check

```bash
curl http://localhost:8080/health
```

Returns `status` and application `version` (defined in `internal/version/version.go`).

## Project structure

```text
cmd/api/              # entrypoint
internal/
  api/                # HTTP handlers, middleware
  config/             # env configuration
  payment/            # domain logic + Store interface
  storage/            # SQLite implementation
  version/            # release version
  webhook/            # webhook & ITN delivery
web/templates/        # HTML pages
web/static/           # CSS, logos
docs/openapi.yaml     # OpenAPI spec
CHANGELOG.md          # release history (Keep a Changelog)
```

## Author

**Roman Malik** — [github.com/rmalik](https://github.com/rmalik)

## License

MIT
