# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2026-10-24

### Added

- GitHub Actions CI and deploy to the apps host (`payments-vps`): push to `main` or `workflow_dispatch` calls the shared `reusable-deploy-vps` template (git checkout on the VPS, Compose, Apache sync, `/health`).

## [1.5.0] - 2026-10-20

### Added

- OpenAPI lists `GET /health` and the public server `https://payments.raspberry-vision.cloud`.
- Hosted pages send Open Graph tags and Plausible. `GET /favicon.svg`, `GET /robots.txt`, and `GET /sitemap.xml` are public.

### Fixed

- The image healthcheck and the VPS Compose check call `/health` on `PORT`, so the apps host on `8117` is not probed on `8080`.

## [1.4.0] - 2026-10-20

### Added

- Hosted pages use the Sandbox Payments brand: `#BD1C00` tokens, source lockups, a 40px horizontal crop in the header, `favicon-32.png`, and the Open Graph image.

## [1.3.0] - 2026-10-16

### Added

- Apache pack for `payments.raspberry-vision.cloud` on the shared apps host. The service is no longer part of the iBiyahe VPS.

## [1.2.0] - 2026-10-11

### Changed

- Hosted pages use Raspberry Vision tokens, the raspberry mark, and the studio credit footer.

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

[1.6.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.6.0
[1.5.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.5.0
[1.4.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.4.0
[1.3.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.3.0
[1.2.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.2.0
[1.1.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.1.0
[1.0.0]: https://github.com/rmalik/sandbox-payments/releases/tag/v1.0.0
