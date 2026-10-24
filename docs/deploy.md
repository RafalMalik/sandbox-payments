# Deploy

Host: org VPS `145.239.87.149` (`raspberry-vision`).
Public URL: `https://payments.raspberry-vision.cloud`
Upstream: Apache → `127.0.0.1:8117`.

GitHub repo `RafalMalik/sandbox-payments` (default branch `main`). Environment **`payments-vps`** secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY` (org VPS `145.239.87.149`, user `ubuntu`). Push to `main` or **Run workflow** runs CI then Deploy.

`GET /health` reports the version compiled into the binary (`internal/version`).

Apache pack: [`../deploy`](../deploy).

## First install

1. On the VPS:

```bash
sudo mkdir -p /opt/raspberry-vision/sandbox-payments
sudo chown -R ubuntu: /opt/raspberry-vision/sandbox-payments
```

2. Create `/opt/raspberry-vision/sandbox-payments/.env` (kept across deploys; not overwritten by the workflow):

```
APP_ENV=prod
PORT=8117
BASE_URL=https://payments.raspberry-vision.cloud
```

3. Set GitHub environment secrets on **`payments-vps`**, then push `main` (or run **CI and deploy** via `workflow_dispatch`). The job rsyncs to `/opt/raspberry-vision/sandbox-payments` and runs:

```bash
docker compose -f docker-compose.yml -f docker-compose.vps.yml up -d --build
```

then `deploy/sync-apache.sh`, and checks `http://127.0.0.1:8117/health` plus the public URL.

## Overlay

[`docker-compose.vps.yml`](../docker-compose.vps.yml) binds the API to `127.0.0.1:${PORT}` (production `8117`).

## Local

```bash
cp .env.example .env
docker compose up --build
go test ./...
```
