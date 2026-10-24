# Deploy

Host: org VPS `145.239.87.149` (`raspberry-vision`).
Public URL: `https://payments.raspberry-vision.cloud`
Upstream: Apache → `127.0.0.1:8117`.
VPS path: `/opt/raspberry-vision/sandbox-payments`.

GitHub: default branch `main`, environment **`payments-vps`** (`VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`). CI calls the shared reusable deploy workflow (git checkout on the VPS — see `raspberry-vision/templates` `github/reusable-deploy-vps.yml`).

`GET /health` reports `internal/version`.

Apache pack: [`../deploy`](../deploy).

## First install

1. Clone once on the VPS to `/opt/raspberry-vision/sandbox-payments`.
2. Create `.env` there (not in git):

```
APP_ENV=prod
PORT=8117
BASE_URL=https://payments.raspberry-vision.cloud
```

3. Set `payments-vps` secrets, then push `main` or run **CI and deploy**.

## Overlay

[`docker-compose.vps.yml`](../docker-compose.vps.yml) binds the API to `127.0.0.1:${PORT}` (`8117` in prod).

## Local

```bash
cp .env.example .env
docker compose up --build
go test ./...
```
