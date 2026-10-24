# Deploy

Host: org VPS `145.239.87.149` (`raspberry-vision`).
Public URL: `https://payments.raspberry-vision.cloud`
Upstream: Apache → `127.0.0.1:8117`.

GitHub repo `RafalMalik/sandbox-payments` (default branch `main`). Environment **`payments-vps`** secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY` (org VPS `145.239.87.149`, user `ubuntu`). Push to `main` or **Run workflow** runs CI then Deploy.

Deploy checks out the commit SHA on the VPS (`git fetch` + `git checkout`), runs Compose with the VPS overlay, syncs Apache, and probes `/health`. Do not rsync the tree to this host.

`GET /health` reports the version compiled into the binary (`internal/version`).

Apache pack: [`../deploy`](../deploy).

## First install

1. On the VPS:

```bash
sudo mkdir -p /opt/raspberry-vision
sudo chown ubuntu:ubuntu /opt/raspberry-vision
git clone https://github.com/RafalMalik/sandbox-payments.git /opt/raspberry-vision/sandbox-payments
```

2. Create `/opt/raspberry-vision/sandbox-payments/.env` (kept across deploys; not in git):

```
APP_ENV=prod
PORT=8117
BASE_URL=https://payments.raspberry-vision.cloud
```

3. Set GitHub environment secrets on **`payments-vps`**, then push `main` (or run **CI and deploy** via `workflow_dispatch`).

## Overlay

[`docker-compose.vps.yml`](../docker-compose.vps.yml) binds the API to `127.0.0.1:${PORT}` (production `8117`).

## Local

```bash
cp .env.example .env
docker compose up --build
go test ./...
```
