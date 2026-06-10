# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-06-10

### Added

- Postman collection at `docs/sandbox-payments.postman_collection.json` with local base URL (`http://localhost:8080`) and example request payloads

## [1.0.0] - 2026-06-10

### Added

- Payment API: list methods, create payment, get payment by ID
- Hosted payment page with method selection and approve / fail / cancel actions
- Redirect flow with merchant callback URLs
- Optional webhooks with 3 retry attempts
- ITN (`itn_url`) with configurable delay (`itn_delay_seconds`)
- SQLite persistence (pure Go driver, no CGO)
- Docker and Docker Compose setup with `.env` configuration
- Home page at `/` and interactive API docs at `/docs/api` (Swagger UI)
- OpenAPI specification at `/openapi.yaml`

[1.1.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.1.0
[1.0.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.0.0
